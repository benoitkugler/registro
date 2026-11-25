package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	rd "registro/migrations/fromACVEv1/rawdata"
	cps "registro/sql/camps"
	ev "registro/sql/events"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
)

func TestValidateMigration(t *testing.T) {
	opts := migrationOptions{v1FilesDir: "/tmp/v1"}
	if err := validateMigration(baseV1{}, opts); err == nil {
		t.Fatal("validateMigration() accepted a missing v2 files directory")
	}

	dirs := t.TempDir()
	valid := migrationOptions{v1FilesDir: dirs, v2FilesDir: dirs}
	if err := validateMigration(baseV1{}, valid); err != nil {
		t.Fatalf("validateMigration() rejected valid file directories: %v", err)
	}
	badDestination := migrationOptions{
		v1FilesDir: dirs,
		v2FilesDir: filepath.Join(dirs, "missing"),
	}
	if err := validateMigration(baseV1{}, badDestination); err == nil {
		t.Fatal("validateMigration() accepted a missing v2 files directory")
	}

	trimmed := baseV1{participants: rd.Participants{1: {Id: 1}}}
	if err := validateMigration(trimmed, migrationOptions{}); err == nil {
		t.Fatal("validateMigration() accepted a participant without a facture")
	}
	trimmed.participants[1] = rd.Participant{Id: 1, IdFacture: rd.NewOptionnalId(100)}
	if err := validateMigration(trimmed, migrationOptions{}); err != nil {
		t.Fatalf("validateMigration() = %v, want nil", err)
	}
	withImages := baseV1{imageuploadeds: rd.Imageuploadeds{{IdCamp: 1}}}
	if err := validateMigration(withImages, migrationOptions{}); err == nil {
		t.Fatal("validateMigration() accepted lettre images without a host")
	}
	if err := validateMigration(withImages, migrationOptions{host: "example.com"}); err != nil {
		t.Fatalf("validateMigration() rejected a host: %v", err)
	}
}

func TestLettreHTMLFromV1(t *testing.T) {
	const oldHTML = `<img style="float: right" src="https://acve.fr/api/imageslettre?lien=v1LinkToInsertHere" width="404" height="495"/>`
	const newURL = "https://example.com/service/lettre-images?key=ENCRYPTED"

	got := lettreHTMLFromV1(oldHTML, map[string]string{
		"v1LinkToInsertHere": newURL,
	})
	want := `<img style="float: right" src="` + newURL + `" width="404" height="495"/>`
	if got != want {
		t.Errorf("lettreHTMLFromV1() = %q, want %q", got, want)
	}

	if got := lettreHTMLFromV1(oldHTML, map[string]string{"otherLink": newURL}); got != oldHTML {
		t.Errorf("unrelated lien was replaced: %q", got)
	}
	if got := lettreHTMLFromV1(oldHTML, nil); got != oldHTML {
		t.Errorf("no mapping should leave HTML untouched: %q", got)
	}
}

func TestSpecialPrixParAgeFromV1(t *testing.T) {
	const raw = `{"jour": [0, 0, 0, 0, 0, 0, 0], "active": "statut", "statut": [{"id": 1, "prix": 280, "statut": "Plus de 18 ans"}, {"id": 2, "prix": 231, "statut": "10-14 ans "}, {"id": 3, "prix": 154, "statut": "6-9 ans"}, {"id": 4, "prix": 98, "statut": "3-5 ans "}, {"id": 5, "statut": "0-2 ans "}], "semaine": {"plage_1": {"to": "0001-01-01", "from": "0001-01-01"}, "plage_2": {"to": "0001-01-01", "from": "0001-01-01"}}}`
	var opt rd.OptionPrixCamp
	if err := json.Unmarshal([]byte(raw), &opt); err != nil {
		t.Fatal(err)
	}

	got, ok := specialPrixParAgeFromV1(campIDPrixParAge, opt)
	if !ok {
		t.Fatal("special case not detected")
	}
	want := cps.OptionPrixCamp{
		Active: cps.PrixAge,
		PrixParAge: cps.PrixParAge{
			{AgeMax: 3, Prix: 0},
			{AgeMax: 6, Prix: 9800},
			{AgeMax: 10, Prix: 15400},
			{AgeMax: 15, Prix: 23100},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("specialPrixParAgeFromV1() = %+v, want %+v", got, want)
	}
	if err := got.PrixParAge.Check(); err != nil {
		t.Errorf("invalid grid: %s", err)
	}

	if _, ok := specialPrixParAgeFromV1(campIDPrixParAge+1, opt); ok {
		t.Error("special case applied to another camp")
	}
	changed := opt
	changed.Statut = append([]rd.PrixParStatut{}, opt.Statut...)
	changed.Statut[0].Prix = 300
	if _, ok := specialPrixParAgeFromV1(campIDPrixParAge, changed); ok {
		t.Error("special case applied to modified data")
	}
}

func TestWarningsFromV1(t *testing.T) {
	date := rd.Date(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	trimmed := baseV1{
		camps: rd.Camps{
			2: {Id: 2, Nom: "Camp B", DateDebut: date, OptionPrix: rd.OptionPrixCamp{Active: rd.OptionsPrix.SEMAINE}},
			1: {Id: 1, Nom: "Camp A", DateDebut: date, OptionPrix: rd.OptionPrixCamp{Active: rd.OptionsPrix.JOUR, Jour: []rd.Euros{10, 20}}},
		},
		messages: rd.Messages{
			1: {Id: 1, Kind: rd.MPaiement, IdFacture: 10},
			2: {Id: 2, Kind: rd.MAccuseReception, IdFacture: 20},
			3: {Id: 3, Kind: rd.MAccuseReception, IdFacture: 21}, // non-inscrit participant: fallback, no warning
			4: {Id: 4, Kind: rd.MAccuseReception, IdFacture: 22}, // inscrit participant: no warning
		},
		participants: rd.Participants{
			1: {Id: 1, IdFacture: rd.NewOptionnalId(21), ListeAttente: rd.ListeAttente{Statut: rd.Attente}},
			2: {Id: 2, IdFacture: rd.NewOptionnalId(22)},
		},
		documents: rd.Documents{100: {}, 101: {}},
	}

	v1Warnings := []string{
		"camp 1: Camp A 2024 v1 daily prices are not uniform and were reduced to one v2 price",
		`camp 2 Camp B 2024: v1 pricing option "semaine" has no v2 equivalent`,
		fmt.Sprintf("message 1: kind %d has no v2 equivalent", rd.MPaiement),
		"message 2: no participant found for statuation",
	}
	want := append(append([]string{}, v1Warnings...),
		"2 legacy document contents were not migrated (pass -v1-files-dir to copy them)")
	opts := migrationOptions{skipDocumentContents: true}
	if got := warningsFromV1(trimmed, opts); !reflect.DeepEqual(got, want) {
		t.Errorf("warningsFromV1() = %#v, want %#v", got, want)
	}

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "file_100"), []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	want = append(append([]string{}, v1Warnings...),
		"1 legacy document contents were missing from "+dir)
	opts.v1FilesDir = dir
	if got := warningsFromV1(trimmed, opts); !reflect.DeepEqual(got, want) {
		t.Errorf("warningsFromV1() with partial files = %#v, want %#v", got, want)
	}

	if err := os.WriteFile(filepath.Join(dir, "file_101"), []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := warningsFromV1(trimmed, opts); !reflect.DeepEqual(got, v1Warnings) {
		t.Errorf("warningsFromV1() with all files = %#v, want %#v", got, v1Warnings)
	}
}

func TestWarningsMissingMiniatures(t *testing.T) {
	trimmed := baseV1{
		documents: rd.Documents{1: {}, 2: {}, 3: {}},
		contenuDocuments: rd.ContenuDocuments{
			{IdDocument: 1, Miniature: []byte("miniature")},
			{IdDocument: 2},
		},
	}
	got := warningsFromV1(trimmed, migrationOptions{})
	want := []string{
		"3 legacy document contents were not migrated (pass -v1-files-dir to copy them)",
		"2 document(s) have no miniature in contenu_documents (v2 will use a placeholder)",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("warningsFromV1() = %#v, want %#v", got, want)
	}
}

func TestCategorieFromV1(t *testing.T) {
	cases := []struct {
		builtin rd.BuiltinContrainte
		want    fs.Categorie
		label   bool
	}{
		{rd.CBafa, fs.Bafa, false},
		{rd.CBafaEquiv, fs.Bafa, false},
		{rd.CBafdEquiv, fs.Bafd, false},
		{rd.CVaccin, fs.Vaccins, false},
		{rd.CAutre, fs.Autre, false},
		{rd.CInvalide, fs.NoBuiltin, true},
		{rd.CTestNautique, fs.NoBuiltin, true},
		{rd.CHaccp, fs.NoBuiltin, true},
	}
	for _, tc := range cases {
		got, label := categorieFromV1(tc.builtin)
		if got != tc.want {
			t.Errorf("categorieFromV1(%q) categorie = %d, want %d", tc.builtin, got, tc.want)
		}
		if tc.label && label == "" {
			t.Errorf("categorieFromV1(%q) missing legacy label", tc.builtin)
		}
		if !tc.label && label != "" {
			t.Errorf("categorieFromV1(%q) unexpected label %q", tc.builtin, label)
		}
	}
}

func TestIsObsoleteBuiltin(t *testing.T) {
	cases := map[rd.BuiltinContrainte]bool{
		rd.CBafa:                        false,
		rd.CBafaEquiv:                   false,
		rd.CAutre:                       false,
		rd.CInvalide:                    false,
		rd.CTestNautique:                false,
		rd.CHaccp:                       true,
		rd.CCertMedCuisine:              true,
		rd.CScolarite:                   true,
		rd.BuiltinContrainte("unknown"): true,
	}
	for builtin, want := range cases {
		if got := isObsoleteBuiltin(builtin); got != want {
			t.Errorf("isObsoleteBuiltin(%q) = %t, want %t", builtin, got, want)
		}
	}
}

func TestBuiltinDemandeReuse(t *testing.T) {
	var builtins fs.Builtins
	builtins[fs.Bafa] = fs.Demande{
		Id:          fs.IdDemande(fs.Bafa),
		Categorie:   fs.Bafa,
		MaxDocs:     1,
		IdDirecteur: pr.OptIdPersonne{},
	}
	builtins[fs.Autre] = fs.Demande{
		Id:          fs.IdDemande(fs.Autre),
		Categorie:   fs.Autre,
		MaxDocs:     1,
		IdDirecteur: pr.OptIdPersonne{},
	}

	demande, ok := builtinDemandeFromV1(rd.Contrainte{Builtin: rd.CBafa}, builtins)
	if !ok {
		t.Fatal("builtinDemandeFromV1() did not reuse the predeclared row")
	}
	if demande.Id != fs.IdDemande(fs.Bafa) || demande.IdFile.Valid {
		t.Errorf("reused demande = %+v, want predeclared id and no global file", demande)
	}

	autre, ok := builtinDemandeFromV1(rd.Contrainte{Builtin: rd.CAutre}, builtins)
	if !ok || autre.Id != fs.IdDemande(fs.Autre) {
		t.Errorf("CAutre should reuse the v2 Autre builtin: %+v, %t", autre, ok)
	}

	for _, builtin := range []rd.BuiltinContrainte{rd.CInvalide, rd.CTestNautique, rd.CHaccp} {
		custom, ok := builtinDemandeFromV1(rd.Contrainte{Builtin: builtin}, builtins)
		if ok || custom.Id != 0 {
			t.Errorf("builtin %q should not reuse a builtin demande: %+v", builtin, custom)
		}
	}
}

func TestTrimContraintesDropsOnlyObsoleteBuiltins(t *testing.T) {
	b := baseV1{
		camps: rd.Camps{1: {Id: 1}},
		contraintes: rd.Contraintes{
			1: {Id: 1, Builtin: rd.CBafa},
			2: {Id: 2, Builtin: rd.CHaccp},
			3: {Id: 3, Builtin: rd.CTestNautique},
			4: {Id: 4, Builtin: rd.CInvalide, Nom: "custom"},
		},
		campContraintes: rd.CampContraintes{
			{IdCamp: 1, IdContrainte: 1},
			{IdCamp: 1, IdContrainte: 2},
			{IdCamp: 1, IdContrainte: 3},
			{IdCamp: 1, IdContrainte: 4},
		},
	}
	out := baseV1{camps: b.camps}

	contraintes, campLinks, _, _ := trimContraintes(b, out)
	for _, id := range []int64{1, 3, 4} {
		if _, ok := contraintes[id]; !ok {
			t.Errorf("constraint %d was dropped", id)
		}
	}
	if len(contraintes) != 3 {
		t.Errorf("contraintes = %+v, want 3 retained constraints", contraintes)
	}
	if len(campLinks) != 3 {
		t.Fatalf("campLinks = %+v, want 3 retained links", campLinks)
	}
	for _, link := range campLinks {
		if link.IdContrainte == 2 {
			t.Errorf("obsolete constraint link was retained: %+v", link)
		}
	}
}

func TestFileCampsFromV1Dedup(t *testing.T) {
	m := newIDMaps()
	m.camps[1] = cps.Camp{Id: 11}
	m.files[10] = fs.File{Id: 100}
	m.files[20] = fs.File{Id: 200}
	m.files[21] = fs.File{Id: 201}

	links := rd.DocumentCamps{
		{IdDocument: 10, IdCamp: 1, IsLettre: false},
		{IdDocument: 20, IdCamp: 1, IsLettre: true},
		{IdDocument: 21, IdCamp: 1, IsLettre: true},
	}
	got := fileCampsFromV1(links, m)
	want := []fs.FileCamp{
		{IdFile: 200, IdCamp: 11, IsLettre: true},
		{IdFile: 100, IdCamp: 11, IsLettre: false},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("fileCampsFromV1() = %+v, want %+v", got, want)
	}
}

func TestFilePersonnesFromV1SkipsTemp(t *testing.T) {
	m := newIDMaps()
	m.files[10] = fs.File{Id: 100}
	m.files[11] = fs.File{Id: 101}
	m.personnes[1] = 21
	m.personnes[2] = 22
	m.tempPersonnes[2] = true
	m.demandes[5] = fs.IdDemande(5)

	links := rd.DocumentPersonnes{
		{IdDocument: 10, IdPersonne: 1, IdContrainte: 5},
		{IdDocument: 11, IdPersonne: 2, IdContrainte: 5},
	}
	got := filePersonnesFromV1(links, m)
	want := []fs.FilePersonne{{IdFile: 100, IdPersonne: 21, IdDemande: 5}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("filePersonnesFromV1() = %+v, want %+v", got, want)
	}
}

func TestFileAidesFromV1Dedup(t *testing.T) {
	m := newIDMaps()
	m.files[10] = fs.File{Id: 100}
	m.files[11] = fs.File{Id: 101}
	m.aides[7] = cps.IdAide(70)

	links := rd.DocumentAides{
		{IdDocument: 10, IdAide: 7},
		{IdDocument: 11, IdAide: 7},
	}
	got := fileAidesFromV1(links, m)
	want := []fs.FileAide{{IdFile: 100, IdAide: 70}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("fileAidesFromV1() = %+v, want %+v", got, want)
	}
}

func TestFicheSanitaireFromV1(t *testing.T) {
	modified := time.Date(2025, time.June, 1, 12, 0, 0, 0, time.UTC)
	personne := rd.Personne{FicheSanitaire: rd.FicheSanitaire{
		LastModif: rd.Time(modified),
		Maladies:  rd.Maladies{Rubeole: true},
		Allergies: rd.Allergies{
			Asthme:       true,
			Alimentaires: true,
			Autres:       "pollen",
		},
		TraitementMedical: true,
		Tel:               "0102030405",
		Medecin:           rd.Medecin{Nom: "Dr Martin"},
	}}

	got, ok := ficheSanitaireFromV1(personne)
	if !ok {
		t.Fatal("ficheSanitaireFromV1() returned false")
	}
	if got.AllergiesAlimentaires != "oui" || got.TraitementMedical != "oui" {
		t.Errorf("yes/no fields = %q, %q", got.AllergiesAlimentaires, got.TraitementMedical)
	}
	if !strings.Contains(got.DifficultesSante, "Rubéole") ||
		!strings.Contains(got.DifficultesSante, "Asthme") ||
		strings.Contains(got.DifficultesSante, "Dr Martin") {
		t.Errorf("DifficultesSante = %q, want legacy free-text without medecin", got.DifficultesSante)
	}
	if got.AutreContact.Nom != "" || got.AutreContact.Tel != "0102030405" {
		t.Errorf("AutreContact = %+v, want no medecin name", got.AutreContact)
	}
	if got.ReviewStatus != pr.FileReviewPending {
		t.Errorf("ReviewStatus = %d, want pending", got.ReviewStatus)
	}
}

func TestCampFromV1QuotientAndNoJoomeoAlbum(t *testing.T) {
	camp := rd.Camp{
		JoomeoAlbumId: "legacy-album",
		QuotientFamilial: rd.OptionQuotientFamilial{
			{QuotientMin: 0, Pourcentage: 20},
			{QuotientMin: 359, Pourcentage: 40},
			{QuotientMin: 564, Pourcentage: 60},
			{QuotientMin: 714, Pourcentage: 100},
		},
	}
	got := campFromV1(camp)
	if got.IsAlbumVisible || got.AlbumID != "" {
		t.Errorf("deprecated Joomeo album was migrated: %q, %t", got.AlbumID, got.IsAlbumVisible)
	}
	if len(got.OptionQuotientFamilial) != 4 {
		t.Fatalf("quotient grid = %+v", got.OptionQuotientFamilial)
	}
	wantGrid := cps.PrixQuotientFamilial{
		{QuotientMin: 1, Pourcentage: 20},
		{QuotientMin: 360, Pourcentage: 40},
		{QuotientMin: 565, Pourcentage: 60},
		{QuotientMin: 715, Pourcentage: 100},
	}
	if !reflect.DeepEqual(got.OptionQuotientFamilial, wantGrid) {
		t.Errorf("quotient grid = %+v, want %+v", got.OptionQuotientFamilial, wantGrid)
	}
	empty := campFromV1(rd.Camp{})
	if empty.IsAlbumVisible {
		t.Error("empty Joomeo album should not be visible")
	}
	if len(empty.OptionQuotientFamilial) != 0 {
		t.Errorf("empty quotient grid = %+v, want no tranche", empty.OptionQuotientFamilial)
	}
}

func TestEventKindFromV1(t *testing.T) {
	cases := map[rd.MessageKind]ev.EventKind{
		rd.MSupprime:            ev.Supprime,
		rd.MResponsable:         ev.Message,
		rd.MCentre:              ev.Message,
		rd.MAccuseReception:     ev.Statuation,
		rd.MFacture:             ev.DemandePaiement,
		rd.MDocuments:           ev.CampDocs,
		rd.MFactureAcquittee:    ev.Attestation,
		rd.MAttestationPresence: ev.Attestation,
		rd.MSondage:             ev.Sondage,
		rd.MPlaceLiberee:        ev.PlaceLiberee,
	}
	for v1, want := range cases {
		got, ok := eventKindFromV1(v1)
		if !ok || got != want {
			t.Errorf("eventKindFromV1(%d) = %d, %t; want %d", v1, got, ok, want)
		}
	}
	if _, ok := eventKindFromV1(rd.MPaiement); ok {
		t.Error("MPaiement should have no v2 equivalent")
	}
	if _, ok := eventKindFromV1(rd.MInscription); ok {
		t.Error("MInscription events should be ignored")
	}
}

func TestInscritParticipantsForFacture(t *testing.T) {
	participant := func(id, idFacture int64, statut rd.StatutAttente) rd.Participant {
		return rd.Participant{
			Id:           id,
			IdFacture:    rd.NewOptionnalId(idFacture),
			ListeAttente: rd.ListeAttente{Statut: statut},
		}
	}
	participants := rd.Participants{
		1: participant(1, 10, rd.Inscrit),
		2: participant(2, 10, rd.Attente),
		3: participant(3, 10, rd.Inscrit),
		4: participant(4, 11, rd.Inscrit),
		5: {Id: 5, ListeAttente: rd.ListeAttente{Statut: rd.Inscrit}},
	}

	got := inscritParticipantsForFacture(participants, 10)
	seen := make(map[int64]bool)
	for _, p := range got {
		seen[p.Id] = true
	}
	if len(seen) != 2 || !seen[1] || !seen[3] {
		t.Fatalf("inscrit participants = %+v, want 1 and 3", got)
	}

	noInscrits := rd.Participants{
		2: participant(2, 10, rd.Attente),
		4: participant(4, 11, rd.Inscrit),
	}
	if got := inscritParticipantsForFacture(noInscrits, 10); len(got) != 0 {
		t.Errorf("no inscrit participants = %+v, want empty", got)
	}
	fallback, ok := firstParticipantForFacture(noInscrits, 10)
	if !ok || fallback.Id != 2 {
		t.Errorf("fallback participant = %+v, %t; want 2", fallback, ok)
	}
}

func TestTrimMessagesIgnoresMInscription(t *testing.T) {
	b := baseV1{
		factures: rd.Factures{1: {Id: 1}},
		messages: rd.Messages{
			1: {Id: 1, IdFacture: 1, Kind: rd.MResponsable},
			2: {Id: 2, IdFacture: 1, Kind: rd.MInscription},
		},
		messageMessages: rd.MessageMessages{
			{IdMessage: 1, Contenu: "kept"},
			{IdMessage: 2, Contenu: "dropped"},
		},
	}
	out := baseV1{factures: b.factures}

	messages, _, libres, _, _, _, _ := trimMessages(b, out)
	if _, ok := messages[1]; !ok {
		t.Error("MResponsable message was dropped")
	}
	if _, ok := messages[2]; ok {
		t.Error("MInscription message was retained")
	}
	if len(libres) != 1 || libres[0].IdMessage != 1 {
		t.Errorf("messageMessages = %+v, want only message 1", libres)
	}
}

func TestEventMessageFromV1VuDirection(t *testing.T) {
	fromResponsable := eventMessageFromV1(rd.Message{Kind: rd.MResponsable, Vu: true}, "hello")
	if fromResponsable.Origine != ev.Espaceperso || !fromResponsable.VuBackoffice || fromResponsable.VuEspaceperso {
		t.Errorf("MResponsable = %+v", fromResponsable)
	}

	fromCentre := eventMessageFromV1(rd.Message{Kind: rd.MCentre, Vu: true}, "hello")
	if fromCentre.Origine != ev.Backoffice || !fromCentre.VuEspaceperso || fromCentre.VuBackoffice {
		t.Errorf("MCentre = %+v", fromCentre)
	}
}

func TestLegacyDocumentCopyIsStaged(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "file_10"), []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	// No file_10_min on disk: the miniature must come from contenu_documents.

	opts := migrationOptions{v1FilesDir: src, v2FilesDir: dst}
	pending, err := stageLegacyDocument(opts, 10, 100, []byte("mini"))
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 1 || string(pending[0].content) != "content" || string(pending[0].miniature) != "mini" {
		t.Fatalf("staged copies = %+v", pending)
	}
	if _, err := os.Stat(filepath.Join(dst, "file_100")); !os.IsNotExist(err) {
		t.Fatalf("file written before commit: %v", err)
	}

	if err := applyPendingFiles(opts, pending); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(dst, "file_100"))
	if err != nil || string(content) != "content" {
		t.Fatalf("staged content = %q, %v", content, err)
	}
	miniature, err := os.ReadFile(filepath.Join(dst, "file_100_min"))
	if err != nil || string(miniature) != "mini" {
		t.Fatalf("staged miniature = %q, %v", miniature, err)
	}

	pending, err = stageLegacyDocument(opts, 10, 100, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 1 || pending[0].miniature != nil {
		t.Fatalf("missing miniature should stay empty: %+v", pending)
	}
}

func TestBaseV1OmitsDoncs(t *testing.T) {
	typ := reflect.TypeOf(baseV1{})
	for i := 0; i < typ.NumField(); i++ {
		if strings.EqualFold(typ.Field(i).Name, "doncs") {
			t.Fatal("baseV1 must not load v1 doncs")
		}
	}
}
