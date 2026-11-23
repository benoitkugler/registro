package services

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"registro/config"
	"registro/crypto"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	tu "registro/utils/testutils"

	"github.com/labstack/echo/v4"
)

func TestRedirectLegacyEspacePerso(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	ct := NewController(db.DB, crypto.Encrypter{}, config.SMTP{}, config.Asso{})
	respo, err := pr.Personne{}.Insert(db.DB)
	tu.AssertNoErr(t, err)
	dossier, err := ds.Dossier{IdTaux: 1, IdResponsable: respo.Id, KeyV1: "legacy-key"}.Insert(db.DB)
	tu.AssertNoErr(t, err)

	e := echo.New()
	e.GET("/espace_perso/:key", ct.RedirectLegacyEspacePerso)

	const host = "acve.fr"
	redirect := func(path string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Host = host
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec
	}

	rec := redirect("/espace_perso/legacy-key")
	tu.Assert(t, rec.Code == http.StatusSeeOther)
	location, err := url.Parse(rec.Header().Get("Location"))
	tu.AssertNoErr(t, err)
	tu.Assert(t, location.Scheme == "https")
	tu.Assert(t, location.Host == host)
	tu.Assert(t, location.Path == "/espace-perso")
	token := location.Query().Get("token")
	id, err := crypto.DecryptID[ds.IdDossier](ct.key, token)
	tu.AssertNoErr(t, err)
	tu.Assert(t, id == dossier.Id)

	rec = redirect("/espace_perso/unknown-key")
	tu.Assert(t, rec.Code == http.StatusSeeOther)
	tu.Assert(t, rec.Header().Get("Location") == "/")
}

func TestSearchMail(t *testing.T) {
	db := tu.NewTestDB(t, "../../migrations/create_1_tables.sql",
		"../../migrations/create_2_json_funcs.sql", "../../migrations/create_3_constraints.sql",
		"../../migrations/init.sql")
	defer db.Remove()

	tu.LoadEnv(t, "../../env.sh")
	smtp, err := config.NewSMTP(false)
	tu.AssertNoErr(t, err)
	asso, err := config.NewAsso()
	tu.AssertNoErr(t, err)

	ct := NewController(db.DB, crypto.Encrypter{}, smtp, asso)

	got, err := ct.searchMailAndSend("", "")
	tu.AssertNoErr(t, err)
	tu.Assert(t, got.Found == 0)

	respo, err := pr.Personne{Identite: pr.Identite{Mail: "xx@free.fr", DateNaissance: shared.NewDate(2000, 1, 1)}}.Insert(ct.db)
	tu.AssertNoErr(t, err)
	p2, err := pr.Personne{}.Insert(ct.db)
	tu.AssertNoErr(t, err)
	p3, err := pr.Personne{}.Insert(ct.db)
	tu.AssertNoErr(t, err)
	_, err = pr.Personne{}.Insert(ct.db)
	tu.AssertNoErr(t, err)

	camp, err := cps.Camp{IdTaux: 1}.Insert(db)
	tu.AssertNoErr(t, err)

	dossier, err := ds.Dossier{IdTaux: 1, IdResponsable: respo.Id, KeyV1: "legacy-key-1"}.Insert(ct.db)
	tu.AssertNoErr(t, err)
	_, err = ds.Dossier{IdTaux: 1, IdResponsable: respo.Id, KeyV1: "legacy-key-2"}.Insert(ct.db)
	tu.AssertNoErr(t, err)

	_, err = cps.Participant{IdCamp: camp.Id, IdDossier: dossier.Id, IdTaux: 1, IdPersonne: p2.Id}.Insert(ct.db)
	tu.AssertNoErr(t, err)
	_, err = cps.Participant{IdCamp: camp.Id, IdDossier: dossier.Id, IdTaux: 1, IdPersonne: p3.Id}.Insert(ct.db)
	tu.AssertNoErr(t, err)

	out, err := ct.searchMailAndSend("", "xx@free.fr")
	tu.AssertNoErr(t, err)
	tu.Assert(t, out.Found == 2)
}
