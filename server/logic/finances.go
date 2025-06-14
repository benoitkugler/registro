package logic

import (
	"fmt"

	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	fs "registro/sql/files"
	"registro/utils"
)

// DossiersFinances adds aides and paiements
type DossiersFinances struct {
	Dossiers

	taux       ds.Tauxs
	aides      map[cps.IdParticipant]cps.Aides
	aidesFiles map[cps.IdAide]fs.File // enough for [aides]
	structures cps.Structureaides
	paiements  map[ds.IdDossier]ds.Paiements
}

// LoadDossiersFinance is a convenient wrapper around [LoadDossiersFinances]
func LoadDossiersFinance(db ds.DB, id ds.IdDossier) (DossierFinance, error) {
	ld, err := LoadDossiersFinances(db, id)
	if err != nil {
		return DossierFinance{}, err
	}
	return ld.For(id), nil
}

// LoadDossiersFinances wraps the SQL error
func LoadDossiersFinances(db ds.DB, ids ...ds.IdDossier) (DossiersFinances, error) {
	dossiers, err := LoadDossiers(db, ids...)
	if err != nil {
		return DossiersFinances{}, err
	}

	tauxs, err := ds.SelectTauxs(db, dossiers.Dossiers.IdTauxs()...)
	if err != nil {
		return DossiersFinances{}, utils.SQLError(err)
	}

	aides, err := cps.SelectAidesByIdParticipants(db, dossiers.allParticipants...)
	if err != nil {
		return DossiersFinances{}, utils.SQLError(err)
	}
	// justificatifs
	links, err := fs.SelectFileAidesByIdAides(db, aides.IDs()...)
	if err != nil {
		return DossiersFinances{}, utils.SQLError(err)
	}
	files, err := fs.SelectFiles(db, links.IdFiles()...)
	if err != nil {
		return DossiersFinances{}, utils.SQLError(err)
	}
	aidesFiles := make(map[cps.IdAide]fs.File)
	for _, link := range links {
		aidesFiles[link.IdAide] = files[link.IdFile]
	}

	structures, err := cps.SelectStructureaides(db, aides.IdStructureaides()...)
	if err != nil {
		return DossiersFinances{}, utils.SQLError(err)
	}
	paiements, err := ds.SelectPaiementsByIdDossiers(db, ids...)
	if err != nil {
		return DossiersFinances{}, utils.SQLError(err)
	}

	return DossiersFinances{dossiers, tauxs, aides.ByIdParticipant(), aidesFiles, structures, paiements.ByIdDossier()}, nil
}

func (df DossiersFinances) For(id ds.IdDossier) DossierFinance {
	out := df.Dossiers.For(id)
	aides := make(map[cps.IdParticipant]cps.Aides)
	for _, part := range out.Participants {
		aides[part.Id] = df.aides[part.Id]
	}
	return DossierFinance{out, df.taux[out.Dossier.IdTaux], aides, df.aidesFiles, df.structures, df.paiements[id]}
}

type DossierFinance struct {
	Dossier

	taux ds.Taux

	aides      map[cps.IdParticipant]cps.Aides // including not validated
	aidesFiles map[cps.IdAide]fs.File          // enough for [aides]
	structures cps.Structureaides              // enough for [aides]

	paiements ds.Paiements // liste exacte
}

// Bilan additionne le prix pour chaque participant et
// décompte les paiements effectués.
//
// Seuls les participants inscrits (en liste principale) sont pris en compte,
// les participants en liste d'attente étant ignorés.
//
// Les aides en cours de validation sont ignorées.
func (df DossierFinance) Bilan() BilanFinances {
	inscrits := map[cps.IdParticipant]BilanParticipant{}
	demande, recu, aides := df.taux.Zero(), df.taux.Zero(), df.taux.Zero()

	for _, participant := range df.Participants {
		if participant.Statut != cps.Inscrit {
			continue
		}
		data := pc{participant, df.camps[participant.IdCamp], df.aides[participant.Id], df.structures}
		bilan := data.bilan()
		inscrits[participant.Id] = bilan
		demande.Add(bilan.net(df.taux))
		aides.Add(bilan.totalAides(df.taux))
	}

	for _, paiement := range df.paiements {
		if paiement.IsRemboursement {
			recu.Sub(paiement.Montant)
		} else {
			recu.Add(paiement.Montant)
		}
	}

	// [demande], [recu] and [aides] have the same currency
	return BilanFinances{inscrits, demande.Cent, recu.Cent, aides.Cent, demande.Currency}
}

// BilanFinances résume l'état financier d'un dossier
type BilanFinances struct {
	inscrits map[cps.IdParticipant]BilanParticipant

	// totaux, en centimes

	demande int // montant demandé final (avant déduction des paiements déjà effectués)
	recu    int // somme des paiements reçus

	aides int // total des aides

	currency ds.Currency
}

func (b BilanFinances) IsAcquitte() bool { return b.demande <= b.recu }

// ApresPaiement renvoie le montant restant à payer
func (b BilanFinances) ApresPaiement() ds.Montant {
	return ds.Montant{Cent: b.demande - b.recu, Currency: b.currency}
}

type StatutPaiement uint8

const (
	_           StatutPaiement = iota //
	NonCommence                       // Non commencé
	EnCours                           // En cours
	Complet                           // Complet
)

func (b BilanFinances) StatutPaiement() StatutPaiement {
	if b.IsAcquitte() {
		return Complet
	} else if b.recu > 0 {
		return EnCours
	} else {
		return NonCommence
	}
}

type pc struct {
	cps.Participant
	cps.Camp

	aides      cps.Aides
	structures cps.Structureaides // enough for [aides]
}

// BilanParticipant exposes the finances for one [Participant]
//
// Le prix final est calculé en suivant 3 étapes :
//   - applique les options et le quotient familial
//   - applique les aides extérieureurs
//   - applique les remises internes
type BilanParticipant struct {
	AvecOption            cps.Montant // prend en compte une éventuelle option et le quotient familial
	AvecOptionDescription string      // courte description affichée dans la facture

	Remises cps.Remises

	// Aides validées seulement
	Aides []AideResolved
}

func (bp BilanParticipant) totalAides(taux ds.Taux) cps.Montant {
	totalAide := taux.Zero()
	for _, aide := range bp.Aides {
		totalAide.Add(aide.Montant)
	}
	return totalAide.Montant
}

// prixSansRemises renvoie le prix du séjour si on applique les aides (extérieures)
// mais pas les remises.
func (bp BilanParticipant) prixSansRemises(taux ds.Taux) cps.Montant {
	out := taux.Convertible(bp.AvecOption)

	// totalAides ignore les aides non validées
	totalAide := bp.totalAides(taux)

	out.Sub(totalAide)

	p := out.Montant
	if p.Cent < 0 {
		p.Cent = 0
	}
	return p
}

// Net renvoie le prix à payer après avoir déduit les aides extérieures et les remises
func (bp BilanParticipant) net(taux ds.Taux) cps.Montant {
	v1 := bp.prixSansRemises(taux).Remise(bp.Remises.ReducEnfants + bp.Remises.ReducEquipiers)
	val := taux.Convertible(v1)
	val.Sub(bp.Remises.ReducSpeciale)
	if val.Montant.Cent < 0 {
		val.Montant.Cent = 0
	}
	return val.Montant
}

// AideResolved shows resolved [AideResolved]
type AideResolved struct {
	Structure string
	Montant   cps.Montant
}

func (p pc) bilan() (out BilanParticipant) {
	out.AvecOption, out.AvecOptionDescription = p.prixBase()
	out.Remises = p.Participant.Remises

	duree := p.duree()

	for _, id := range utils.MapKeysSorted(p.aides) {
		aide := p.aides[id]
		if !aide.Valide {
			continue
		}
		out.Aides = append(out.Aides, AideResolved{p.structures[aide.IdStructureaide].Nom, aide.Resolve(duree)})
	}
	return out
}

// duree renvoie le nombre de jours de présence du participant
// en prenant en compte une éventuelle option JOUR.
func (p pc) duree() int {
	options := p.Camp.OptionPrix
	optPart := p.Participant.OptionPrix

	switch options.Active {
	case cps.PrixJour:
		return optPart.Jour.NbJours(p.Camp.Duree)
	}
	return p.Camp.Duree
}

// prixBase renvoie le prix du camp, en prenant en compte une éventuelle option et le quotient familial
// une courte description est aussi renvoyée
func (p pc) prixBase() (cps.Montant, string) {
	optPart := p.Participant.OptionPrix
	optCamp := p.Camp.OptionPrix

	prix := p.Camp.Prix
	descOption, descQF := "", ""
	switch optCamp.Active {
	case cps.PrixStatut:
		for _, info := range optCamp.Statuts {
			if info.Id == optPart.IdStatut {
				prix.Cent = info.Prix
				descOption = info.Label
			}
		}
	case cps.PrixJour:
		nbJours := optPart.Jour.NbJours(p.Camp.Duree)
		prixJours := optCamp.Jours
		if nbJours == 0 || nbJours >= len(prixJours) {
			// invalide -> camp entier
			// potentiellement < somme(prixJours)
		} else {
			prix = optPart.Jour.CalculePrix(prixJours, p.Camp.Prix.Currency)
			plural := ""
			if nbJours > 1 {
				plural = "s"
			}
			descOption = fmt.Sprintf("%d jour%s", nbJours, plural)
		}
	}

	// réduction quotient familial
	qf := p.Participant.QuotientFamilial
	if qfCamp := p.Camp.OptionQuotientFamilial; qfCamp.IsActive() && qf > 0 {
		ratio := qfCamp.Percentage(qf)
		prix.Cent = prix.Cent * ratio / 100
		descQF = fmt.Sprintf("QF %d", qf)
	}

	if descOption != "" && descQF != "" {
		descOption += " - "
	}
	desc := descOption + descQF

	return prix, desc
}
