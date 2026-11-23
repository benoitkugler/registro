package main

import (
	"sort"

	comptaV1 "registro/migrations/fromACVEv1/compta"
	rd "registro/migrations/fromACVEv1/rawdata"
)

// parameters from trimming

const (
	donsStartYear  = 2025
	campsStartYear = 2024
)

// many inscrits without dossier
var blackListCamps = map[int64]bool{
	245: true, 263: true, 265: true, 283: true, 312: true, 315: true,
}

// returns a trimmed, deep copy
func (b baseV1) trim() (out baseV1) {
	out.dons, out.donateurs = trimDons(b.dons, b.donateurs)
	out.organismes = trimOrganismes(out.dons, out.donateurs, b.organismes)

	out.camps = trimCamps(b.camps)

	out.equipiers = trimEquipiers(b.equipiers, out)
	out.participants, out.factures = trimParticipants(b.participants, b.factures, out)

	out.paiements = trimPaiements(b.paiements, out)
	out.aides = trimAides(b.aides, out)
	out.structureaides = trimStructureaides(b.structureaides, out)
	out.sondages = trimSondages(b.sondages, out)

	out.groupes = trimGroupes(b.groupes, out)
	out.groupeParticipants = trimGroupeLinks(b, out)

	out.contraintes, out.campContraintes, out.equipierContraintes,
		out.documentPersonnes = trimContraintes(b, out)

	out.personnes = trimPersonnes(b.personnes, out)

	out.documents, out.contenuDocuments, out.documentCamps,
		out.documentPersonnes, out.documentAides = trimDocuments(b, out)

	out.lettredirecteurs = trimLettredirecteurs(b.lettredirecteurs, out)
	out.imageuploadeds = trimImageuploadeds(b.imageuploadeds, out)

	out.messages, out.messageDocuments, out.messageMessages, out.messagePlaceliberes,
		out.messageAttestations, out.messageSondages, out.messageViews = trimMessages(b, out)

	out.comptas = trimComptas(b.comptas, out)
	out.records = trimRecords(b.records, out)
	out.subrecords, out.filecomptas = trimSubrecords(b.subrecords, b.filecomptas, out)
	out.budgets = trimBudgets(b.budgets, out)
	out.budgetConfigs = b.budgetConfigs
	out.busRecords = trimBusRecords(b.busRecords, out)

	return out
}

type idSet = rd.Set

func newIDSet(ids ...int64) idSet {
	out := make(idSet, len(ids))
	for _, id := range ids {
		out.Add(id)
	}
	return out
}

// corePersonIds returns the legacy person IDs referenced by the retained
// core entities.
func corePersonIds(b, out baseV1) idSet {
	used := newIDSet()
	for _, link := range b.documentPersonnes {
		used.Add(link.IdPersonne)
	}
	for _, link := range out.donateurs {
		if link.IdPersonne.Valid {
			used.Add(link.IdPersonne.Int64)
		}
	}
	for _, equipier := range out.equipiers {
		used.Add(equipier.IdPersonne)
	}
	for _, facture := range out.factures {
		used.Add(facture.IdPersonne)
	}
	for _, participant := range out.participants {
		used.Add(participant.IdPersonne)
	}
	return used
}

func trimPaiements(paiements rd.Paiements, out baseV1) rd.Paiements {
	trimmed := make(rd.Paiements)
	for id, paiement := range paiements {
		if _, ok := out.factures[paiement.IdFacture]; ok {
			trimmed[id] = paiement
		}
	}
	return trimmed
}

func trimAides(aides rd.Aides, out baseV1) rd.Aides {
	trimmed := make(rd.Aides)
	for id, aide := range aides {
		if _, ok := out.participants[aide.IdParticipant]; ok {
			trimmed[id] = aide
		}
	}
	return trimmed
}

func trimStructureaides(structures rd.Structureaides, out baseV1) rd.Structureaides {
	used := newIDSet()
	for _, aide := range out.aides {
		used.Add(aide.IdStructureaide)
	}
	trimmed := make(rd.Structureaides)
	for id, structure := range structures {
		if used.Has(id) {
			trimmed[id] = structure
		}
	}
	return trimmed
}

func trimSondages(sondages rd.Sondages, out baseV1) rd.Sondages {
	trimmed := make(rd.Sondages)
	for id, sondage := range sondages {
		_, hasCamp := out.camps[sondage.IdCamp]
		_, hasFacture := out.factures[sondage.IdFacture]
		if hasCamp && hasFacture {
			trimmed[id] = sondage
		}
	}
	return trimmed
}

func trimGroupes(groupes rd.Groupes, out baseV1) rd.Groupes {
	trimmed := make(rd.Groupes)
	for id, groupe := range groupes {
		if _, ok := out.camps[groupe.IdCamp]; ok {
			trimmed[id] = groupe
		}
	}
	return trimmed
}

func trimGroupeLinks(b baseV1, out baseV1) rd.GroupeParticipants {
	participants := make(rd.GroupeParticipants, 0, len(b.groupeParticipants))
	for _, link := range b.groupeParticipants {
		_, hasParticipant := out.participants[link.IdParticipant]
		_, hasGroupe := out.groupes[link.IdGroupe]
		_, hasCamp := out.camps[link.IdCamp]
		if hasParticipant && hasGroupe && hasCamp {
			participants = append(participants, link)
		}
	}

	// v1 groupe_equipiers and participant_equipiers have no v2 equivalent
	// and are not migrated.
	return participants
}

func warnObsoleteBuiltins(counts map[rd.BuiltinContrainte]int) {
	if len(counts) == 0 {
		return
	}
	kinds := make([]string, 0, len(counts))
	for kind := range counts {
		kinds = append(kinds, string(kind))
	}
	sort.Strings(kinds)
	for _, kind := range kinds {
		builtin := rd.BuiltinContrainte(kind)
		label := rd.BuiltinContrainteLabels[builtin]
		if label == "" {
			label = kind
		}
		warningf("dropping %d obsolete v1 builtin constraint(s) of kind %q (%s): no v2 equivalent", counts[builtin], kind, label)
	}
}

// trimContraintes selects the constraints actually used by retained entities.
// V1 builtins without a v2 equivalent are ignored; custom constraints and
// test_nautique are preserved as v2 custom demandes. V1 groupe_contraintes
// links are ignored as they have no v2 equivalent.
func trimContraintes(b baseV1, out baseV1) (rd.Contraintes, rd.CampContraintes, rd.EquipierContraintes, rd.DocumentPersonnes) {
	obsolete := make(map[rd.BuiltinContrainte]int)
	keep := func(id int64) bool {
		contrainte, ok := b.contraintes[id]
		return ok && !isObsoleteBuiltin(contrainte.Builtin)
	}
	for _, contrainte := range b.contraintes {
		if isObsoleteBuiltin(contrainte.Builtin) {
			obsolete[contrainte.Builtin]++
		}
	}
	warnObsoleteBuiltins(obsolete)

	used := newIDSet()
	persons := corePersonIds(b, out)

	for _, link := range b.campContraintes {
		if _, ok := out.camps[link.IdCamp]; ok && keep(link.IdContrainte) {
			used.Add(link.IdContrainte)
		}
	}
	for _, link := range b.equipierContraintes {
		if _, ok := out.equipiers[link.IdEquipier]; ok && keep(link.IdContrainte) {
			used.Add(link.IdContrainte)
		}
	}
	for _, link := range b.documentPersonnes {
		if persons.Has(link.IdPersonne) && keep(link.IdContrainte) {
			used.Add(link.IdContrainte)
		}
	}
	for id := range used {
		if contrainte, ok := b.contraintes[id]; ok && contrainte.IdPersonne.Valid {
			persons.Add(contrainte.IdPersonne.Int64)
		}
	}

	contraintes := make(rd.Contraintes)
	for id, contrainte := range b.contraintes {
		if used.Has(id) && keep(id) {
			contraintes[id] = contrainte
		}
	}

	campLinks := make(rd.CampContraintes, 0, len(b.campContraintes))
	for _, link := range b.campContraintes {
		_, hasCamp := out.camps[link.IdCamp]
		_, hasContrainte := contraintes[link.IdContrainte]
		if hasCamp && hasContrainte {
			campLinks = append(campLinks, link)
		}
	}

	equipierLinks := make(rd.EquipierContraintes, 0, len(b.equipierContraintes))
	for _, link := range b.equipierContraintes {
		_, hasEquipier := out.equipiers[link.IdEquipier]
		_, hasContrainte := contraintes[link.IdContrainte]
		if hasEquipier && hasContrainte {
			equipierLinks = append(equipierLinks, link)
		}
	}

	personneLinks := make(rd.DocumentPersonnes, 0, len(b.documentPersonnes))
	for _, link := range b.documentPersonnes {
		hasPersonne := persons.Has(link.IdPersonne)
		_, hasContrainte := contraintes[link.IdContrainte]
		if hasPersonne && hasContrainte {
			personneLinks = append(personneLinks, link)
		}
	}
	return contraintes, campLinks, equipierLinks, personneLinks
}

func trimDocuments(b baseV1, out baseV1) (rd.Documents, rd.ContenuDocuments, rd.DocumentCamps, rd.DocumentPersonnes, rd.DocumentAides) {
	used := newIDSet()

	camps := make(rd.DocumentCamps, 0, len(out.documentCamps))
	for _, link := range b.documentCamps {
		if _, ok := out.camps[link.IdCamp]; !ok {
			continue
		}
		camps = append(camps, link)
		used.Add(link.IdDocument)
	}

	personnes := make(rd.DocumentPersonnes, 0, len(out.documentPersonnes))
	for _, link := range out.documentPersonnes {
		_, hasPersonne := out.personnes[link.IdPersonne]
		_, hasContrainte := out.contraintes[link.IdContrainte]
		if !hasPersonne || !hasContrainte {
			continue
		}
		personnes = append(personnes, link)
		used.Add(link.IdDocument)
	}

	aides := make(rd.DocumentAides, 0, len(b.documentAides))
	for _, link := range b.documentAides {
		if _, ok := out.aides[link.IdAide]; !ok {
			continue
		}
		aides = append(aides, link)
		used.Add(link.IdDocument)
	}

	for _, contrainte := range out.contraintes {
		if contrainte.IdDocument.Valid {
			used.Add(contrainte.IdDocument.Int64)
		}
	}

	documents := make(rd.Documents)
	for id, document := range b.documents {
		if used.Has(id) {
			documents[id] = document
		}
	}

	contenus := make(rd.ContenuDocuments, 0, len(b.contenuDocuments))
	for _, contenu := range b.contenuDocuments {
		if _, ok := documents[contenu.IdDocument]; ok {
			contenus = append(contenus, contenu)
		}
	}
	return documents, contenus, camps, personnes, aides
}

func trimLettredirecteurs(lettres rd.Lettredirecteurs, out baseV1) rd.Lettredirecteurs {
	trimmed := make(rd.Lettredirecteurs, 0, len(lettres))
	for _, lettre := range lettres {
		if _, ok := out.camps[lettre.IdCamp]; ok {
			trimmed = append(trimmed, lettre)
		}
	}
	return trimmed
}

func trimImageuploadeds(images rd.Imageuploadeds, out baseV1) rd.Imageuploadeds {
	trimmed := make(rd.Imageuploadeds, 0, len(images))
	for _, image := range images {
		if _, ok := out.camps[image.IdCamp]; ok {
			trimmed = append(trimmed, image)
		}
	}
	return trimmed
}

func trimMessages(b baseV1, out baseV1) (rd.Messages, rd.MessageDocuments, rd.MessageMessages, rd.MessagePlaceliberes, rd.MessageAttestations, rd.MessageSondages, rd.MessageViews) {
	messages := make(rd.Messages)
	for id, message := range b.messages {
		// v1 MInscription messages are ignored : the inscription date is
		// already stored on the participant.
		if message.Kind == rd.MInscription {
			continue
		}
		if _, ok := out.factures[message.IdFacture]; ok {
			messages[id] = message
		}
	}

	documents := make(rd.MessageDocuments, 0, len(b.messageDocuments))
	for _, link := range b.messageDocuments {
		_, hasMessage := messages[link.IdMessage]
		_, hasCamp := out.camps[link.IdCamp]
		if hasMessage && hasCamp {
			documents = append(documents, link)
		}
	}

	libres := make(rd.MessageMessages, 0, len(b.messageMessages))
	for _, link := range b.messageMessages {
		if _, ok := messages[link.IdMessage]; ok {
			libres = append(libres, link)
		}
	}

	places := make(rd.MessagePlaceliberes, 0, len(b.messagePlaceliberes))
	for _, link := range b.messagePlaceliberes {
		_, hasMessage := messages[link.IdMessage]
		_, hasParticipant := out.participants[link.IdParticipant]
		if hasMessage && hasParticipant {
			places = append(places, link)
		}
	}

	attestations := make(rd.MessageAttestations, 0, len(b.messageAttestations))
	for _, link := range b.messageAttestations {
		if _, ok := messages[link.IdMessage]; ok {
			attestations = append(attestations, link)
		}
	}

	sondages := make(rd.MessageSondages, 0, len(b.messageSondages))
	for _, link := range b.messageSondages {
		_, hasMessage := messages[link.IdMessage]
		_, hasCamp := out.camps[link.IdCamp]
		if hasMessage && hasCamp {
			sondages = append(sondages, link)
		}
	}

	views := make(rd.MessageViews, 0, len(b.messageViews))
	for _, link := range b.messageViews {
		_, hasMessage := messages[link.IdMessage]
		_, hasCamp := out.camps[link.IdCamp]
		if hasMessage && hasCamp {
			views = append(views, link)
		}
	}
	return messages, documents, libres, places, attestations, sondages, views
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
		// v2 participants always belong to a dossier. Legacy participants
		// without a facture cannot be migrated faithfully.
		if !participant.IdFacture.Valid {
			continue
		}
		facture, ok := factures[participant.IdFacture.Int64]
		if !ok {
			continue
		}
		trimmedParticipants[id] = participant
		trimmedFactures[participant.IdFacture.Int64] = facture
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
	for _, contrainte := range trimmed.contraintes {
		if contrainte.IdPersonne.Valid {
			used.Add(contrainte.IdPersonne.Int64)
		}
	}
	for _, link := range trimmed.documentPersonnes {
		used.Add(link.IdPersonne)
	}

	out := make(rd.Personnes)
	for id := range used {
		p, ok := v1[id]
		if !ok {
			continue
		}
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
