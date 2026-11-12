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
	comptaV1 "registro/migrations/fromACVEv1/compta"
	rd "registro/migrations/fromACVEv1/rawdata"
	cps "registro/sql/camps"
	"registro/sql/compta"
	"registro/sql/dons"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	"registro/utils"
)

// parameters from trimming

const (
	donsStartYear  = 2025
	campsStartYear = 2024
)

// many inscrits without dossier
var blackListCamps = map[int64]bool{
	245: true, 263: true, 265: true, 283: true, 312: true, 315: true,
	318: true, // TODO: remove this camp one the Calista special case has been handled
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	commit := flag.Bool("commit", false, "actually write the v2 DB")
	flag.Parse()

	fmt.Println("Connecting to v1...")
	dbV1, err := connectV1()
	check(err)
	defer dbV1.Close()
	fmt.Println("Done.")

	fmt.Println("Loading v1 data...")
	originV1, err := loadV1(dbV1)
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

	if *commit {
		fmt.Println("Connecting to v2...")
		dbV2, err := connectV2()
		check(err)
		defer dbV2.Close()
		fmt.Println("Done.")

		fmt.Println("Inserting in v2...")
		err = insert(dbV2, originV1, trimmedV1)
		check(err)
		fmt.Println("Done.")
	}
}

func euros(v rd.Euros) ds.Montant { return ds.NewEuros(float64(v)) }

func md(v rd.ModePaiment) ds.ModePaiement {
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

	// TODO: events

	documents rd.Documents
	// documentsPersonnes

	comptas       comptaV1.Comptas
	records       comptaV1.Records
	subrecords    comptaV1.Subrecords
	filecomptas   comptaV1.Filecomptas
	budgets       comptaV1.ComptaBudgets
	budgetConfigs comptaV1.Budgetconfigs
	busRecords    comptaV1.RecordBuss
}

func loadV1(db *sql.DB) (out baseV1, err error) {
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
	out.documents, err = rd.SelectAllDocuments(db)
	if err != nil {
		return baseV1{}, fmt.Errorf("table documents: %s", err)
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

// returns a trimmed, deep copy
func (b baseV1) trim() (out baseV1) {
	out.dons, out.donateurs = trimDons(b.dons, b.donateurs)
	out.organismes = trimOrganismes(out.dons, out.donateurs, b.organismes)

	out.documents = b.documents // TODO: filter unused categories

	out.camps = trimCamps(b.camps)

	out.equipiers = trimEquipiers(b.equipiers, out)
	out.participants, out.factures = trimParticipants(b.participants, b.factures, out)

	out.personnes = trimPersonnes(b.personnes, out)

	out.comptas = trimComptas(b.comptas, out)
	out.records = trimRecords(b.records, out)
	out.subrecords, out.filecomptas = trimSubrecords(b.subrecords, b.filecomptas, out)
	out.budgets = trimBudgets(b.budgets, out)
	out.budgetConfigs = b.budgetConfigs
	out.busRecords = trimBusRecords(b.busRecords, out)

	return out
}

func trimCamps(v1 rd.Camps) rd.Camps {
	out := make(rd.Camps)

	for id, c := range v1 {
		if c.InscriptionSimple || c.DateDebut.Time().Year() < campsStartYear {
			continue
		}
		if blackListCamps[id] {
			continue
		}
		out[id] = c
	}

	return out
}

func trimEquipiers(equipiers rd.Equipiers, out baseV1) rd.Equipiers {
	trimmed := make(rd.Equipiers)
	for id, equipier := range equipiers {
		if _, ok := out.camps[equipier.IdCamp]; ok {
			trimmed[id] = equipier
		}
	}
	return trimmed
}

// Keep registrations belonging to camps retained by trimCamps, and only the
// factures needed to create dossiers for those participants.
func trimParticipants(participants rd.Participants, factures rd.Factures, out baseV1) (rd.Participants, rd.Factures) {
	trimmedParticipants := make(rd.Participants)
	trimmedFactures := make(rd.Factures)
	for id, participant := range participants {
		if _, ok := out.camps[participant.IdCamp]; !ok {
			continue
		}
		trimmedParticipants[id] = participant
		if participant.IdFacture.Valid {
			if facture, ok := factures[participant.IdFacture.Int64]; ok {
				trimmedFactures[participant.IdFacture.Int64] = facture
			}
		}
	}
	return trimmedParticipants, trimmedFactures
}

// only keep recents year >= [donFirstYear], also drop anonymous gifts
// properly trim link table
func trimDons(dons rd.Dons, donateurs rd.DonDonateurs) (rd.Dons, rd.DonDonateurs) {
	byDon := donateurs.ByIdDon()

	outDons := make(rd.Dons)
	for id, don := range dons {
		if year := don.DateReception.Time().Year(); !(donsStartYear <= year) {
			continue
		} else if _, hasDonateur := byDon[id]; !hasDonateur {
			continue
		}
		outDons[id] = don
	}

	var outLinks rd.DonDonateurs
	for _, link := range donateurs {
		if _, has := outDons[link.IdDon]; !has {
			continue
		}
		outLinks = append(outLinks, link)
	}

	return outDons, outLinks
}

// only keep organismes required in [dons]
func trimOrganismes(dons rd.Dons, donateurs rd.DonDonateurs, organismes rd.Organismes) rd.Organismes {
	byDon := donateurs.ByIdDon()
	usedOrganismes := rd.NewSet()
	for idDon := range dons {
		donateur := byDon[idDon]
		if donateur.IdOrganisme.Valid {
			usedOrganismes.Add(donateur.IdOrganisme.Int64)
		}
	}
	out := make(rd.Organismes)
	for id, v := range organismes {
		if !usedOrganismes.Has(id) {
			continue
		}
		out[id] = v
	}
	return out
}

// only keep personnes used in other tables
func trimPersonnes(v1 rd.Personnes, trimmed baseV1) rd.Personnes {
	used := rd.NewSet()
	for _, link := range trimmed.donateurs {
		if link.IdPersonne.Valid {
			used.Add(link.IdPersonne.Int64)
		}
	}
	for _, eq := range trimmed.equipiers {
		used.Add(eq.IdPersonne)
	}
	for _, facture := range trimmed.factures {
		used.Add(facture.IdPersonne)
	}
	for _, participant := range trimmed.participants {
		used.Add(participant.IdPersonne)
	}

	out := make(rd.Personnes)
	for id := range used {
		p := v1[id]
		// TODO: for now, accept tmp people, but that should be
		// checked before real migration
		// if p.IsTemporaire {
		// 	panic("temporary profile with ID: " + strconv.Itoa(int(id)) + " " + string(p.NomPrenom()))
		// }
		out[id] = p
	}
	return out
}

// use the last equipier; returns a map keyed by IdPersonne
func (b baseV1) inferDiplomeAppro() map[int64]rd.Equipier {
	lastEquipier := map[int64]rd.Equipier{}
	for _, equipier := range b.equipiers {
		date := b.camps[equipier.IdCamp].DateDebut.Time()
		idPersonne := equipier.IdPersonne
		last, ok := lastEquipier[idPersonne]
		lastDate := b.camps[last.IdCamp].DateDebut.Time()
		if !ok || lastDate.Before(date) {
			lastEquipier[idPersonne] = equipier
		}
	}
	return lastEquipier
}

// remove unused camps
func trimComptas(in comptaV1.Comptas, out baseV1) comptaV1.Comptas {
	comptas := make(comptaV1.Comptas)
	for i, c := range in {
		_, has := out.camps[c.IdCamp]
		if !has {
			continue
		}
		comptas[i] = c
	}
	return comptas
}

func trimBudgets(in comptaV1.ComptaBudgets, out baseV1) comptaV1.ComptaBudgets {
	budgets := make(comptaV1.ComptaBudgets, 0, len(in))
	for _, budget := range in {
		if _, has := out.comptas[budget.IdCompta]; has {
			budgets = append(budgets, budget)
		}
	}
	return budgets
}

func trimRecords(in comptaV1.Records, out baseV1) comptaV1.Records {
	records := make(comptaV1.Records)
	for i, r := range in {
		_, has := out.comptas[r.IdCompta]
		if !has {
			continue
		}
		records[i] = r
	}
	return records
}

func trimSubrecords(in1 comptaV1.Subrecords, in2 comptaV1.Filecomptas, out baseV1) (comptaV1.Subrecords, comptaV1.Filecomptas) {
	subrecords, files := make(comptaV1.Subrecords), make(comptaV1.Filecomptas)
	for i, r := range in1 {
		_, has := out.records[r.IdRecord]
		if !has {
			continue
		}
		subrecords[i] = r
		if fId := r.IdFilecompta; fId.Valid {
			files[fId.Id] = in2[fId.Id]
		}
	}
	return subrecords, files
}

func trimBusRecords(recordBuss comptaV1.RecordBuss, out baseV1) comptaV1.RecordBuss {
	var bus comptaV1.RecordBuss
	for _, link := range recordBuss {
		if _, has := out.subrecords[link.IdSubrecord]; has {
			bus = append(bus, link)
		}
	}
	return bus
}

// TODO: fiche sanitaire + fiche equipier

func insert(dbV2 *sql.DB, origin, trimmed baseV1) error {
	// diplomeAppros := origin.inferDiplomeAppro()

	oM := make(map[int64]dons.IdOrganisme)
	pM := make(map[int64]pr.IdPersonne)
	cM := make(map[int64]cps.Camp)
	dM := make(map[int64]ds.IdDossier)
	coM := make(map[comptaV1.IdCompta]compta.IdComptaCamp)
	recordsM := make(map[comptaV1.IdRecord]compta.IdComptaRecord)
	subRecordsM := make(map[comptaV1.IdSubrecord]compta.IdComptaSubrecord)
	comptafilesM := make(map[comptaV1.IdFilecompta]compta.IdComptaFile)

	return utils.InTx(dbV2, func(tx *sql.Tx) error {
		var err error

		// insert personnes
		fmt.Println("\tpersonnes...")
		for _, v1 := range trimmed.personnes {
			// equipier := diplomeAppros[v1.Id]
			v2 := pr.Personne{
				Identite: pr.Identite{
					Nom:           string(v1.Nom),
					Prenom:        string(v1.Prenom),
					Sexe:          sexe(v1.Sexe),
					DateNaissance: shared.NewDateFrom(v1.DateNaissance.Time()),
					// VilleNaissance:       string(v1.VilleNaissance),
					// DepartementNaissance: pr.Departement(v1.DepartementNaissance),
					Nationnalite: pr.Nationnalite{},
					Tels:         tels(v1.Tels),
					Mail:         string(v1.Mail),
					Adresse:      string(v1.Adresse),
					CodePostal:   string(v1.CodePostal),
					Ville:        string(v1.Ville),
					Pays:         pr.Pays(v1.Pays),
					// NomJeuneFille:        string(v1.NomJeuneFille),
					// Profession:           string(v1.Profession),
					// Etudiant:             bool(v1.Etudiant),
					// Fonctionnaire:        bool(v1.Fonctionnaire),
					// Diplome:              diplome(equipier.Diplome),
					// Approfondissement:    approfondissement(equipier.Appro),
				},
			}
			v2, err = v2.Insert(tx)
			if err != nil {
				return err
			}
			pM[v1.Id] = v2.Id
		}

		// insert organismes
		fmt.Println("\torganismes...")
		for _, v1 := range trimmed.organismes {
			var coord rd.Coordonnees
			if contact := v1.IdContactDon; contact.Valid {
				p := origin.personnes[contact.Int64]
				coord = p.Coordonnees()
			} else if v1.ContactPropre {
				coord = v1.Contact
			} else if contact := v1.IdContact; contact.Valid {
				p := origin.personnes[contact.Int64]
				coord = p.Coordonnees()
			}
			v2 := dons.Organisme{
				Nom:        string(v1.Nom),
				Mail:       string(coord.Mail),
				Adresse:    string(coord.Adresse),
				CodePostal: string(coord.CodePostal),
				Ville:      string(coord.Ville),
				Pays:       pr.Pays(coord.Pays),
			}
			v2, err = v2.Insert(tx)
			if err != nil {
				return err
			}
			oM[v1.Id] = v2.Id
		}

		// insert dons
		donateursByDon := origin.donateurs.ByIdDon()
		fmt.Println("\tdons...")
		for _, v1 := range trimmed.dons {
			details := string(v1.Details)
			if v1.ModePaiement == rd.MPHelloasso {
				details += v1.Infos.IdPaiementHelloAsso
			}
			v2 := dons.Don{
				Montant:      euros(v1.Valeur),
				ModePaiement: md(v1.ModePaiement),
				Date:         shared.NewDateFrom(v1.DateReception.Time()),
				Affectation:  string(v1.Affectation),
				Details:      details,
				Remercie:     bool(v1.Remercie),
			}
			donateur := donateursByDon[v1.Id]
			if idV1 := donateur.IdPersonne; idV1.Valid {
				v2.IdPersonne = pM[idV1.Int64].Opt()
			} else if idV1 := donateur.IdOrganisme; idV1.Valid {
				v2.IdOrganisme = oM[idV1.Int64].Opt()
			}
			v2, err = v2.Insert(tx)
			if err != nil {
				return err
			}
		}

		// insert camps
		fmt.Println("\tcamps...")
		for _, camp := range trimmed.camps {
			v2 := cps.Camp{
				IdTaux:                 1, // the default taux, already created
				Nom:                    string(camp.Nom),
				DateDebut:              shared.NewDateFrom(camp.DateDebut.Time()),
				Duree:                  camp.Duree(),
				Lieu:                   string(camp.Lieu),
				Agrement:               string(camp.NumeroJS),
				Description:            string(camp.Infos),
				Navette:                navette(camp.Options.Bus),
				Places:                 int(camp.NbPlaces),
				AgeMin:                 int(camp.AgeMin),
				AgeMax:                 int(camp.AgeMax),
				NeedEquilibreGF:        bool(camp.NeedEquilibreGf),
				Statut:                 statutCamp(camp.Ouvert),
				Prix:                   cps.Montant{Cent: int(camp.Prix.Centimes()), Currency: ds.Euros},
				OptionPrix:             cps.OptionPrixCamp{},       // TODO
				OptionQuotientFamilial: cps.PrixQuotientFamilial{}, // TODO
				Password:               string(camp.Password),
				DocumentsReady:         !camp.Envois.Locked,
				DocumentsToShow:        cps.DocumentsToShow{}, // TODO
				Vetements:              cps.ListeVetements{},  // TODO
				AlbumID:                "",                    // TODO: migration from Joomeo
				IsAlbumVisible:         true,                  // TODO
				PlacesReservees:        int(camp.NbPlacesReservees),
			}
			v2, err = v2.Insert(tx)
			if err != nil {
				return err
			}
			cM[camp.Id] = v2
		}

		// insert equipiers
		fmt.Println("\tequipiers...")
		for _, eq := range trimmed.equipiers {
			camp := cM[eq.IdCamp]
			v2 := cps.Equipier{
				IdCamp:        camp.Id,
				IdPersonne:    pM[eq.IdPersonne],
				Roles:         roles(eq.Roles),
				Presence:      presenceOffsets(eq.Presence, camp.Plage()),
				FormStatus:    formStatusEquipier(eq.InvitationEquipier),
				AccepteCharte: nullBool(eq.Charte),
			}
			v2, err = v2.Insert(tx)
			if err != nil {
				return err
			}
		}

		// insert dossiers, one per retained legacy facture
		fmt.Println("\tdossiers...")
		momentsInscription := factureMomentsInscription(origin.participants)
		for id, facture := range trimmed.factures {
			if _, ok := origin.personnes[facture.IdPersonne]; !ok {
				return fmt.Errorf("facture %d: missing v1 responsible person %d", id, facture.IdPersonne)
			}
			idResponsable, ok := pM[facture.IdPersonne]
			if !ok {
				return fmt.Errorf("facture %d: missing migrated responsible person %d", id, facture.IdPersonne)
			}
			v2 := dossierFromV1(facture, 1, idResponsable, momentsInscription[id])
			v2, err = v2.Insert(tx)
			if err != nil {
				return err
			}
			dM[id] = v2.Id
		}

		// insert participants
		fmt.Println("\tparticipants...")
		for id, participant := range trimmed.participants {
			if !participant.IdFacture.Valid {
				return fmt.Errorf("participant %d: missing facture", id)
			}
			idDossier, ok := dM[participant.IdFacture.Int64]
			if !ok {
				return fmt.Errorf("participant %d: missing migrated facture %d", id, participant.IdFacture.Int64)
			}
			camp, ok := cM[participant.IdCamp]
			if !ok {
				return fmt.Errorf("participant %d: missing migrated camp %d", id, participant.IdCamp)
			}
			if _, ok := origin.personnes[participant.IdPersonne]; !ok {
				return fmt.Errorf("participant %d: missing v1 person %d", id, participant.IdPersonne)
			}
			idPersonne, ok := pM[participant.IdPersonne]
			if !ok {
				return fmt.Errorf("participant %d: missing migrated person %d", id, participant.IdPersonne)
			}
			v2, err := participantFromV1(participant, camp.Id, idPersonne, idDossier, camp.IdTaux)
			if err != nil {
				return fmt.Errorf("participant %d: %w", id, err)
			}
			if _, err = v2.Insert(tx); err != nil {
				return err
			}
		}

		// insert comptas
		fmt.Println("\tcomptas...")
		for _, c := range trimmed.comptas {
			v2 := compta.ComptaCamp{
				IdCamp:                cM[c.IdCamp].Id,
				FileStatus:            compta.FileStatus(c.FileStatus),
				State:                 compta.State(c.State),
				LastSentMailAdmin:     c.LastReviewNotif,
				LastSentMailDirecteur: c.LastReviewNotif,
				Account:               compta.AccountACVE(c.Account),
				FraisMDR:              fraisACVE(c.Fraismdr),
			}
			v2, err = v2.Insert(tx)
			if err != nil {
				return err
			}
			coM[c.Id] = v2.Id
		}

		// insert record
		fmt.Println("\trecords...")
		for _, r := range trimmed.records {
			account := compta.AccountCamp
			if r.IsOtherAccount {
				account = compta.AccountAsso
			}
			v2 := compta.ComptaRecord{
				IdCamp:      coM[r.IdCompta],
				IsReceipe:   r.IsReceipe,
				SpentBy:     compta.OptIdEquipier{},
				Reviews:     reviews(r.Reviews),
				IsValidated: r.IsValidated,
				Account:     account,
			}
			v2, err = v2.Insert(tx)
			if err != nil {
				return err
			}
			recordsM[r.Id] = v2.Id
		}

		// insert filecomptas, required for subrecords
		fmt.Println("\tfilecomptas...")
		// TODO: properly handle file content
		// TODO: probably switch to drive mode
		for _, f := range trimmed.filecomptas {
			v2 := compta.ComptaFile{Name: f.Name, DriveID: f.DriveID, DriveIDMiniature: f.DriveIDMiniature}
			v2, err = v2.Insert(tx)
			if err != nil {
				return err
			}
			comptafilesM[f.Id] = v2.Id
		}
		// insert subrecord
		fmt.Println("\tsubrecords...")
		for _, r := range trimmed.subrecords {
			idFile := compta.OptIdComptaFile{}
			if r.IdFilecompta.Valid {
				idFile = compta.NewOptIdComptaFile(comptafilesM[r.IdFilecompta.Id])
			}
			v2 := compta.ComptaSubrecord{
				IdRecord:    recordsM[r.IdRecord],
				Date:        shared.NewDateFrom(r.Date.Time()),
				Description: r.Title,
				Mode:        compta.ModePaiementRecord(r.Mode),
				Serial:      r.Serial,
				Currency:    ds.Euros,
				Montants:    montants(r.Montants),
				IdFile:      idFile,
			}
			v2, err = v2.Insert(tx)
			if err != nil {
				return err
			}
			subRecordsM[r.Id] = v2.Id
		}

		// insert record_bus
		fmt.Println("\trecord_buss...")
		for _, link := range trimmed.busRecords {
			v2 := compta.ComptaRecordBus{
				IdCamp:      coM[link.IdCompta],
				IdSubrecord: subRecordsM[link.IdSubrecord],
			}
			err = v2.Insert(tx)
			if err != nil {
				return err
			}
		}

		// insert budget configs
		fmt.Println("\tbudget configs...")
		for _, c := range trimmed.budgetConfigs {
			if c.Year == 2018 { // we already have one default config inserted
				continue
			}
			_, err = budgetConfigFromV1(c).Insert(tx)
			if err != nil {
				return err
			}
		}

		// insert budgets
		fmt.Println("\tbudgets...")
		for _, b := range trimmed.budgets {
			v2 := budgetFromV1(b, coM[b.IdCompta])
			err = v2.Insert(tx)
			if err != nil {
				return err
			}
		}

		return nil
	})
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
