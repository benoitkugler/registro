package logic

import (
	"errors"
	"fmt"
	"time"

	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	sh "registro/sql/shared"
	"registro/utils"
)

// DossiersFinances adds aides and paiements
type DossiersFinances struct {
	Dossiers

	taux              ds.Tauxs
	aides             map[cps.IdParticipant]cps.Aide
	aidesFiles        map[cps.IdAide]fs.File        // enough for [aides]
	aidesFilesAttente map[cps.IdParticipant]fs.File // justificatifs en attente d'une aide
	qfFiles           map[ds.IdDossier]fs.File      // attestation de quotient familial, au plus une par dossier
	structures        cps.Structureaides
	paiements         map[ds.IdDossier]ds.Paiements
}

// LoadDossiersFinance is a convenient wrapper around [LoadDossiersFinances]
func LoadDossiersFinance(db ds.DB, id ds.IdDossier) (DossierFinance, error) {
	ld, err := LoadDossiersFinances(db, id)
	if err != nil {
		return DossierFinance{}, err
	}
	if _, has := ld.Dossiers.Dossiers[id]; !has {
		return DossierFinance{}, errors.New("Dossier introuvable")
	}
	return ld.For(id), nil
}

// LoadDossiersFinances wraps the SQL error
func LoadDossiersFinances(db ds.DB, ids ...ds.IdDossier) (DossiersFinances, error) {
	dossiers, err := LoadDossiers(db, ids)
	if err != nil {
		return DossiersFinances{}, err
	}

	tauxs, err := ds.SelectTauxs(db, dossiers.Dossiers.IdTauxs()...)
	if err != nil {
		return DossiersFinances{}, utils.SQLError(err)
	}

	aides, err := cps.SelectAidesByIdParticipants(db, dossiers.participantsIDs...)
	if err != nil {
		return DossiersFinances{}, utils.SQLError(err)
	}
	// justificatifs, y compris ceux en attente d'une aide
	links, err := fs.SelectFileAidesByIdParticipants(db, dossiers.participantsIDs...)
	if err != nil {
		return DossiersFinances{}, utils.SQLError(err)
	}
	files, err := fs.SelectFiles(db, links.IdFiles()...)
	if err != nil {
		return DossiersFinances{}, utils.SQLError(err)
	}
	aidesFiles := make(map[cps.IdAide]fs.File)
	aidesFilesAttente := make(map[cps.IdParticipant]fs.File)
	for _, link := range links {
		file := files[link.IdFile]
		if link.IdAide.Valid {
			aidesFiles[link.IdAide.Id] = file
		} else {
			aidesFilesAttente[link.IdParticipant] = file
		}
	}

	// attestation de quotient familial : au plus un fichier par dossier
	qfLinks, err := fs.SelectFileQuotientFamilialsByIdDossiers(db, ids...)
	if err != nil {
		return DossiersFinances{}, utils.SQLError(err)
	}
	qfFilesByID, err := fs.SelectFiles(db, qfLinks.IdFiles()...)
	if err != nil {
		return DossiersFinances{}, utils.SQLError(err)
	}
	qfFiles := make(map[ds.IdDossier]fs.File, len(qfLinks))
	for _, link := range qfLinks {
		qfFiles[link.IdDossier] = qfFilesByID[link.IdFile]
	}

	aidesByParticipant := aides.ByIdParticipant()

	structures, err := cps.SelectStructureaides(db, aides.IdStructureaides()...)
	if err != nil {
		return DossiersFinances{}, utils.SQLError(err)
	}
	paiements, err := ds.SelectPaiementsByIdDossiers(db, ids...)
	if err != nil {
		return DossiersFinances{}, utils.SQLError(err)
	}

	return DossiersFinances{dossiers, tauxs, aidesByParticipant, aidesFiles, aidesFilesAttente, qfFiles, structures, paiements.ByIdDossier()}, nil
}

func (df DossiersFinances) For(id ds.IdDossier) DossierFinance {
	out := df.Dossiers.For(id)
	aides := make(map[cps.IdParticipant]cps.Aide)
	aidesFilesAttente := make(map[cps.IdParticipant]fs.File)
	for _, part := range out.Participants {
		if aide, ok := df.aides[part.Id]; ok {
			aides[part.Id] = aide
		}
		if file, ok := df.aidesFilesAttente[part.Id]; ok {
			aidesFilesAttente[part.Id] = file
		}
	}
	return DossierFinance{out, df.taux[out.Dossier.IdTaux], aides, df.aidesFiles, aidesFilesAttente, df.qfFiles[id], df.structures, df.paiements[id]}
}

type DossierFinance struct {
	Dossier

	Taux ds.Taux

	aides             map[cps.IdParticipant]cps.Aide // at most one per participant
	aidesFiles        map[cps.IdAide]fs.File         // enough for [aides]
	aidesFilesAttente map[cps.IdParticipant]fs.File  // justificatifs en attente d'une aide
	qfFile            fs.File                        // attestation de quotient familial, zéro si absente
	structures        cps.Structureaides             // enough for [aides]

	paiements ds.Paiements // liste exacte
}

// FilesToValidate returns the number of files awaiting backoffice validation:
// justificatifs without an aide, and the quotient familial attestation when
// it is still pending.
func (df DossierFinance) FilesToValidate() int {
	count := 0
	for _, file := range df.aidesFilesAttente {
		if file.ReviewStatus == pr.FileReviewPending {
			count++
		}
	}
	if df.qfFile.Id != 0 && df.qfFile.ReviewStatus == pr.FileReviewPending {
		count++
	}
	return count
}

// MontantsPayesParDevise returns the net payments in their original currencies.
func (de DossierFinance) MontantsPayesParDevise() ds.MultiCurrencies {
	var out ds.MultiCurrencies
	for _, paiement := range de.paiements {
		montant := paiement.Montant
		if paiement.IsRemboursement {
			montant.Cent = -montant.Cent
		}
		out.AddM(montant)
	}
	return out
}

// Bilan additionne le prix pour chaque participant et
// décompte les paiements effectués.
//
// Seuls les participants inscrits (en liste principale) sont pris en compte,
// les participants en liste d'attente étant ignorés.
//
// Les justificatifs en attente n'influencent pas le bilan : l'aide n'existe pas encore.
func (df DossierFinance) Bilan() BilanFinances {
	inscrits := map[cps.IdParticipant]BilanParticipant{}
	demande, aides, demandeEnAttente := df.Taux.Zero(), df.Taux.Zero(), df.Taux.Zero()
	recu, fondsSoutien := df.Taux.Zero(), df.Taux.Zero()

	for _, participant := range df.Participants {
		camp := df.camps[participant.IdCamp]
		if participant.Statut != cps.Inscrit {
			// simplify by just using the camp Prix
			demandeEnAttente.Add(camp.Prix)
			continue
		}
		data := partExt{
			Participant:   participant,
			Camp:          camp,
			dateNaissance: df.PersonneFor(participant).DateNaissance,
			inscription:   df.Dossier.Dossier.MomentInscription,
			aide:          df.aides[participant.Id], // valeur zéro si pas d'aide
			structures:    df.structures,
		}
		bilan := data.bilan()
		inscrits[participant.Id] = bilan
		demande.Add(bilan.Net(df.Taux))
		aides.Add(bilan.Aide.Montant)
	}

	for _, paiement := range df.paiements {
		montant := paiement.Montant
		if paiement.IsRemboursement {
			montant.Cent = -montant.Cent
		}

		recu.Add(montant)

		if isFondsSoutien := paiement.Payeur == ds.PayeurFondSoutien; isFondsSoutien {
			fondsSoutien.Add(montant)
		}
	}

	// [demande], [demandeEnAttente], [recu], [fondsSoutien] and [aides] have the same currency
	return BilanFinances{inscrits, demande.Cent, demandeEnAttente.Cent, recu.Cent, fondsSoutien.Cent, aides.Cent, demande.Currency}
}

// BilanFinances résume l'état financier d'un dossier
type BilanFinances struct {
	Inscrits map[cps.IdParticipant]BilanParticipant

	// totaux, en centimes

	demande          int // montant demandé final (avant déduction des paiements déjà effectués)
	demandeEnAttente int // pour les participants en liste d'attente
	recu             int // somme des paiements reçus

	fondsSoutien int // part de [recu] payé par le fonds de soutien

	aides int // total des aides

	currency ds.Currency
}

func (b BilanFinances) Recu() ds.Montant {
	return ds.Montant{Cent: b.recu, Currency: b.currency}
}

// FondsSoutien renvoie la part de [Recu] attribuée au fonds de soutien.
func (b BilanFinances) FondsSoutien() ds.Montant {
	return ds.Montant{Cent: b.fondsSoutien, Currency: b.currency}
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

func (sp StatutPaiement) String() string {
	switch sp {
	case NonCommence:
		return "Non commencé"
	case EnCours:
		return "En cours"
	case Complet:
		return "Complet"
	default:
		return "<invalid statut>"
	}
}

func (b BilanFinances) StatutPaiement() StatutPaiement {
	if b.IsAcquitte() {
		return Complet
	} else if b.recu > 0 {
		return EnCours
	} else {
		return NonCommence
	}
}

// RequiredPaiementState décrit ce qui est demandé à la famille
// dans l'espace personnel.
type RequiredPaiementState uint8

const (
	// aucune information de paiement n'est affichée
	PaiementDesactive RequiredPaiementState = iota // Paiement désactivé
	// seul l'acompte est à régler
	PaiementAcompte // Acompte
	// le solde complet est à régler
	PaiementTotal // Paiement total
)

func (pm RequiredPaiementState) String() string {
	switch pm {
	case PaiementDesactive:
		return "Paiement désactivé"
	case PaiementAcompte:
		return "Acompte"
	case PaiementTotal:
		return "Paiement total"
	default:
		return "<invalid mode>"
	}
}

// AcompteInfo résume l'acompte à régler pour un dossier.
type AcompteInfo struct {
	Requis bool // au moins un inscrit est concerné

	Du      ds.Montant // montant total de l'acompte, plafonné au montant total dû
	Restant ds.Montant // montant de l'acompte non encore réglé
}

// acompte calcule l'acompte dû par le dossier. Il est calculé sur le
// prix de base des séjours demandant un acompte, et plafonné au montant
// total dû (après aides et remises).
func (df DossierFinance) acompte(bilan BilanFinances, percent int) AcompteInfo {
	var out AcompteInfo
	if percent <= 0 {
		return out
	}
	total := df.Taux.Zero()
	for _, part := range df.Participants {
		if part.Statut != cps.Inscrit {
			continue
		}
		camp := df.camps[part.IdCamp]
		if !camp.DemandeAcompte {
			continue
		}
		total.Add(camp.Acompte(percent))
		out.Requis = true
	}
	if !out.Requis {
		return out
	}
	out.Du = total.Convert(bilan.currency)
	if out.Du.Cent > bilan.demande {
		out.Du.Cent = bilan.demande
	}
	if out.Du.Cent <= 0 { // nothing to ask for
		return AcompteInfo{}
	}
	out.Restant = ds.Montant{Cent: max(0, out.Du.Cent-bilan.recu), Currency: out.Du.Currency}
	return out
}

// isPaiementOpen returns [true] if the Espace perso
// should allow paiements
func (de *DossierFinance) isPaiementOpen() bool {
	for _, part := range de.Participants {
		// every participant must have been validated
		if part.Statut == cps.AStatuer {
			return false
		}
		// justificatif en attente de création de l'aide
		if _, ok := de.aidesFilesAttente[part.Id]; ok {
			return false
		}
	}
	return true
}

// PaiementMode détermine l'état des paiements dans l'espace personnel :
// aucun paiement tant que le dossier n'est pas validé, acompte seul
// tant qu'il n'est pas réglé, puis paiement du solde.
//
// Lorsque requireDemandePaiement est vrai, le paiement du solde attend
// l'envoi d'un évènement DemandePaiement. L'acompte reste toujours permis.
func (df DossierFinance) PaiementMode(percent int, requireDemandePaiement bool) RequiredPaiementState {
	if !df.isPaiementOpen() {
		return PaiementDesactive
	}
	bilan := df.Bilan()
	if bilan.IsAcquitte() {
		return PaiementDesactive
	}
	if ac := df.acompte(bilan, percent); ac.Requis && ac.Restant.Cent > 0 {
		return PaiementAcompte
	}
	if _, has := df.Events.HasDemandePaiement(); requireDemandePaiement && !has {
		return PaiementDesactive
	}
	return PaiementTotal
}

// MontantARegler returns the amount currently asked to the family,
// according to [PaiementMode]. I returns false when no
// paiement is currently required.
func (df DossierFinance) MontantARegler(percent int, requireDemandePaiement bool) (ds.Montant, bool) {
	switch df.PaiementMode(percent, requireDemandePaiement) {
	case PaiementAcompte:
		bilan := df.Bilan()
		return df.acompte(bilan, percent).Restant, true
	case PaiementTotal:
		return df.Bilan().ApresPaiement(), true
	default:
		return ds.Montant{}, false
	}
}

// DispatchPaiements split the received amount between all inscrits,
// proportionnal to the sejour price.
func (b BilanFinances) DispatchPaiements(taux ds.Taux) map[cps.IdParticipant]ds.Montant {
	totalAsked, totalReceived := b.demande, b.recu
	if len(b.Inscrits) == 0 || totalAsked == 0 {
		return nil
	}
	out := make(map[cps.IdParticipant]ds.Montant, len(b.Inscrits))
	currency := b.currency
	sum, maxId := 0, cps.IdParticipant(0)
	for id, inscrit := range b.Inscrits {
		asked := taux.Convertible(inscrit.Net(taux)).Convert(currency).Cent
		// use the proportion to dispatch the total
		received := asked * totalReceived / totalAsked
		out[id] = ds.Montant{Cent: received, Currency: currency}
		// be carefull for rounding
		sum += received
		if id > maxId {
			maxId = id
		}
	}
	// attribute any leftover arbitrarily, in a deterministic fashion
	out[maxId] = ds.Montant{Cent: out[maxId].Cent + totalReceived - sum, Currency: currency}

	return out
}

type partExt struct {
	cps.Participant
	cps.Camp

	dateNaissance sh.Date
	inscription   time.Time

	aide       cps.Aide           // valeur zéro si pas d'aide
	structures cps.Structureaides // enough for [aide]
}

// BilanParticipant exposes the finances for one [Participant]
//
// Le prix final est calculé en suivant 3 étapes :
//   - applique les options et le quotient familial
//   - applique les aides extérieureurs
//   - applique les remises internes
type BilanParticipant struct {
	// Prix de base du séjour
	PrixCamp ds.Montant

	AvecOption            cps.Montant // prend en compte une éventuelle option et le quotient familial
	AvecOptionDescription string      // courte description affichée dans la facture

	Remises cps.Remises

	// Aide extérieure éventuelle (Aide.Valid indique si elle existe)
	Aide AideResolved
}

// prixSansRemises renvoie le prix du séjour si on applique les aides (extérieures)
// mais pas les remises.
func (bp BilanParticipant) prixSansRemises(taux ds.Taux) cps.Montant {
	out := taux.Convertible(bp.AvecOption)
	// le montant d'une aide absente est nul
	out.Sub(bp.Aide.Montant)

	p := out.Montant
	if p.Cent < 0 {
		p.Cent = 0
	}
	return p
}

// Net renvoie le prix à payer après avoir déduit les aides extérieures et les remises
func (bp BilanParticipant) Net(taux ds.Taux) cps.Montant {
	v1 := bp.prixSansRemises(taux).Remise(bp.Remises.Famille + bp.Remises.Equipiers)
	val := taux.Convertible(v1)
	val.Sub(bp.Remises.Speciale)
	return val.Montant
}

// ResolvedRemises renvoie le total des remises appliquées à ce participant,
// comptés positivement.
func (bp BilanParticipant) ResolvedRemises(taux ds.Taux) cps.Montant {
	before := bp.prixSansRemises(taux)
	after := bp.Net(taux)
	// both montant have been converted to the same unit
	before.Cent -= after.Cent
	return before
}

// AideResolved est l'aide éventuelle d'un participant, avec son montant résolu.
type AideResolved struct {
	Valid     bool // faux si le participant n'a pas d'aide
	Structure string
	Montant   cps.Montant
}

func (ar AideResolved) String() string {
	if !ar.Valid {
		return ""
	}
	return fmt.Sprintf("%s : %s", ar.Structure, ar.Montant)
}

func (p partExt) bilan() (out BilanParticipant) {
	out.PrixCamp = p.Camp.Prix

	out.AvecOption, out.AvecOptionDescription = p.prixBase()
	out.Remises = p.Participant.Remises

	out.Aide = AideResolved{
		Valid:     p.aide != (cps.Aide{}),
		Structure: p.structures[p.aide.IdStructureaide].Nom,
		Montant:   p.aide.Resolve(p.duree()),
	}
	return out
}

// duree renvoie le nombre de jours de présence du participant
// en prenant en compte une éventuelle option JOUR.
func (p partExt) duree() int {
	options := p.Camp.OptionPrix
	optPart := p.Participant.OptionPrix

	switch options.Active {
	case cps.PrixJour:
		return optPart.Jour.NbJours(p.Camp.Duree)
	}
	return p.Camp.Duree
}

// prixBase renvoie le prix du séjour, en prenant en compte une éventuelle option et le quotient familial
// une courte description est aussi renvoyée
func (p partExt) prixBase() (cps.Montant, string) {
	optPart := p.Participant.OptionPrix
	optCamp := p.Camp.OptionPrix

	prix := p.Camp.Prix
	descOption, descQF := "", ""
	switch optCamp.Active {
	case cps.PrixInscriptionRapide:
		if optCamp.InscriptionRapide.Limite.AfterOrEqual(p.inscription) {
			descOption = "Inscription rapide"
			prix.Cent = optCamp.InscriptionRapide.Prix
		}
	case cps.PrixJour:
		nbJours := optPart.Jour.NbJours(p.Camp.Duree)
		prixJours := optCamp.PrixParJour
		if nbJours == 0 {
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
	case cps.PrixAge:
		age := p.Camp.AgeDebutCamp(p.dateNaissance)
		if prixAge, ok := optCamp.PrixParAge.Prix(age); ok {
			prix.Cent = prixAge
			descOption = fmt.Sprintf("Âge %d ans", age)
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

// DossierRecap is used for backoffice and directeurs
// to display the main information regarding one [Dossier].
type DossierRecap struct {
	Dossier     ds.Dossier
	Responsable pr.Personne
	Reglement   StatutPaiement
}

func (de *DossierFinance) Reglement() DossierRecap {
	return DossierRecap{de.Dossier.Dossier, de.Responsable(), de.Bilan().StatutPaiement()}
}
