package logic

import (
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/utils"
)

// DossiersFinances adds aides and paiements
type DossiersFinances struct {
	Dossiers

	taux       ds.Tauxs
	aides      map[cps.IdParticipant]cps.Aides
	structures cps.Structureaides
	paiements  map[ds.IdDossier]ds.Paiements
}

// LoadDossiers wraps the SQL error
func NewDossiersFinances(db ds.DB, ids ...ds.IdDossier) (DossiersFinances, error) {
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
	structures, err := cps.SelectStructureaides(db, aides.IdStructureaides()...)
	if err != nil {
		return DossiersFinances{}, utils.SQLError(err)
	}
	paiements, err := ds.SelectPaiementsByIdDossiers(db, ids...)
	if err != nil {
		return DossiersFinances{}, utils.SQLError(err)
	}

	return DossiersFinances{dossiers, tauxs, aides.ByIdParticipant(), structures, paiements.ByIdDossier()}, nil
}

func (df DossiersFinances) For(id ds.IdDossier) DossierFinance {
	out := df.Dossiers.For(id)
	aides := make(map[cps.IdParticipant]cps.Aides)
	for _, part := range out.Participants {
		aides[part.Id] = df.aides[part.Id]
	}
	return DossierFinance{out, df.taux[out.Dossier.IdTaux], aides, df.structures, df.paiements[id]}
}

type DossierFinance struct {
	Dossier

	taux ds.Taux

	aides      map[cps.IdParticipant]cps.Aides
	structures cps.Structureaides // enough for [aides]

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
	demande, recu := df.taux.Zero(), df.taux.Zero()

	for _, participant := range df.Participants {
		if participant.Statut != cps.Inscrit {
			continue
		}
		data := pc{participant, df.camps[participant.IdCamp], df.aides[participant.Id], df.structures}
		net := data.finances().Net(df.taux)
		demande.Add(net)
	}

	for _, paiement := range df.paiements {
		if paiement.IsRemboursement {
			recu.Sub(paiement.Montant)
		} else {
			recu.Add(paiement.Montant)
		}
	}

	// [demande] and [recu] have the same currency
	return BilanFinances{demande.Currency, demande.Cent, recu.Cent}
}

type pc struct {
	cps.Participant
	cps.Camp

	aides      cps.Aides
	structures cps.Structureaides // enough for [aides]
}

// BilanParticipant exposes the finances for one [Participant]
type BilanParticipant struct {
	Base    cps.Montant
	Remises cps.Remises

	// Aides validées
	Aides []Aide
}

// prixSansRemises renvoie le prix du séjour si on applique les aides (extérieures)
// mais pas les remises.
func (bp BilanParticipant) prixSansRemises(taux ds.Taux) cps.Montant {
	out := taux.Convertible(bp.Base)
	out.Sub(bp.totalAides(taux))
	p := out.Montant
	if p.Cent < 0 {
		p.Cent = 0
	}
	return p
}

// totalAides ignore les aides non validées
func (bp BilanParticipant) totalAides(taux ds.Taux) cps.Montant {
	out := taux.Zero()
	for _, aide := range bp.Aides {
		out.Add(aide.Montant)
	}
	return out.Montant
}

// Net renvoie le prix à payer après avoir déduit les aides extérieures et les remises
func (bp BilanParticipant) Net(taux ds.Taux) cps.Montant {
	v1 := bp.prixSansRemises(taux).Remise(bp.Remises.ReducEnfants + bp.Remises.ReducEquipiers)
	val := taux.Convertible(v1)
	val.Sub(bp.Remises.ReducSpeciale)
	if val.Montant.Cent < 0 {
		val.Montant.Cent = 0
	}
	return val.Montant
}

// Aide shows resolved [Aide]
type Aide struct {
	Structure string
	Montant   cps.Montant
}

func (p pc) finances() (out BilanParticipant) {
	out.Base = p.prixBase()
	out.Remises = p.Participant.Remises

	duree := p.duree()
	for _, aide := range p.aides {
		out.Aides = append(out.Aides, Aide{p.structures[aide.IdStructureaide].Nom, aide.Resolve(duree)})
	}
	return out
}

// duree renvoie le nombre de jours de présence du participant
// en prenant en compte une éventuelle option SEMAINE ou JOUR.
func (p pc) duree() int {
	options := p.Camp.OptionPrix
	optPart := p.Participant.OptionPrix

	switch options.Active {
	case cps.PrixSemaine:
		detailsOpt := options.Semaine
		switch optPart.Semaine {
		case cps.Semaine1:
			return detailsOpt.Plage1.Duree
		case cps.Semaine2:
			return detailsOpt.Plage2.Duree
		}
	case cps.PrixJour:
		return optPart.Jour.NbJours(p.Camp.Duree)
	}
	return p.Camp.Duree
}

// prixBase renvoie le prix du camp, en prenant en compte une éventuelle option
func (p pc) prixBase() cps.Montant {
	optPart := p.Participant.OptionPrix
	optCamp := p.Camp.OptionPrix

	prix := p.Camp.Prix
	switch optCamp.Active {
	case cps.PrixSemaine:
		semaine := optPart.Semaine
		if semaine == cps.Semaine1 {
			prix = optCamp.Semaine.Prix1
		} else if semaine == cps.Semaine2 {
			prix = optCamp.Semaine.Prix2
		}
	case cps.PrixStatut:
		for _, info := range optCamp.Statuts {
			if info.Id == optPart.IdStatut {
				prix = info.Prix
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
		}
	}

	// réduction quotient familial
	qf := p.Participant.QuotientFamilial
	if qfCamp := p.Camp.OptionQuotientFamilial; qfCamp.IsActive() && qf > 0 {
		ratio := qfCamp.Percentage(qf)
		prix.Cent = prix.Cent * ratio / 100
	}

	return prix
}

// BilanFinances résume l'état financier d'un dossier
type BilanFinances struct {
	currency ds.Currency

	// en centimes

	demande int // montant demandé final (avant déduction des paiements déjà effectués)
	recu    int // somme des paiements reçus
}

func (b BilanFinances) IsAcquitte() bool { return b.demande <= b.recu }

// ApresPaiement renvoie le montant restant à payer
func (b BilanFinances) ApresPaiement() ds.Montant {
	return ds.Montant{Cent: b.demande - b.recu, Currency: b.currency}
}

type StatutPaiement uint8

const (
	_ StatutPaiement = iota
	Complet
	EnCours
	NonCommence
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
