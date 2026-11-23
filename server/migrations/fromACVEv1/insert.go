package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"registro/controllers/directeurs"
	"registro/crypto"
	comptaV1 "registro/migrations/fromACVEv1/compta"
	rd "registro/migrations/fromACVEv1/rawdata"
	cps "registro/sql/camps"
	"registro/sql/compta"
	"registro/sql/dons"
	ds "registro/sql/dossiers"
	ev "registro/sql/events"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	"registro/utils"
)

// migrationOptions controls the optional migration of legacy document bytes.
// Compta file bytes are never migrated: only their metadata is imported.
type migrationOptions struct {
	v1FilesDir           string
	v2FilesDir           string
	host                 string // public host used in migrated lettre image URLs
	skipDocumentContents bool
	encrypter            crypto.Encrypter
}

type pendingFileCopy struct {
	id        fs.IdFile
	content   []byte
	miniature []byte
}

type idMaps struct {
	organismes    map[int64]dons.IdOrganisme
	personnes     map[int64]pr.IdPersonne
	tempPersonnes map[int64]bool
	camps         map[int64]cps.Camp
	equipiers     map[int64]cps.IdEquipier
	dossiers      map[int64]ds.IdDossier
	participants  map[int64]cps.IdParticipant
	structures    map[int64]cps.IdStructureaide
	aides         map[int64]cps.IdAide
	groupes       map[int64]cps.IdGroupe
	sondages      map[int64]cps.IdSondage
	files         map[int64]fs.File
	demandes      map[int64]fs.IdDemande
	events        map[int64]ev.IdEvent

	comptas     map[comptaV1.IdCompta]compta.IdComptaCamp
	records     map[comptaV1.IdRecord]compta.IdComptaRecord
	subrecords  map[comptaV1.IdSubrecord]compta.IdComptaSubrecord
	comptafiles map[comptaV1.IdFilecompta]compta.IdComptaFile
}

func newIDMaps() idMaps {
	return idMaps{
		organismes:    make(map[int64]dons.IdOrganisme),
		personnes:     make(map[int64]pr.IdPersonne),
		tempPersonnes: make(map[int64]bool),
		camps:         make(map[int64]cps.Camp),
		equipiers:     make(map[int64]cps.IdEquipier),
		dossiers:      make(map[int64]ds.IdDossier),
		participants:  make(map[int64]cps.IdParticipant),
		structures:    make(map[int64]cps.IdStructureaide),
		aides:         make(map[int64]cps.IdAide),
		groupes:       make(map[int64]cps.IdGroupe),
		sondages:      make(map[int64]cps.IdSondage),
		files:         make(map[int64]fs.File),
		demandes:      make(map[int64]fs.IdDemande),
		events:        make(map[int64]ev.IdEvent),
		comptas:       make(map[comptaV1.IdCompta]compta.IdComptaCamp),
		records:       make(map[comptaV1.IdRecord]compta.IdComptaRecord),
		subrecords:    make(map[comptaV1.IdSubrecord]compta.IdComptaSubrecord),
		comptafiles:   make(map[comptaV1.IdFilecompta]compta.IdComptaFile),
	}
}

func checkMigrationOptions(opts migrationOptions) error {
	if opts.v1FilesDir == "" {
		return nil
	}
	if opts.v2FilesDir == "" {
		return fmt.Errorf("-v2-files-dir is required when -v1-files-dir is set")
	}
	if info, err := os.Stat(opts.v1FilesDir); err != nil {
		return fmt.Errorf("v1 files directory %s: %w", opts.v1FilesDir, err)
	} else if !info.IsDir() {
		return fmt.Errorf("v1 files directory %s is not a directory", opts.v1FilesDir)
	}
	info, err := os.Stat(opts.v2FilesDir)
	if err != nil {
		return fmt.Errorf("v2 files directory %s: %w", opts.v2FilesDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("v2 files directory %s is not a directory", opts.v2FilesDir)
	}
	probe, err := os.CreateTemp(opts.v2FilesDir, ".registro-migration-*")
	if err != nil {
		return fmt.Errorf("v2 files directory %s is not writable: %w", opts.v2FilesDir, err)
	}
	probeName := probe.Name()
	if err := probe.Close(); err != nil {
		_ = os.Remove(probeName)
		return err
	}
	return os.Remove(probeName)
}

// validateMigration performs the checks that can be made without a v2 DB.
// A commit performs additional structural and FK checks inside the transaction.
func validateMigration(trimmed baseV1, opts migrationOptions) error {
	if err := checkMigrationOptions(opts); err != nil {
		return err
	}
	for id, participant := range trimmed.participants {
		if !participant.IdFacture.Valid {
			return fmt.Errorf("participant %d has no facture", id)
		}
	}
	if len(trimmed.imageuploadeds) > 0 && opts.host == "" {
		return fmt.Errorf("-host is required when migrating lettre images")
	}
	return nil
}

func insert(dbV2 *sql.DB, origin, trimmed baseV1, opts migrationOptions) error {
	if err := validateMigration(trimmed, opts); err != nil {
		return err
	}
	var pending []pendingFileCopy
	err := utils.InTx(dbV2, func(tx *sql.Tx) error {
		var err error
		pending, err = migrate(tx, origin, trimmed, opts)
		return err
	})
	if err != nil {
		return err
	}
	// File contents are written only after the SQL transaction has committed:
	// filesystem writes cannot be rolled back.
	return applyPendingFiles(opts, pending)
}

func migrate(tx *sql.Tx, origin, trimmed baseV1, opts migrationOptions) ([]pendingFileCopy, error) {
	m := newIDMaps()
	lastEquipiers := trimmed.inferDiplomeAppro()

	if err := migratePersonnes(tx, trimmed, m, lastEquipiers); err != nil {
		return nil, err
	}
	if err := migrateOrganismes(tx, origin, trimmed, m); err != nil {
		return nil, err
	}
	if err := migrateDons(tx, origin, trimmed, m); err != nil {
		return nil, err
	}
	if err := migrateCamps(tx, trimmed, m); err != nil {
		return nil, err
	}
	if err := migrateEquipiers(tx, trimmed, m); err != nil {
		return nil, err
	}
	if err := migrateDossiers(tx, trimmed, m); err != nil {
		return nil, err
	}
	if err := migrateParticipants(tx, origin, trimmed, m); err != nil {
		return nil, err
	}
	if err := migratePaiements(tx, trimmed, m); err != nil {
		return nil, err
	}
	if err := migrateStructuresAndAides(tx, trimmed, m); err != nil {
		return nil, err
	}
	if err := migrateGroupes(tx, trimmed, m); err != nil {
		return nil, err
	}
	if err := migrateSondages(tx, trimmed, m); err != nil {
		return nil, err
	}
	if err := migrateLettresImages(tx, trimmed, m, opts); err != nil {
		return nil, err
	}
	pending, err := migrateFilesDemandes(tx, trimmed, m, opts)
	if err != nil {
		return nil, err
	}
	if err := migrateEvents(tx, trimmed, m); err != nil {
		return nil, err
	}
	if err := migrateCompta(tx, trimmed, m); err != nil {
		return nil, err
	}

	fmt.Println("\tintentional omissions: v1 doncs, users, groupe_equipiers, participant_equipiers, groupe_contraintes (no v2 equivalent)")
	return pending, nil
}

func migratePersonnes(tx *sql.Tx, trimmed baseV1, m idMaps, lastEquipiers map[int64]rd.Equipier) error {
	fmt.Println("\tpersonnes...")
	for _, v1 := range trimmed.personnes {
		v2, err := personneFromV1(v1).Insert(tx)
		if err != nil {
			return err
		}
		m.personnes[v1.Id] = v2.Id
		if v2.IsTemp {
			m.tempPersonnes[v1.Id] = true
		}

		if fiche, ok := ficheSanitaireFromV1(v1); ok && !v2.IsTemp {
			fiche.IdPersonne = v2.Id
			if err = fiche.Insert(tx); err != nil {
				return fmt.Errorf("fiche sanitaire %d: %w", v1.Id, err)
			}
		}
		if equipier, ok := lastEquipiers[v1.Id]; ok && !v2.IsTemp {
			fiche := ficheEquipierFromV1(v1, equipier)
			fiche.IdPersonne = v2.Id
			if err = fiche.Insert(tx); err != nil {
				return fmt.Errorf("fiche equipier %d: %w", v1.Id, err)
			}
		}
	}
	return nil
}

func migrateOrganismes(tx *sql.Tx, origin, trimmed baseV1, m idMaps) error {
	fmt.Println("\torganismes...")
	for _, v1 := range trimmed.organismes {
		var coord rd.Coordonnees
		if contact := v1.IdContactDon; contact.Valid {
			coord = origin.personnes[contact.Int64].Coordonnees()
		} else if v1.ContactPropre {
			coord = v1.Contact
		} else if contact := v1.IdContact; contact.Valid {
			coord = origin.personnes[contact.Int64].Coordonnees()
		}
		v2, err := dons.Organisme{
			Nom:        string(v1.Nom),
			Mail:       string(coord.Mail),
			Adresse:    string(coord.Adresse),
			CodePostal: string(coord.CodePostal),
			Ville:      string(coord.Ville),
			Pays:       pr.Pays(coord.Pays),
		}.Insert(tx)
		if err != nil {
			return err
		}
		m.organismes[v1.Id] = v2.Id
	}
	return nil
}

func migrateDons(tx *sql.Tx, origin, trimmed baseV1, m idMaps) error {
	donateursByDon := origin.donateurs.ByIdDon()
	fmt.Println("\tdons...")
	for _, v1 := range trimmed.dons {
		details := string(v1.Details)
		if v1.ModePaiement == rd.MPHelloasso {
			details += v1.Infos.IdPaiementHelloAsso
		}
		v2 := dons.Don{
			Montant:      euros(v1.Valeur),
			ModePaiement: modePaiement(v1.ModePaiement),
			Date:         shared.NewDateFrom(v1.DateReception.Time()),
			Affectation:  string(v1.Affectation),
			Details:      details,
			Remercie:     bool(v1.Remercie),
		}
		donateur := donateursByDon[v1.Id]
		if idV1 := donateur.IdPersonne; idV1.Valid {
			v2.IdPersonne = m.personnes[idV1.Int64].Opt()
		} else if idV1 := donateur.IdOrganisme; idV1.Valid {
			v2.IdOrganisme = m.organismes[idV1.Int64].Opt()
		}
		if _, err := v2.Insert(tx); err != nil {
			return err
		}
	}
	return nil
}

func migrateCamps(tx *sql.Tx, trimmed baseV1, m idMaps) error {
	fmt.Println("\tcamps...")
	for _, v1 := range trimmed.camps {
		v2, err := campFromV1(v1).Insert(tx)
		if err != nil {
			return err
		}
		m.camps[v1.Id] = v2
	}
	return nil
}

func migrateEquipiers(tx *sql.Tx, trimmed baseV1, m idMaps) error {
	fmt.Println("\tequipiers...")
	for _, v1 := range trimmed.equipiers {
		camp, ok := m.camps[v1.IdCamp]
		if !ok {
			return fmt.Errorf("equipier %d: missing migrated camp %d", v1.Id, v1.IdCamp)
		}
		v2, err := cps.Equipier{
			IdCamp:        camp.Id,
			IdPersonne:    m.personnes[v1.IdPersonne],
			Roles:         roles(v1.Roles),
			Presence:      presenceOffsets(v1.Presence, camp.Plage()),
			FormStatus:    formStatusEquipier(v1.InvitationEquipier),
			AccepteCharte: nullBool(v1.Charte),
		}.Insert(tx)
		if err != nil {
			return err
		}
		m.equipiers[v1.Id] = v2.Id
	}
	return nil
}

func migrateDossiers(tx *sql.Tx, trimmed baseV1, m idMaps) error {
	fmt.Println("\tdossiers...")
	moments := factureMomentsInscription(trimmed.participants)
	for id, facture := range trimmed.factures {
		idResponsable, ok := m.personnes[facture.IdPersonne]
		if !ok {
			return fmt.Errorf("facture %d: missing migrated responsible person %d", id, facture.IdPersonne)
		}
		v2, err := dossierFromV1(facture, 1, idResponsable, moments[id]).Insert(tx)
		if err != nil {
			return err
		}
		m.dossiers[id] = v2.Id
	}
	return nil
}

func migrateParticipants(tx *sql.Tx, origin, trimmed baseV1, m idMaps) error {
	fmt.Println("\tparticipants...")
	for id, v1 := range trimmed.participants {
		if !v1.IdFacture.Valid {
			return fmt.Errorf("participant %d: missing facture", id)
		}
		idDossier, ok := m.dossiers[v1.IdFacture.Int64]
		if !ok {
			return fmt.Errorf("participant %d: missing migrated facture %d", id, v1.IdFacture.Int64)
		}
		camp, ok := m.camps[v1.IdCamp]
		if !ok {
			return fmt.Errorf("participant %d: missing migrated camp %d", id, v1.IdCamp)
		}
		idPersonne, ok := m.personnes[v1.IdPersonne]
		if !ok {
			return fmt.Errorf("participant %d: missing migrated person %d", id, v1.IdPersonne)
		}
		v2, err := participantFromV1(v1, camp.Id, idPersonne, idDossier, camp.IdTaux)
		if err != nil {
			return fmt.Errorf("participant %d: %w", id, err)
		}
		inserted, err := v2.Insert(tx)
		if err != nil {
			return err
		}
		m.participants[id] = inserted.Id
	}
	return nil
}

func migratePaiements(tx *sql.Tx, trimmed baseV1, m idMaps) error {
	fmt.Println("\tpaiements...")
	for _, v1 := range trimmed.paiements {
		idDossier, ok := m.dossiers[v1.IdFacture]
		if !ok {
			return fmt.Errorf("paiement %d: missing migrated facture %d", v1.Id, v1.IdFacture)
		}
		v2 := paiementFromV1(v1)
		v2.IdDossier = idDossier
		if _, err := v2.Insert(tx); err != nil {
			return err
		}
	}
	return nil
}

func migrateStructuresAndAides(tx *sql.Tx, trimmed baseV1, m idMaps) error {
	fmt.Println("\tstructures d'aides...")
	for id, v1 := range trimmed.structureaides {
		v2, err := structureAideFromV1(v1).Insert(tx)
		if err != nil {
			return err
		}
		m.structures[id] = v2.Id
	}

	fmt.Println("\taides...")
	for id, v1 := range trimmed.aides {
		idStructure, ok := m.structures[v1.IdStructureaide]
		if !ok {
			return fmt.Errorf("aide %d: missing migrated structure %d", id, v1.IdStructureaide)
		}
		idParticipant, ok := m.participants[v1.IdParticipant]
		if !ok {
			return fmt.Errorf("aide %d: missing migrated participant %d", id, v1.IdParticipant)
		}
		v2 := aideFromV1(v1)
		v2.IdStructureaide = idStructure
		v2.IdParticipant = idParticipant
		inserted, err := v2.Insert(tx)
		if err != nil {
			return err
		}
		m.aides[id] = inserted.Id
	}
	return nil
}

func migrateGroupes(tx *sql.Tx, trimmed baseV1, m idMaps) error {
	fmt.Println("\tgroupes...")
	for id, v1 := range trimmed.groupes {
		camp, ok := m.camps[v1.IdCamp]
		if !ok {
			return fmt.Errorf("groupe %d: missing migrated camp %d", id, v1.IdCamp)
		}
		v2 := groupeFromV1(v1)
		v2.IdCamp = camp.Id
		inserted, err := v2.Insert(tx)
		if err != nil {
			return err
		}
		m.groupes[id] = inserted.Id
	}

	fmt.Println("\tgroupe_participants...")
	for _, link := range trimmed.groupeParticipants {
		camp, ok := m.camps[link.IdCamp]
		if !ok {
			continue
		}
		v2 := cps.GroupeParticipant{
			IdParticipant: m.participants[link.IdParticipant],
			IdGroupe:      m.groupes[link.IdGroupe],
			IdCamp:        camp.Id,
			Manuel:        link.Manuel,
		}
		if err := v2.Insert(tx); err != nil {
			return err
		}
	}
	return nil
}

func migrateSondages(tx *sql.Tx, trimmed baseV1, m idMaps) error {
	fmt.Println("\tsondages...")
	for id, v1 := range trimmed.sondages {
		camp, ok := m.camps[v1.IdCamp]
		if !ok {
			return fmt.Errorf("sondage %d: missing migrated camp %d", id, v1.IdCamp)
		}
		idDossier, ok := m.dossiers[v1.IdFacture]
		if !ok {
			return fmt.Errorf("sondage %d: missing migrated facture %d", id, v1.IdFacture)
		}
		v2 := sondageFromV1(v1)
		v2.IdCamp = camp.Id
		v2.IdDossier = idDossier
		inserted, err := v2.Insert(tx)
		if err != nil {
			return err
		}
		m.sondages[id] = inserted.Id
	}
	return nil
}

// v1LettreImageURLPrefix is the prefix of image sources stored in v1 lettres.
const v1LettreImageURLPrefix = "https://acve.fr/api/imageslettre?lien="

// lettreHTMLFromV1 replaces the v1 image links (identified by their Lien)
// with the dynamic v2 URLs. The Lien is only unique within a camp, so the
// mapping is scoped by camp by the caller.
func lettreHTMLFromV1(html string, lienURLs map[string]string) string {
	if len(lienURLs) == 0 {
		return html
	}
	replacements := make([]string, 0, len(lienURLs)*2)
	for lien, url := range lienURLs {
		replacements = append(replacements, v1LettreImageURLPrefix+lien, url)
	}
	return strings.NewReplacer(replacements...).Replace(html)
}

func migrateLettresImages(tx *sql.Tx, trimmed baseV1, m idMaps, opts migrationOptions) error {
	fmt.Println("\timageuploadeds...")
	lienURLs := make(map[int64]map[string]string)
	for _, v1 := range trimmed.imageuploadeds {
		inserted, err := (cps.LettreImage{Filename: v1.Filename, Content: v1.Content}).Insert(tx)
		if err != nil {
			return err
		}
		key := crypto.EncryptID(opts.encrypter, inserted.Id)
		url := utils.BuildUrl(opts.host, directeurs.EndpointLettreImages, utils.QP("key", key))
		if lienURLs[v1.IdCamp] == nil {
			lienURLs[v1.IdCamp] = make(map[string]string)
		}
		lienURLs[v1.IdCamp][v1.Lien] = url
	}

	fmt.Println("\tlettredirecteurs...")
	for _, v1 := range trimmed.lettredirecteurs {
		camp, ok := m.camps[v1.IdCamp]
		if !ok {
			continue
		}
		v2 := cps.Lettredirecteur{
			IdCamp:             camp.Id,
			Html:               lettreHTMLFromV1(v1.Html, lienURLs[v1.IdCamp]),
			UseCoordCentre:     v1.UseCoordCentre,
			ShowAdressePostale: v1.ShowAdressePostale,
			ColorCoord:         v1.ColorCoord,
		}
		if err := v2.Insert(tx); err != nil {
			return err
		}
	}
	return nil
}

func optFileFromV1(v1 rd.OptionnalId, m map[int64]fs.File) fs.OptIdFile {
	if !v1.Valid {
		return fs.OptIdFile{}
	}
	file, ok := m[v1.Int64]
	if !ok {
		return fs.OptIdFile{}
	}
	return fs.OptIdFile{Id: file.Id, Valid: true}
}

func optPersonneFromV1(v1 rd.OptionnalId, m map[int64]pr.IdPersonne) pr.OptIdPersonne {
	if !v1.Valid {
		return pr.OptIdPersonne{}
	}
	id, ok := m[v1.Int64]
	if !ok {
		return pr.OptIdPersonne{}
	}
	return pr.OptIdPersonne{Id: id, Valid: true}
}

func legacyDocumentPath(dir string, idV1 int64) string {
	return filepath.Join(dir, fmt.Sprintf("file_%d", idV1))
}

// stageLegacyDocument reads the document content from the legacy files
// directory. The miniature comes from the v1 contenu_documents table and is
// passed in by the caller: there is no fallback to legacy *_min files.
func stageLegacyDocument(opts migrationOptions, idV1 int64, idV2 fs.IdFile, miniature []byte) ([]pendingFileCopy, error) {
	if opts.v1FilesDir == "" {
		return nil, nil
	}
	content, err := os.ReadFile(legacyDocumentPath(opts.v1FilesDir, idV1))
	if err != nil {
		return nil, nil
	}
	return []pendingFileCopy{{id: idV2, content: content, miniature: miniature}}, nil
}

func saveFileAtomic(dir string, id fs.IdFile, content []byte, miniature bool) error {
	name := fmt.Sprintf("file_%d", id)
	if miniature {
		name += "_min"
	}
	tmp, err := os.CreateTemp(dir, ".registro-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(content); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(os.ModePerm); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, filepath.Join(dir, name))
}

func applyPendingFiles(opts migrationOptions, pending []pendingFileCopy) error {
	if len(pending) == 0 {
		return nil
	}
	if err := checkMigrationOptions(opts); err != nil {
		return err
	}
	for _, copy := range pending {
		if err := saveFileAtomic(opts.v2FilesDir, copy.id, copy.content, false); err != nil {
			return fmt.Errorf("file %d: %w", copy.id, err)
		}
		if copy.miniature != nil {
			if err := saveFileAtomic(opts.v2FilesDir, copy.id, copy.miniature, true); err != nil {
				return fmt.Errorf("file %d miniature: %w", copy.id, err)
			}
		}
	}
	return nil
}

// builtinDemandeFromV1 reuses a predeclared v2 demande instead of inserting a
// duplicate. The global Demande.IdFile is deliberately left untouched: a v1
// template document cannot be attached to a shared builtin row.
func builtinDemandeFromV1(v1 rd.Contrainte, builtins fs.Builtins) (fs.Demande, bool) {
	categorie, _ := categorieFromV1(v1.Builtin)
	if categorie == fs.NoBuiltin {
		return fs.Demande{}, false
	}
	return builtins[categorie], true
}

type campDemandeKey struct {
	camp    cps.IdCamp
	demande fs.IdDemande
}

type equipierDemandeKey struct {
	equipier cps.IdEquipier
	demande  fs.IdDemande
}

func fileCampsFromV1(links rd.DocumentCamps, m idMaps) []fs.FileCamp {
	out := make([]fs.FileCamp, 0, len(links))
	seenFiles := make(map[fs.IdFile]bool)
	seenLetterCamps := make(map[cps.IdCamp]bool)
	// v2 allows at most one letter per camp, so prefer letter documents.
	for pass := 0; pass < 2; pass++ {
		for _, link := range links {
			if link.IsLettre != (pass == 0) {
				continue
			}
			camp, ok := m.camps[link.IdCamp]
			if !ok {
				continue
			}
			file, ok := m.files[link.IdDocument]
			if !ok || seenFiles[file.Id] {
				continue
			}
			if link.IsLettre && seenLetterCamps[camp.Id] {
				continue
			}
			out = append(out, fs.FileCamp{IdFile: file.Id, IdCamp: camp.Id, IsLettre: link.IsLettre})
			seenFiles[file.Id] = true
			if link.IsLettre {
				seenLetterCamps[camp.Id] = true
			}
		}
	}
	return out
}

func filePersonnesFromV1(links rd.DocumentPersonnes, m idMaps) []fs.FilePersonne {
	out := make([]fs.FilePersonne, 0, len(links))
	seenFiles := make(map[fs.IdFile]bool)
	for _, link := range links {
		file, ok := m.files[link.IdDocument]
		if !ok || seenFiles[file.Id] {
			continue
		}
		// The guard FK in v2 forbids linking documents to temporary profiles.
		if m.tempPersonnes[link.IdPersonne] {
			continue
		}
		idDemande, ok := m.demandes[link.IdContrainte]
		if !ok {
			continue
		}
		idPersonne, ok := m.personnes[link.IdPersonne]
		if !ok {
			continue
		}
		out = append(out, fs.FilePersonne{
			IdFile:     file.Id,
			IdPersonne: idPersonne,
			IdDemande:  idDemande,
		})
		seenFiles[file.Id] = true
	}
	return out
}

func fileAidesFromV1(links rd.DocumentAides, m idMaps) []fs.FileAide {
	out := make([]fs.FileAide, 0, len(links))
	seenFiles := make(map[fs.IdFile]bool)
	seenAides := make(map[cps.IdAide]bool)
	for _, link := range links {
		file, ok := m.files[link.IdDocument]
		if !ok || seenFiles[file.Id] {
			continue
		}
		idAide, ok := m.aides[link.IdAide]
		if !ok || seenAides[idAide] {
			continue
		}
		out = append(out, fs.FileAide{IdFile: file.Id, IdAide: idAide})
		seenFiles[file.Id] = true
		seenAides[idAide] = true
	}
	return out
}

func migrateFilesDemandes(tx *sql.Tx, trimmed baseV1, m idMaps, opts migrationOptions) ([]pendingFileCopy, error) {
	fmt.Println("\tfiles...")
	var pending []pendingFileCopy
	miniatures := trimmed.contenuDocuments.ByIdDocument()
	for id, v1 := range trimmed.documents {
		v2, err := fileFromV1(v1).Insert(tx)
		if err != nil {
			return nil, err
		}
		m.files[id] = v2
		copy, err := stageLegacyDocument(opts, id, v2.Id, miniatures[id].Miniature)
		if err != nil {
			return nil, err
		}
		pending = append(pending, copy...)
	}

	fmt.Println("\tdemandes...")
	builtins, err := fs.LoadBuiltins(tx)
	if err != nil {
		return nil, err
	}
	for id, v1 := range trimmed.contraintes {
		if builtin, ok := builtinDemandeFromV1(v1, builtins); ok {
			m.demandes[id] = builtin.Id
			continue
		}
		// Defensive: obsolete builtins are removed by trimContraintes and
		// warned about there. Never migrate one as a custom Demande.
		if isObsoleteBuiltin(v1.Builtin) {
			continue
		}
		_, legacyLabel := categorieFromV1(v1.Builtin)
		demande := demandeFromV1(
			v1, fs.NoBuiltin, legacyLabel,
			optFileFromV1(v1.IdDocument, m.files),
			optPersonneFromV1(v1.IdPersonne, m.personnes),
		)
		inserted, err := demande.Insert(tx)
		if err != nil {
			return nil, err
		}
		m.demandes[id] = inserted.Id
	}

	fmt.Println("\tdemande_camps / demande_equipiers...")
	seenCamps := make(map[campDemandeKey]bool)
	for _, link := range trimmed.campContraintes {
		camp, ok := m.camps[link.IdCamp]
		if !ok {
			continue
		}
		idDemande, ok := m.demandes[link.IdContrainte]
		if !ok {
			continue
		}
		key := campDemandeKey{camp: camp.Id, demande: idDemande}
		if seenCamps[key] {
			continue
		}
		v2 := fs.DemandeCamp{IdCamp: camp.Id, IdDemande: idDemande}
		if err := v2.Insert(tx); err != nil {
			return nil, err
		}
		seenCamps[key] = true
	}
	seenEquipiers := make(map[equipierDemandeKey]bool)
	for _, link := range trimmed.equipierContraintes {
		idDemande, ok := m.demandes[link.IdContrainte]
		if !ok {
			continue
		}
		idEquipier, ok := m.equipiers[link.IdEquipier]
		if !ok {
			continue
		}
		key := equipierDemandeKey{equipier: idEquipier, demande: idDemande}
		if seenEquipiers[key] {
			continue
		}
		v2 := fs.DemandeEquipier{
			IdEquipier:  idEquipier,
			IdDemande:   idDemande,
			Optionnelle: link.Optionnel,
		}
		if err := v2.Insert(tx); err != nil {
			return nil, err
		}
		seenEquipiers[key] = true
	}

	fmt.Println("\tfile_camps / file_personnes / file_aides...")
	for _, v2 := range fileCampsFromV1(trimmed.documentCamps, m) {
		if err := v2.Insert(tx); err != nil {
			return nil, err
		}
	}
	for _, v2 := range filePersonnesFromV1(trimmed.documentPersonnes, m) {
		if err := v2.Insert(tx); err != nil {
			return nil, err
		}
	}
	for _, v2 := range fileAidesFromV1(trimmed.documentAides, m) {
		if err := v2.Insert(tx); err != nil {
			return nil, err
		}
	}
	return pending, nil
}

func firstParticipantForFacture(participants rd.Participants, idFacture int64) (rd.Participant, bool) {
	for _, participant := range participants {
		if participant.IdFacture.Valid && participant.IdFacture.Int64 == idFacture {
			return participant, true
		}
	}
	return rd.Participant{}, false
}

// inscritParticipantsForFacture returns the participants of a facture that are
// still inscrit. When none is inscrit, callers fall back to an arbitrary
// participant so that the v1 validation is still migrated.
func inscritParticipantsForFacture(participants rd.Participants, idFacture int64) []rd.Participant {
	out := make([]rd.Participant, 0)
	for _, participant := range participants {
		if !participant.IdFacture.Valid || participant.IdFacture.Int64 != idFacture {
			continue
		}
		if participant.ListeAttente.IsInscrit() {
			out = append(out, participant)
		}
	}
	return out
}

func campIDs(m idMaps) map[int64]cps.IdCamp {
	out := make(map[int64]cps.IdCamp, len(m.camps))
	for id, camp := range m.camps {
		out[id] = camp.Id
	}
	return out
}

func migrateEvents(tx *sql.Tx, trimmed baseV1, m idMaps) error {
	fmt.Println("\tmessages -> events...")
	messagesByID := trimmed.messageMessages.ByIdMessage()
	documentsByID := trimmed.messageDocuments.ByIdMessage()
	placesByID := trimmed.messagePlaceliberes.ByIdMessage()
	attestationsByID := trimmed.messageAttestations.ByIdMessage()
	sondagesByID := trimmed.messageSondages.ByIdMessage()
	viewsByID := trimmed.messageViews.ByIdMessage()
	campM := campIDs(m)

	for _, v1 := range trimmed.messages {
		idDossier, ok := m.dossiers[v1.IdFacture]
		if !ok {
			return fmt.Errorf("message %d: missing migrated facture %d", v1.Id, v1.IdFacture)
		}
		event, ok := eventFromV1(v1, idDossier)
		if !ok {
			continue
		}
		inserted, err := event.Insert(tx)
		if err != nil {
			return err
		}
		idEvent := inserted.Id
		m.events[v1.Id] = idEvent

		switch v1.Kind {
		case rd.MResponsable, rd.MCentre:
			contenu := ""
			if complement, ok := messagesByID[v1.Id]; ok {
				contenu = string(complement.Contenu)
			}
			message := eventMessageFromV1(v1, contenu)
			message.IdEvent = idEvent
			if err := message.Insert(tx); err != nil {
				return err
			}
			for _, view := range eventMessageVusFromV1(viewsByID[v1.Id], campM) {
				view.IdEvent = idEvent
				if err := view.Insert(tx); err != nil {
					return err
				}
			}
		case rd.MAccuseReception:
			targets := inscritParticipantsForFacture(trimmed.participants, v1.IdFacture)
			if len(targets) == 0 {
				participant, ok := firstParticipantForFacture(trimmed.participants, v1.IdFacture)
				if !ok {
					continue
				}
				targets = []rd.Participant{participant}
			}
			baseEventUsed := false
			for _, participant := range targets {
				camp, ok := m.camps[participant.IdCamp]
				if !ok {
					continue
				}
				idStatuationEvent := idEvent
				if baseEventUsed {
					// EventStatuation has UNIQUE(IdEvent): each statuation needs its own event.
					extra, err := event.Insert(tx)
					if err != nil {
						return err
					}
					idStatuationEvent = extra.Id
				}
				baseEventUsed = true
				statuation := ev.EventStatuation{
					IdEvent:       idStatuationEvent,
					IdCamp:        camp.Id,
					IsBackoffice:  true,
					IdParticipant: m.participants[participant.Id],
					Statut:        cps.Inscrit,
				}
				if err := statuation.Insert(tx); err != nil {
					return err
				}
			}
		case rd.MDocuments:
			complement, ok := documentsByID[v1.Id]
			if !ok {
				continue
			}
			camp, ok := m.camps[complement.IdCamp]
			if !ok {
				continue
			}
			v2 := ev.EventCampDocs{IdEvent: idEvent, IdCamp: camp.Id}
			if err := v2.Insert(tx); err != nil {
				return err
			}
		case rd.MSondage:
			complement, ok := sondagesByID[v1.Id]
			if !ok {
				continue
			}
			camp, ok := m.camps[complement.IdCamp]
			if !ok {
				continue
			}
			v2 := ev.EventSondage{IdEvent: idEvent, IdCamp: camp.Id}
			if err := v2.Insert(tx); err != nil {
				return err
			}
		case rd.MPlaceLiberee:
			complement, ok := placesByID[v1.Id]
			if !ok {
				continue
			}
			idParticipant, ok := m.participants[complement.IdParticipant]
			if !ok {
				continue
			}
			v2 := ev.EventPlaceLiberee{
				IdEvent:       idEvent,
				IdParticipant: idParticipant,
				Accepted:      v1.Vu,
			}
			if err := v2.Insert(tx); err != nil {
				return err
			}
		case rd.MFactureAcquittee, rd.MAttestationPresence:
			complement, ok := attestationsByID[v1.Id]
			if !ok {
				continue
			}
			v2 := ev.EventAttestation{
				IdEvent:      idEvent,
				Distribution: distributionFromV1(complement.Distribution),
				IsPresence:   v1.Kind == rd.MAttestationPresence,
			}
			if err := v2.Insert(tx); err != nil {
				return err
			}
		}
	}
	return nil
}

func migrateCompta(tx *sql.Tx, trimmed baseV1, m idMaps) error {
	fmt.Println("\tcomptas...")
	for _, c := range trimmed.comptas {
		camp, ok := m.camps[c.IdCamp]
		if !ok {
			return fmt.Errorf("compta %d: missing migrated camp %d", c.Id, c.IdCamp)
		}
		v2, err := compta.ComptaCamp{
			IdCamp:                camp.Id,
			FileStatus:            compta.FileStatus(c.FileStatus),
			State:                 compta.State(c.State),
			LastSentMailAdmin:     c.LastReviewNotif,
			LastSentMailDirecteur: c.LastReviewNotif,
			Account:               compta.AccountACVE(c.Account),
			FraisMDR:              fraisACVE(c.Fraismdr),
		}.Insert(tx)
		if err != nil {
			return err
		}
		m.comptas[c.Id] = v2.Id
	}

	fmt.Println("\trecords...")
	for _, r := range trimmed.records {
		account := compta.AccountCamp
		if r.IsOtherAccount {
			account = compta.AccountAsso
		}
		v2, err := (compta.ComptaRecord{
			IdCamp:      m.comptas[r.IdCompta],
			IsReceipe:   r.IsReceipe,
			SpentBy:     compta.OptIdEquipier{},
			Reviews:     reviews(r.Reviews),
			IsValidated: r.IsValidated,
			Account:     account,
		}).Insert(tx)
		if err != nil {
			return err
		}
		m.records[r.Id] = v2.Id
	}

	fmt.Println("\tfilecomptas (metadata only)...")
	// Compta file contents are intentionally out of scope: only Name and the
	// two Drive IDs are migrated.
	for _, f := range trimmed.filecomptas {
		v2, err := (compta.ComptaFile{
			Name:             f.Name,
			DriveID:          f.DriveID,
			DriveIDMiniature: f.DriveIDMiniature,
		}).Insert(tx)
		if err != nil {
			return err
		}
		m.comptafiles[f.Id] = v2.Id
	}

	fmt.Println("\tsubrecords...")
	for _, r := range trimmed.subrecords {
		idFile := compta.OptIdComptaFile{}
		if r.IdFilecompta.Valid {
			idFile = compta.NewOptIdComptaFile(m.comptafiles[r.IdFilecompta.Id])
		}
		v2, err := (compta.ComptaSubrecord{
			IdRecord:    m.records[r.IdRecord],
			Date:        shared.NewDateFrom(r.Date.Time()),
			Description: r.Title,
			Mode:        compta.ModePaiementRecord(r.Mode),
			Serial:      r.Serial,
			Currency:    ds.Euros,
			Montants:    montants(r.Montants),
			IdFile:      idFile,
		}).Insert(tx)
		if err != nil {
			return err
		}
		m.subrecords[r.Id] = v2.Id
	}

	fmt.Println("\trecord_buss...")
	for _, link := range trimmed.busRecords {
		v2 := compta.ComptaRecordBus{
			IdCamp:      m.comptas[link.IdCompta],
			IdSubrecord: m.subrecords[link.IdSubrecord],
		}
		if err := v2.Insert(tx); err != nil {
			return err
		}
	}

	fmt.Println("\tbudget configs...")
	existingConfigs, err := compta.SelectAllComptaConfigs(tx)
	if err != nil {
		return err
	}
	existingYears := make(map[int]bool, len(existingConfigs))
	for _, config := range existingConfigs {
		existingYears[config.Year] = true
	}
	for _, c := range trimmed.budgetConfigs {
		if existingYears[c.Year] {
			warningf("compta config for %d already exists in v2, skipping", c.Year)
			continue
		}
		if _, err := budgetConfigFromV1(c).Insert(tx); err != nil {
			return err
		}
		existingYears[c.Year] = true
	}

	fmt.Println("\tbudgets...")
	for _, b := range trimmed.budgets {
		if err := budgetFromV1(b, m.comptas[b.IdCompta]).Insert(tx); err != nil {
			return err
		}
	}
	return nil
}
