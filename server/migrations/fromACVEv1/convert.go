package main

import (
	"fmt"
	"strings"

	rd "registro/migrations/fromACVEv1/rawdata"
	cps "registro/sql/camps"
	"registro/sql/dossiers"
	ev "registro/sql/events"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/sql/shared"
)

func joinNonEmpty(parts ...string) string {
	var out []string
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return strings.Join(out, " ; ")
}

func personneFromV1(v1 rd.Personne) pr.Personne {
	return pr.Personne{
		Identite: pr.Identite{
			Nom:           string(v1.Nom),
			Prenom:        string(v1.Prenom),
			Sexe:          sexe(v1.Sexe),
			DateNaissance: shared.NewDateFrom(v1.DateNaissance.Time()),
			Tels:          tels(v1.Tels),
			Mail:          string(v1.Mail),
			Adresse:       string(v1.Adresse),
			CodePostal:    string(v1.CodePostal),
			Ville:         string(v1.Ville),
			Pays:          pr.Pays(v1.Pays),
		},
		Publicite: pr.Publicite{
			VersionPapier: bool(v1.VersionPapier),
			PubHiver:      bool(v1.PubHiver),
			PubEte:        bool(v1.PubEte),
			EchoRocher:    bool(v1.EchoRocher),
			Eonews:        bool(v1.Eonews),
		},
		IsTemp:          bool(v1.IsTemporaire),
		SecuriteSociale: string(v1.SecuriteSociale),
	}
}

// ficheSanitaireFromV1 returns false when the v1 fiche was never filled in.
func ficheSanitaireFromV1(v1 rd.Personne) (pr.Fichesanitaire, bool) {
	if v1.FicheSanitaire.IsNone() {
		return pr.Fichesanitaire{}, false
	}
	fiche := v1.FicheSanitaire
	out := pr.Fichesanitaire{
		DifficultesSante: joinNonEmpty(
			string(fiche.DifficultesSante),
			string(fiche.Recommandations),
			strings.Join(fiche.Maladies.List(), ", "),
			strings.Join(fiche.Allergies.List(), ", "),
			fiche.Allergies.ConduiteATenir,
		),
		AllergiesAlimentaires: boolToOui(fiche.Allergies.Alimentaires),
		TraitementMedical:     boolToOui(fiche.TraitementMedical),
		AutreContact:          pr.NomTel{Tel: pr.Tel(fiche.Tel)},
		Modified:              fiche.LastModif.Time(),
		Owners:                pr.Mails(fiche.Mails),
		ReviewStatus:          pr.FileReviewPending,
	}
	if fiche.Handicap {
		out.DifficultesSante = joinNonEmpty(out.DifficultesSante, "Handicap")
	}
	return out, true
}

func boolToOui(v bool) string {
	if v {
		return "oui"
	}
	return ""
}

// ficheEquipierFromV1 uses the latest equipier record for the person, as v2
// stores a single fiche per person.
func ficheEquipierFromV1(personne rd.Personne, equipier rd.Equipier) pr.Ficheequipier {
	return pr.Ficheequipier{
		IdPersonne:     0, // filled by the caller after Personne.Insert
		Diplome:        diplome(equipier.Diplome),
		Qualifications: qualification(equipier.Appro),
		Fonctionnaire:  bool(personne.Fonctionnaire),
		Profession:     string(personne.Profession),
	}
}

// campIDPrixParAge is the only v1 camp whose "statut" price option can be
// translated into the v2 age-based grid.
const campIDPrixParAge int64 = 338

// statutsPrixV1 is the expected v1 pricing of camp 338, ordered by increasing
// AgeMax. An AgeMax of 0 marks the open-ended adult category, which is stored
// as the camp's base price in v2 instead of being part of the grid.
var statutsPrixV1 = [...]struct {
	id     int64
	label  string
	prix   rd.Euros
	ageMax int
}{
	{5, "0-2 ans", 0, 3},
	{4, "3-5 ans", 98, 6},
	{3, "6-9 ans", 154, 10},
	{2, "10-14 ans", 231, 15},
	{1, "Plus de 18 ans", 280, 0},
}

// specialPrixParAgeFromV1 detects the exact "statut" price option of camp 338
// and converts it to the v2 PrixParAge grid. Other camps and changed values
// keep the generic behavior (no v2 equivalent).
func specialPrixParAgeFromV1(idCamp int64, v1 rd.OptionPrixCamp) (cps.OptionPrixCamp, bool) {
	if idCamp != campIDPrixParAge || v1.Active != rd.OptionsPrix.STATUT || len(v1.Statut) != len(statutsPrixV1) {
		return cps.OptionPrixCamp{}, false
	}

	byID := make(map[int64]rd.PrixParStatut, len(v1.Statut))
	for _, statut := range v1.Statut {
		byID[statut.Id] = statut
	}
	grid := make(cps.PrixParAge, 0, len(byID)-1)
	for _, expected := range statutsPrixV1 {
		statut, ok := byID[expected.id]
		if !ok || statut.Prix != expected.prix || strings.TrimSpace(string(statut.Statut)) != expected.label {
			return cps.OptionPrixCamp{}, false
		}
		if expected.ageMax > 0 {
			grid = append(grid, cps.PrixAgeTranche{AgeMax: expected.ageMax, Prix: int(statut.Prix.Centimes())})
		}
	}
	return cps.OptionPrixCamp{Active: cps.PrixAge, PrixParAge: grid}, true
}

func optionPrixCampFromV1(v1 rd.OptionPrixCamp) (cps.OptionPrixCamp, bool) {
	switch v1.Active {
	case rd.OptionsPrix.JOUR:
		var price int64
		uniform := true
		for _, day := range v1.Jour {
			if day == 0 {
				continue
			}
			if price == 0 {
				price = day.Centimes()
			} else if price != day.Centimes() {
				uniform = false
			}
		}
		if price == 0 {
			return cps.OptionPrixCamp{}, false
		}
		return cps.OptionPrixCamp{Active: cps.PrixJour, PrixParJour: int(price)}, !uniform
	default:
		// v1 "semaine" and "statut" options have no v2 equivalent.
		return cps.OptionPrixCamp{}, false
	}
}

func prixQuotientFamilialFromV1(v1 rd.OptionQuotientFamilial) cps.PrixQuotientFamilial {
	var out cps.PrixQuotientFamilial
	for _, tranche := range v1 {
		// v1 stores an inactive grid as four zero-valued tranches. Skipping
		// zero percentages keeps that case distinct from a real first
		// tranche, whose legacy bound is also zero.
		if tranche.Pourcentage == 0 {
			continue
		}
		if tranche.QuotientMin < 0 {
			continue
		}
		// v1 bounds are upper bounds; v2 uses inclusive lower bounds, so
		// every bound must be shifted by one (0/359/564/714 -> 1/360/565/715).
		quotientMin := int(tranche.QuotientMin) + 1
		pourcentage := tranche.Pourcentage
		out = append(out, cps.TrancheQuotient{
			QuotientMin: quotientMin,
			Pourcentage: int(pourcentage + 0.5),
		})
	}
	if len(out) > 0 && out[len(out)-1].Pourcentage != 100 {
		out[len(out)-1].Pourcentage = 100
	}
	return out
}

func documentsToShowFromV1(v1 rd.Envois) cps.DocumentsToShow {
	return cps.DocumentsToShow{
		LettreDirecteur:   v1.LettreDirecteur,
		ListeVetements:    v1.ListeVetements,
		ListeParticipants: v1.ListeParticipants,
	}
}

func vetementsFromV1(v1 rd.ListeVetements) cps.ListeVetements {
	out := cps.ListeVetements{
		Vetements:  make([]cps.Vetement, 0, len(v1.Liste)),
		Complement: v1.Complement,
	}
	for _, v := range v1.Liste {
		out.Vetements = append(out.Vetements, cps.Vetement{
			Quantite:    v.Quantite,
			Description: v.Description,
			Important:   v.Obligatoire,
		})
	}
	return out
}

func campFromV1(v1 rd.Camp) cps.Camp {
	optionPrix, _ := optionPrixCampFromV1(v1.OptionPrix)
	if special, ok := specialPrixParAgeFromV1(v1.Id, v1.OptionPrix); ok {
		optionPrix = special
	}
	return cps.Camp{
		IdTaux:                 1, // the default taux, already created
		Nom:                    string(v1.Nom),
		DateDebut:              shared.NewDateFrom(v1.DateDebut.Time()),
		Duree:                  v1.Duree(),
		Lieu:                   string(v1.Lieu),
		Agrement:               string(v1.NumeroJS),
		Description:            string(v1.Infos),
		Navette:                navette(v1.Options.Bus),
		Places:                 int(v1.NbPlaces),
		AgeMin:                 int(v1.AgeMin),
		AgeMax:                 int(v1.AgeMax),
		NeedEquilibreGF:        bool(v1.NeedEquilibreGf),
		Statut:                 statutCamp(v1.Ouvert),
		Prix:                   cps.Montant{Cent: int(v1.Prix.Centimes()), Currency: dossiers.Euros},
		OptionPrix:             optionPrix,
		OptionQuotientFamilial: prixQuotientFamilialFromV1(v1.QuotientFamilial),
		Password:               string(v1.Password),
		DocumentsReady:         !v1.Envois.Locked,
		DocumentsToShow:        documentsToShowFromV1(v1.Envois),
		Vetements:              vetementsFromV1(v1.ListeVetements),
		// v1 Joomeo album IDs are deprecated: do not carry them into v2.
		AlbumID:         "",
		IsAlbumVisible:  false,
		PlacesReservees: int(v1.NbPlacesReservees),
	}
}

func structureAideFromV1(v1 rd.Structureaide) cps.Structureaide {
	return cps.Structureaide{
		Nom:             string(v1.Nom),
		Immatriculation: string(v1.Immatriculation),
		Info:            joinNonEmpty(string(v1.Info), string(v1.Adresse), string(v1.CodePostal), string(v1.Ville), string(v1.Telephone)),
	}
}

func aideFromV1(v1 rd.Aide) cps.Aide {
	return cps.Aide{
		IdStructureaide: 0, // filled by the caller
		IdParticipant:   0, // filled by the caller
		Valide:          bool(v1.Valide),
		Valeur:          euros(v1.Valeur),
		ParJour:         bool(v1.ParJour),
		NbJoursMax:      int(v1.NbJoursMax),
	}
}

func satisfaction(v1 rd.Satisfaction) cps.Satisfaction {
	// v1 and v2 use the same 0..4 encoding.
	return cps.Satisfaction(v1)
}

func sondageFromV1(v1 rd.Sondage) cps.Sondage {
	return cps.Sondage{
		IdCamp:    0, // filled by the caller
		IdDossier: 0, // filled by the caller
		Modified:  v1.Modified,
		ReponseSondage: cps.ReponseSondage{
			InfosAvantSejour:   satisfaction(v1.InfosAvantSejour),
			InfosPendantSejour: satisfaction(v1.InfosPendantSejour),
			Hebergement:        satisfaction(v1.Hebergement),
			Activites:          satisfaction(v1.Activites),
			Theme:              satisfaction(v1.Theme),
			Nourriture:         satisfaction(v1.Nourriture),
			Hygiene:            satisfaction(v1.Hygiene),
			Ambiance:           satisfaction(v1.Ambiance),
			Ressenti:           satisfaction(v1.Ressenti),
			MessageEnfant:      string(v1.MessageEnfant),
			MessageResponsable: string(v1.MessageResponsable),
		},
	}
}

func groupeFromV1(v1 rd.Groupe) cps.Groupe {
	return cps.Groupe{
		IdCamp:  0, // filled by the caller
		Nom:     string(v1.Nom),
		Couleur: v1.Couleur,
		Fin:     shared.NewDateFrom(v1.Plage.To.Time()),
	}
}

func paiementFromV1(v1 rd.Paiement) dossiers.Paiement {
	bordereau := ""
	if date := v1.InBordereau.Time(); !date.IsZero() {
		bordereau = date.Format("02/01/2006")
	}
	details := joinNonEmpty(
		string(v1.Details),
		string(v1.NomBanque),
		bordereau,
	)
	if v1.IsAcompte {
		details = joinNonEmpty(details, "acompte")
	}
	return dossiers.Paiement{
		IdDossier:       0, // filled by the caller
		IsRemboursement: bool(v1.IsRemboursement),
		Montant:         euros(v1.Valeur),
		Payeur:          string(v1.LabelPayeur),
		Mode:            modePaiement(v1.ModePaiement),
		Time:            v1.DateReglement.Time(),
		Label:           string(v1.Numero),
		Details:         details,
	}
}

// categorieFromV1 maps the legacy builtin constraints to v2 categories.
// An empty builtin denotes a v1 custom constraint, and test_nautique is
// intentionally kept as a custom v2 demande. Other unmatched builtins have no
// v2 equivalent and are dropped by isObsoleteBuiltin.
func categorieFromV1(v1 rd.BuiltinContrainte) (fs.Categorie, string) {
	label := rd.BuiltinContrainteLabels[v1]
	switch v1 {
	case rd.CCarteId:
		return fs.CarteId, ""
	case rd.CPermis:
		return fs.Permis, ""
	case rd.CSb:
		return fs.SB, ""
	case rd.CSecour:
		return fs.Secourisme, ""
	case rd.CBafa, rd.CBafaEquiv:
		return fs.Bafa, ""
	case rd.CBafd, rd.CBafdEquiv:
		return fs.Bafd, ""
	case rd.CCarteVitale:
		return fs.CarteVitale, ""
	case rd.CVaccin:
		return fs.Vaccins, ""
	case rd.CAutre:
		return fs.Autre, ""
	default:
		return fs.NoBuiltin, label
	}
}

// isObsoleteBuiltin reports whether a v1 builtin constraint must be dropped.
// Empty builtins are custom constraints, and test_nautique has no predeclared
// v2 row but is deliberately preserved as a custom demande.
func isObsoleteBuiltin(v1 rd.BuiltinContrainte) bool {
	if v1 == rd.CInvalide || v1 == rd.CTestNautique {
		return false
	}
	categorie, _ := categorieFromV1(v1)
	return categorie == fs.NoBuiltin
}

func fileFromV1(v1 rd.Document) fs.File {
	return fs.File{
		Taille:       int(v1.Taille),
		NomClient:    string(v1.NomClient),
		Uploaded:     v1.DateHeureModif.Time(),
		ReviewStatus: pr.FileReviewPending,
	}
}

func demandeFromV1(v1 rd.Contrainte, categorie fs.Categorie, legacyLabel string, idFile fs.OptIdFile, idDirecteur pr.OptIdPersonne) fs.Demande {
	description := joinNonEmpty(string(v1.Nom), legacyLabel, string(v1.Description))
	if categorie != fs.NoBuiltin {
		description = "" // builtin labels are provided by v2
	}
	maxDocs := v1.MaxDocs
	if maxDocs < 1 {
		maxDocs = 1
	}
	return fs.Demande{
		IdFile:      idFile,
		IdDirecteur: idDirecteur,
		Categorie:   categorie,
		Description: description,
		MaxDocs:     maxDocs,
	}
}

func eventKindFromV1(v1 rd.MessageKind) (ev.EventKind, bool) {
	switch v1 {
	case rd.MSupprime:
		return ev.Supprime, true
	case rd.MResponsable, rd.MCentre:
		return ev.Message, true
	case rd.MInscription:
		// Ignored by trimMessages : the inscription date is already stored
		// on the participant.
		return 0, false
	case rd.MAccuseReception:
		return ev.Statuation, true
	case rd.MFacture:
		return ev.Facture, true
	case rd.MDocuments:
		return ev.CampDocs, true
	case rd.MFactureAcquittee, rd.MAttestationPresence:
		return ev.Attestation, true
	case rd.MSondage:
		return ev.Sondage, true
	case rd.MPlaceLiberee:
		return ev.PlaceLiberee, true
	default:
		// MPaiement is never used in v1 and has no v2 equivalent.
		return 0, false
	}
}

func eventFromV1(v1 rd.Message, idDossier dossiers.IdDossier) (ev.Event, bool) {
	kind, ok := eventKindFromV1(v1.Kind)
	if !ok {
		return ev.Event{}, false
	}
	return ev.Event{
		IdDossier: idDossier,
		Kind:      kind,
		Created:   v1.Created.Time(),
	}, true
}

func distributionFromV1(v1 rd.Distribution) ev.Distribution {
	return ev.Distribution(v1)
}

func eventMessageFromV1(v1 rd.Message, contenu string) ev.EventMessage {
	message := ev.EventMessage{Contenu: contenu}
	switch v1.Kind {
	case rd.MResponsable:
		message.Origine = ev.Espaceperso
		message.VuBackoffice = v1.Vu
	case rd.MCentre:
		message.Origine = ev.Backoffice
		message.VuEspaceperso = v1.Vu
	default:
		message.Origine = ev.Backoffice
	}
	return message
}

func eventMessageVusFromV1(views rd.MessageViews, campM map[int64]cps.IdCamp) []ev.EventMessageVu {
	out := make([]ev.EventMessageVu, 0, len(views))
	for _, view := range views {
		if !view.Vu {
			continue
		}
		idCamp, ok := campM[view.IdCamp]
		if !ok {
			continue
		}
		out = append(out, ev.EventMessageVu{IdCamp: idCamp})
	}
	return out
}

func warningf(format string, args ...interface{}) {
	fmt.Printf("\tWARNING: "+format+"\n", args...)
}
