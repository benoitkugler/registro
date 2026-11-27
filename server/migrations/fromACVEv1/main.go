// Script de migration depuis la v1 ACVE.
//
// Ce script suppose que la base v2 a déjà été créée (sans données) et
//   - charge les données v1
//   - nettoie en conservant uniquement les années récentes
//   - convertit au format v2 et insère
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"

	"registro/config"
	"registro/crypto"
	comptaV1 "registro/migrations/fromACVEv1/compta"
	rd "registro/migrations/fromACVEv1/rawdata"
	cps "registro/sql/camps"
	"registro/sql/compta"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	commit := flag.Bool("commit", false, "actually write the v2 DB")
	v1FilesDir := flag.String("v1-files-dir", "", "optional directory containing legacy document files named file_<id>")
	v2FilesDir := flag.String("v2-files-dir", "", "destination directory for v2 document files (required with -v1-files-dir)")
	host := flag.String("host", "", "public host used in migrated lettre image URLs (required when migrating lettre images)")
	skipDocumentContents := flag.Bool("skip-document-contents", true, "skip the heavy contenu_documents table by default; pass =false to load it")
	flag.Parse()

	if *commit && *skipDocumentContents {
		log.Fatal("-commit requires contenu_documents for miniatures: pass -skip-document-contents=false")
	}

	fmt.Println("Connecting to v1...")
	dbV1, err := connectV1()
	check(err)
	defer dbV1.Close()
	fmt.Println("Done.")

	fmt.Println("Loading v1 data...")
	if *skipDocumentContents {
		fmt.Println("\tskipping contenu_documents (-skip-document-contents)")
	}
	originV1, err := loadV1(dbV1, *skipDocumentContents)
	check(err)
	fmt.Printf("Done. Trimming with CampStart = %d and DonsStart = %d...\n", campsStartYear, donsStartYear)
	trimmedV1 := originV1.trim()
	fmt.Println("Done :")

	fmt.Println("\tpersonnes:", len(originV1.personnes), "->", len(trimmedV1.personnes))
	fmt.Println("\tdocuments:", len(originV1.documents), "->", len(trimmedV1.documents))
	fmt.Println("\tcamps:", len(originV1.camps), "->", len(trimmedV1.camps))
	fmt.Println("\tfactures:", len(originV1.factures), "->", len(trimmedV1.factures))
	fmt.Println("\tparticipants:", len(originV1.participants), "->", len(trimmedV1.participants))
	fmt.Println("\tequipiers:", len(originV1.equipiers), "->", len(trimmedV1.equipiers))
	fmt.Println("\tcomptas:", len(originV1.comptas), "->", len(trimmedV1.comptas))
	fmt.Println("\trecords:", len(originV1.records), "->", len(trimmedV1.records))
	fmt.Println("\tsubrecords:", len(originV1.subrecords), "->", len(trimmedV1.subrecords))
	fmt.Println("\tbusRecords:", len(originV1.busRecords), "->", len(trimmedV1.busRecords))
	fmt.Println("\tfilecomptas:", len(originV1.filecomptas), "->", len(trimmedV1.filecomptas))
	fmt.Println("\tbudgets:", len(originV1.budgets), "->", len(trimmedV1.budgets))
	fmt.Println("\tbudgetConfigs:", len(originV1.budgetConfigs), "->", len(trimmedV1.budgetConfigs))
	fmt.Println("\tpaiements:", len(originV1.paiements), "->", len(trimmedV1.paiements))
	fmt.Println("\taides:", len(originV1.aides), "->", len(trimmedV1.aides))
	fmt.Println("\tstructureaides:", len(originV1.structureaides), "->", len(trimmedV1.structureaides))
	fmt.Println("\tsondages:", len(originV1.sondages), "->", len(trimmedV1.sondages))
	fmt.Println("\tgroupes:", len(originV1.groupes), "->", len(trimmedV1.groupes))
	fmt.Println("\tcontraintes:", len(originV1.contraintes), "->", len(trimmedV1.contraintes))
	fmt.Println("\tdocumentCamps:", len(originV1.documentCamps), "->", len(trimmedV1.documentCamps))
	fmt.Println("\tdocumentPersonnes:", len(originV1.documentPersonnes), "->", len(trimmedV1.documentPersonnes))
	fmt.Println("\tdocumentAides:", len(originV1.documentAides), "->", len(trimmedV1.documentAides))
	fmt.Println("\tlettredirecteurs:", len(originV1.lettredirecteurs), "->", len(trimmedV1.lettredirecteurs))
	fmt.Println("\timageuploadeds:", len(originV1.imageuploadeds), "->", len(trimmedV1.imageuploadeds))
	fmt.Println("\tmessages:", len(originV1.messages), "->", len(trimmedV1.messages))
	fmt.Println("\tmessageDocuments:", len(originV1.messageDocuments), "->", len(trimmedV1.messageDocuments))
	fmt.Println("\tmessageAttestations:", len(originV1.messageAttestations), "->", len(trimmedV1.messageAttestations))
	fmt.Println("\tmessageSondages:", len(originV1.messageSondages), "->", len(trimmedV1.messageSondages))
	fmt.Println("\tmessagePlaceliberes:", len(originV1.messagePlaceliberes), "->", len(trimmedV1.messagePlaceliberes))
	fmt.Println("\tmessageMessages:", len(originV1.messageMessages), "->", len(trimmedV1.messageMessages))
	fmt.Println("\tmessageViews:", len(originV1.messageViews), "->", len(trimmedV1.messageViews))

	opts := migrationOptions{
		v1FilesDir:           *v1FilesDir,
		v2FilesDir:           *v2FilesDir,
		host:                 *host,
		skipDocumentContents: *skipDocumentContents,
	}

	check(checkAidesUnicity(trimmedV1))

	for _, warning := range warningsFromV1(trimmedV1, opts) {
		warningf("%s", warning)
	}

	if *commit {
		fmt.Println("Validating configuration...")
		check(validateMigration(trimmedV1, opts))
		fmt.Println("Done.")

		fmt.Println("Connecting to v2...")
		dbV2, err := connectV2()
		check(err)
		defer dbV2.Close()
		fmt.Println("Done.")

		keys, err := config.NewKeys()
		check(err)
		opts.encrypter = crypto.NewEncrypter(keys.EncryptKey)

		fmt.Println("Inserting in v2...")
		err = insert(dbV2, originV1, trimmedV1, opts)
		check(err)
		fmt.Println("Done.")
	}
}

func euros(v rd.Euros) ds.Montant { return ds.NewEuros(float64(v)) }

func modePaiement(v rd.ModePaiment) ds.ModePaiement {
	switch v {
	default:
		panic("unsupported ModePaiement " + v)
	case "": // unfortunate
		return ds.Especes
	case rd.MPVirement:
		return ds.Virement
	case rd.MPCheque:
		return ds.Cheque
	case rd.MPEspece:
		return ds.Especes
	case rd.MPCarte:
		return ds.EnLigne
	case rd.MPAncv:
		return ds.Ancv
	case rd.MPHelloasso:
		return ds.Helloasso
	}
}

func sexe(v rd.Sexe) pr.Sexe {
	switch v {
	case rd.SHomme:
		return pr.Man
	case rd.SFemme:
		return pr.Woman
	default:
		return pr.NoSexe
	}
}

func tels(tels rd.Tels) (out pr.Tels) {
	copy(out[:], tels)
	return
}

func diplome(v rd.Diplome) pr.Diplome {
	switch v {
	default:
		return pr.DAucun
	case rd.DBafa:
		return pr.DAcveBafa
	case rd.DBafaStag:
		return pr.DAcveBafaStag
	case rd.DBafd:
		return pr.DAcveBafd
	case rd.DBafdStag:
		return pr.DAcveBafdStag
	case rd.DCap:
		return pr.DAcveCap
	case rd.DAssSociale:
		return pr.DAcveAssSociale
	case rd.DEducSpe:
		return pr.DAcveEducSpe
	case rd.DMonEduc:
		return pr.DAcveMonEduc
	case rd.DInstit:
		return pr.DAcveInstit
	case rd.DProf:
		return pr.DAcveProf
	case rd.DAgreg:
		return pr.DAcveAgreg
	case rd.DBjeps:
		return pr.DAcveBjeps
	case rd.DDut:
		return pr.DAcveDut
	case rd.DEje:
		return pr.DAcveEje
	case rd.DDeug:
		return pr.DAcveDeug
	case rd.DStaps:
		return pr.DAcveStaps
	case rd.DBapaat:
		return pr.DAcveBapaat
	case rd.DBeatep:
		return pr.DAcveBeatep
	case rd.DZzautre:
		return pr.DAutre
	}
}

func qualification(v rd.Approfondissement) pr.Qualifications {
	switch v {
	default:
		return pr.Qualifications{}
	case rd.ASb:
		return pr.Qualifications{SB: true}
	case rd.ACanoe:
		return pr.Qualifications{Canoe: true}
	case rd.AVoile:
		return pr.Qualifications{Voile: true}
	}
}

func navette(bus rd.BusCamp) cps.OptionNavette {
	return cps.OptionNavette{Actif: bus.Actif, Commentaire: string(bus.Commentaire)}
}

func statutCamp(ouvert rd.Bool) cps.StatutCamp {
	if ouvert {
		return cps.Ouvert
	}
	return cps.Ferme
}

func roles(roles rd.Roles) cps.Roles {
	out := make(cps.Roles, len(roles))
	for i, r := range roles {
		switch r {
		case rd.RDirecteur:
			out[i] = cps.Direction
		case rd.RAdjoint:
			out[i] = cps.Adjoint
		case rd.RAnimation:
			out[i] = cps.Animation
		case rd.RAideAnimation:
			out[i] = cps.AideAnimation
		case rd.RChauffeur:
			out[i] = cps.Chauffeur
		case rd.RIntend:
			out[i] = cps.Intendance
		case rd.RBabysiter:
			out[i] = cps.Babysiter
		case rd.RMen:
			out[i] = cps.Menage
		case rd.RFactotum:
			out[i] = cps.Factotum
		case rd.RInfirm:
			out[i] = cps.Infirmerie
		case rd.RCuis:
			out[i] = cps.Cuisine
		case rd.RLing:
			out[i] = cps.Lingerie
		case rd.RAutre:
			out[i] = cps.AutreRole
		default:
			out[i] = cps.AutreRole
		}
	}
	return out
}

func presenceOffsets(ps rd.OptionnalPlage, camp shared.Plage) cps.PresenceOffsets {
	if !ps.Active {
		return cps.PresenceOffsets{}
	}
	debutE, finE := ps.Plage.From.Time(), ps.Plage.To.Time()
	diffDebut, diffFin := debutE.Sub(camp.From.Time()), finE.Sub(camp.To().Time())
	return cps.PresenceOffsets{Debut: int(diffDebut.Hours() / 24), Fin: int(diffFin.Hours() / 24)}
}

func nullBool(optionnalBool rd.OptionnalBool) sql.NullBool {
	switch optionnalBool {
	case rd.OBNon:
		return sql.NullBool{Valid: true, Bool: false}
	case rd.OBOui:
		return sql.NullBool{Valid: true, Bool: true}
	case rd.OBPeutEtre:
		return sql.NullBool{Valid: false}
	default:
		panic("exhaustive switch")
	}
}

func formStatusEquipier(invitationEquipier rd.InvitationEquipier) cps.FormStatusEquipier {
	switch invitationEquipier {
	case rd.NonInvite:
		return cps.NotSend
	case rd.Invite:
		return cps.Pending
	case rd.Verifie:
		return cps.Answered
	default:
		panic("exhaustive switch")
	}
}

func reviews(rv comptaV1.Reviews) compta.Reviews {
	out := make(compta.Reviews, len(rv))
	for i, r := range rv {
		out[i] = compta.Review(r)
	}
	return out
}

func montants(rv comptaV1.Montants) compta.Montants {
	out := make(compta.Montants, len(rv))
	for i, r := range rv {
		out[i] = compta.MontantPart{Cent: compta.Cent(r.Montant), Kind: compta.Ventilation(r.Kind)}
	}
	return out
}

func amountFromV1(v comptaV1.Cent) ds.Montant {
	return ds.Montant{Cent: int(v), Currency: ds.Euros}
}

// Legacy-only fields without a compatible destination representation are omitted :
// we loose a bit of precision in Repas
func budgetFromV1(v1 comptaV1.ComptaBudget, newID compta.IdComptaCamp) compta.ComptaBudget {
	out := compta.ComptaBudget{
		IdCompta:           newID,
		SimulatedInscrits:  v1.SimulatedInscrits,
		SimulatedEquipiers: v1.SimulatedEquipiers,
		MaterielPeda:       amountFromV1(v1.MaterielPeda),
		ProduitPharma:      amountFromV1(v1.ProduitPharma),
		TeamTrips:          amountFromV1(v1.TeamTrips),
		ExternalCarRent:    amountFromV1(v1.ExternalCarRent),
		AssoCarKms:         compta.CarsKms(v1.CarsMDRKms),
		ExternalHost:       make(compta.Rents, len(v1.ExternalHost)),
		Activities:         make(compta.Activities, len(v1.Activities)),
	}
	for i, rent := range v1.ExternalHost {
		out.ExternalHost[i] = compta.NamedCost{Name: rent.Name, Cost: amountFromV1(rent.Cost)}
	}
	for i, activity := range v1.Activities {
		out.Activities[i] = compta.NamedCost{Name: activity.Name, Cost: amountFromV1(activity.Cost)}
	}
	return out
}

func budgetConfigFromV1(v1 comptaV1.Budgetconfig) compta.ComptaConfig {
	out := compta.ComptaConfig{
		Year:                v1.Year,
		SmallMealThreshold:  v1.SmallMealThreshold,
		SmallMealMajoration: v1.SmallMealMajoration,
		ChauffageMDR:        int(v1.ChauffageMDR),
		EntretienMDR:        int(v1.EntretienMDR),
		MaterielMDR:         int(v1.MaterielMDR),
	}
	for i, cost := range v1.MealsCost {
		out.MealsCost[i] = int32(cost)
	}
	for i, cost := range v1.CarsMDR {
		out.AssoCars[i] = int32(cost)
	}
	return out
}

type baseV1 struct {
	personnes  rd.Personnes
	dons       rd.Dons
	donateurs  rd.DonDonateurs
	organismes rd.Organismes

	camps rd.Camps

	equipiers rd.Equipiers

	factures     rd.Factures
	participants rd.Participants
	paiements    rd.Paiements

	aides          rd.Aides
	structureaides rd.Structureaides
	sondages       rd.Sondages

	groupes            rd.Groupes
	groupeParticipants rd.GroupeParticipants

	// v1 'groupe_equipiers' and 'participant_equipiers' are intentionally
	// not migrated : v2 has no equivalent.

	contraintes         rd.Contraintes
	campContraintes     rd.CampContraintes
	equipierContraintes rd.EquipierContraintes

	// v1 'groupe_contraintes' is intentionally not migrated : v2 has no equivalent.

	documents         rd.Documents
	contenuDocuments  rd.ContenuDocuments
	documentCamps     rd.DocumentCamps
	documentPersonnes rd.DocumentPersonnes
	documentAides     rd.DocumentAides

	lettredirecteurs rd.Lettredirecteurs
	imageuploadeds   rd.Imageuploadeds

	messages            rd.Messages
	messageDocuments    rd.MessageDocuments
	messageMessages     rd.MessageMessages
	messagePlaceliberes rd.MessagePlaceliberes
	messageAttestations rd.MessageAttestations
	messageSondages     rd.MessageSondages
	messageViews        rd.MessageViews

	// v1 'doncs' is intentionally not migrated : v2 has no equivalent.

	comptas       comptaV1.Comptas
	records       comptaV1.Records
	subrecords    comptaV1.Subrecords
	filecomptas   comptaV1.Filecomptas
	budgets       comptaV1.ComptaBudgets
	budgetConfigs comptaV1.Budgetconfigs
	busRecords    comptaV1.RecordBuss
}

func loadV1(db *sql.DB, skipDocumentContents bool) (out baseV1, err error) {
	out.personnes, err = rd.SelectAllPersonnes(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table personnes: %s", err)
	}
	out.dons, err = rd.SelectAllDons(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table dons: %s", err)
	}
	out.donateurs, err = rd.SelectAllDonDonateurs(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table donateurs: %s", err)
	}
	out.organismes, err = rd.SelectAllOrganismes(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table organismes: %s", err)
	}
	out.camps, err = rd.SelectAllCamps(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table camps: %s", err)
	}
	out.equipiers, err = rd.SelectAllEquipiers(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table equipiers: %s", err)
	}
	out.factures, err = rd.SelectAllFactures(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table factures: %s", err)
	}
	out.participants, err = rd.SelectAllParticipants(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table participants: %s", err)
	}
	out.paiements, err = rd.SelectAllPaiements(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table paiements: %s", err)
	}
	out.aides, err = rd.SelectAllAides(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table aides: %s", err)
	}
	out.structureaides, err = rd.SelectAllStructureaides(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table structureaides: %s", err)
	}
	out.sondages, err = rd.SelectAllSondages(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table sondages: %s", err)
	}
	out.groupes, err = rd.SelectAllGroupes(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table groupes: %s", err)
	}
	out.groupeParticipants, err = rd.SelectAllGroupeParticipants(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table groupe_participants: %s", err)
	}
	out.contraintes, err = rd.SelectAllContraintes(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table contraintes: %s", err)
	}
	out.campContraintes, err = rd.SelectAllCampContraintes(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table camp_contraintes: %s", err)
	}
	out.equipierContraintes, err = rd.SelectAllEquipierContraintes(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table equipier_contraintes: %s", err)
	}
	out.documents, err = rd.SelectAllDocuments(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table documents: %s", err)
	}
	if !skipDocumentContents {
		out.contenuDocuments, err = rd.SelectAllContenuDocuments(db)
		if err != nil {
			return baseV1{}, fmt.Errorf("table contenu_documents: %s", err)
		}
	}
	out.documentCamps, err = rd.SelectAllDocumentCamps(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table document_camps: %s", err)
	}
	out.documentPersonnes, err = rd.SelectAllDocumentPersonnes(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table document_personnes: %s", err)
	}
	out.documentAides, err = rd.SelectAllDocumentAides(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table document_aides: %s", err)
	}
	out.lettredirecteurs, err = rd.SelectAllLettredirecteurs(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table lettredirecteurs: %s", err)
	}
	out.imageuploadeds, err = rd.SelectAllImageuploadeds(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table imageuploadeds: %s", err)
	}
	out.messages, err = rd.SelectAllMessages(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table messages: %s", err)
	}
	out.messageDocuments, err = rd.SelectAllMessageDocuments(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table message_documents: %s", err)
	}
	out.messageMessages, err = rd.SelectAllMessageMessages(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table message_messages: %s", err)
	}
	out.messagePlaceliberes, err = rd.SelectAllMessagePlaceliberes(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table message_placeliberes: %s", err)
	}
	out.messageAttestations, err = rd.SelectAllMessageAttestations(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table message_attestations: %s", err)
	}
	out.messageSondages, err = rd.SelectAllMessageSondages(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table message_sondages: %s", err)
	}
	out.messageViews, err = rd.SelectAllMessageViews(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table message_views: %s", err)
	}
	out.comptas, err = comptaV1.SelectAllComptas(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table comptas: %s", err)
	}
	out.records, err = comptaV1.SelectAllRecords(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table records: %s", err)
	}
	out.subrecords, err = comptaV1.SelectAllSubrecords(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table subrecords: %s", err)
	}
	out.filecomptas, err = comptaV1.SelectAllFilecomptas(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table filecomptas: %s", err)
	}
	out.budgets, err = comptaV1.SelectAllComptaBudgets(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table compta_budgets: %s", err)
	}
	out.budgetConfigs, err = comptaV1.SelectAllBudgetconfigs(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table budgetconfigs: %s", err)
	}
	out.busRecords, err = comptaV1.SelectAllRecordBuss(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table record_buss: %s", err)
	}
	return
}

func factureMomentsInscription(participants rd.Participants) map[int64]time.Time {
	out := make(map[int64]time.Time)
	for _, participant := range participants {
		if !participant.IdFacture.Valid {
			continue
		}
		date := participant.DateHeure.Time()
		if date.IsZero() {
			continue
		}
		if current, ok := out[participant.IdFacture.Int64]; !ok || date.Before(current) {
			out[participant.IdFacture.Int64] = date
		}
	}
	return out
}

func dossierFromV1(v1 rd.Facture, idTaux ds.IdTaux, responsable pr.IdPersonne, moment time.Time) ds.Dossier {
	return ds.Dossier{
		IdTaux:            idTaux,
		IdResponsable:     responsable,
		CopiesMails:       pr.Mails(v1.CopiesMails),
		PartageAdressesOK: v1.PartageAdressesOK,
		MomentInscription: moment,
		LastLoadDocuments: v1.LastConnection,
		KeyV1:             string(v1.Key),
	}
}

func participantFromV1(v1 rd.Participant, camp cps.IdCamp, personne pr.IdPersonne, dossier ds.IdDossier, taux ds.IdTaux) (cps.Participant, error) {
	statut, err := statutParticipant(v1.ListeAttente)
	if err != nil {
		return cps.Participant{}, err
	}
	return cps.Participant{
		IdCamp:           camp,
		IdPersonne:       personne,
		IdDossier:        dossier,
		IdTaux:           taux,
		Statut:           statut,
		Remises:          remises(v1.Remises),
		QuotientFamilial: int(v1.QuotientFamilial),
		OptionPrix:       optionPrixParticipant(v1.OptionPrix),
	}, nil
}

func statutParticipant(v1 rd.ListeAttente) (cps.StatutParticipant, error) {
	switch v1.Statut {
	case rd.Inscrit:
		return cps.Inscrit, nil
	case rd.Attente:
		return cps.AttenteCampComplet, nil
	case rd.AttenteReponse:
		return cps.EnAttenteReponse, nil
	case rd.Refuse:
		return cps.Refuse, nil
	default:
		return 0, fmt.Errorf("unsupported waitlist status %d", v1.Statut)
	}
}

func remises(v1 rd.Remises) cps.Remises {
	return cps.Remises{
		Equipiers: int(v1.ReducEquipiers),
		Famille:   int(v1.ReducEnfants),
		Speciale:  ds.Montant{Cent: int(v1.ReducSpeciale.Centimes()), Currency: ds.Euros},
	}
}

func optionPrixParticipant(v1 rd.OptionPrixParticipant) cps.OptionPrixParticipant {
	days := make(cps.JoursPresence, len(v1.Jour))
	for i, day := range v1.Jour {
		days[i] = int32(day)
	}
	return cps.OptionPrixParticipant{Jour: days}
}

func fraisACVE(v1 comptaV1.Fraismdr) compta.FraisMDRACVE {
	out := compta.FraisMDRACVE{Bus: make([]compta.Bus, len(v1.Bus))}
	for i, b := range v1.Bus {
		out.Bus[i] = compta.Bus{
			Name:    b.Name,
			Kind:    compta.CarKind(b.Kind),
			KmStart: b.KmStart, KmEnd: b.KmEnd,
		}
	}
	return out
}

func connectV1() (*sql.DB, error) {
	v1_HOST := os.Getenv("v1_HOST")
	v1_USER := os.Getenv("v1_USER")
	v1_PASSWORD := os.Getenv("v1_PASSWORD")
	v1_NAME := os.Getenv("v1_NAME")
	port := 5432
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		v1_HOST, port, v1_USER, v1_PASSWORD, v1_NAME)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("connexion DB : %s", err)
	}
	return db, nil
}

func connectV2() (*sql.DB, error) {
	config, err := config.NewDB()
	if err != nil {
		return nil, err
	}
	return config.ConnectPostgres()
}
