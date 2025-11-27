package backoffice

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	fsAPI "registro/controllers/files"
	"registro/controllers/inscriptions"
	"registro/generators/sheets"
	"registro/logic"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

func (ct *Controller) CampsGetTaux(c echo.Context) error {
	out, err := ct.loadTaux()
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) loadTaux() (ds.Tauxs, error) {
	return ds.SelectAllTauxs(ct.db)
}

type CampHeader struct {
	Camp              cps.CampExt
	Taux              ds.Taux
	Stats             cps.StatistiquesInscrits
	ParticipantsFiles []fsAPI.DemandeStat
	HasDirecteur      bool

	URLPreselection string
	URLDirecteur    string // with loggin access
}

func (ct *Controller) CampsGet(c echo.Context) error {
	out, err := ct.getCamps(c.Request().Host)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) getCamps(host string) ([]CampHeader, error) {
	camps, err := cps.SelectAllCamps(ct.db)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	ids := camps.IDs()
	taux, err := ds.SelectTauxs(ct.db, camps.IdTauxs()...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	loaders, err := cps.LoadCamps(ct.db, ids)
	if err != nil {
		return nil, err
	}

	directeurs, err := loadDirecteurs(ct.db, ids)
	if err != nil {
		return nil, err
	}

	filesLoader, err := fsAPI.LoadParticipantsFiles(ct.db, ct.key, ids)
	if err != nil {
		return nil, err
	}

	out := make([]CampHeader, len(ids))
	for i, id := range ids {
		loader := loaders.For(id)
		files := filesLoader.For(id)
		_, hasDirecteur := directeurs[id]
		camp := loader.Camp.Ext()
		preselectionURL := utils.BuildUrl(host, inscriptions.EndpointInscription, utils.QP(inscriptions.PreselectionQueryParam, camp.Slug))
		directeurURL := directeurURL(ct.key, host, id)
		out[i] = CampHeader{camp, taux[loader.Camp.IdTaux], loader.Stats(), files.Stats(), hasDirecteur, preselectionURL, directeurURL}
	}
	return out, nil
}

func loadDirecteurs(db cps.DB, camps []cps.IdCamp) (map[cps.IdCamp]pr.Personne, error) {
	tmp, personnes, err := cps.LoadEquipiersByCamps(db, camps...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	equipiersByCamp := tmp.ByIdCamp()

	out := map[cps.IdCamp]pr.Personne{}
	for _, camp := range camps {
		dir, ok := equipiersByCamp[camp].Directeur()
		if ok {
			out[camp] = personnes[dir.IdPersonne]
		}
	}
	return out, nil
}

func ensureTaux(tx *sql.Tx, taux ds.Taux) (ds.Taux, error) {
	if taux.Id <= 0 { // create a new Taux
		return taux.Insert(tx)
	} // else, simply use the Id
	return taux, nil
}

type CampsCreateManyIn struct {
	// If Taux has an [Id] <= 0, it is created first,
	// Otherwise, only its [Id] is used.
	Taux ds.Taux
	// Count is the number of Camp to create
	Count int
}

func (ct *Controller) CampsCreateMany(c echo.Context) error {
	var args CampsCreateManyIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := ct.createManyCamp(args, c.Request().Host)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) createManyCamp(args CampsCreateManyIn, host string) (out []CampHeader, _ error) {
	err := utils.InTx(ct.db, func(tx *sql.Tx) error {
		var err error
		args.Taux, err = ensureTaux(tx, args.Taux)
		if err != nil {
			return err
		}
		for i := 0; i < args.Count; i++ {
			camp := defaultCamp(args.Taux.Id)
			camp.Nom += fmt.Sprintf(" %d", i+1)
			camp, err := camp.Insert(tx)
			if err != nil {
				return err
			}
			directeurURL := directeurURL(ct.key, host, camp.Id)
			out = append(out, CampHeader{camp.Ext(), args.Taux, cps.StatistiquesInscrits{}, nil, false, "", directeurURL})
		}
		return nil
	})
	return out, err
}

func (ct *Controller) CampsCreate(c echo.Context) error {
	out, err := ct.createCamp(c.Request().Host)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func defaultCamp(idTaux ds.IdTaux) cps.Camp {
	return cps.Camp{
		IdTaux:    idTaux,
		Nom:       "Nouveau séjour",
		DateDebut: shared.NewDateFrom(time.Now()), Duree: 10,
		Places: 40, AgeMin: 6, AgeMax: 12,
		Password: utils.RandPassword(6),
		Statut:   cps.VisibleFerme,
		Prix:     ds.NewEuros(100),
		DocumentsToShow: cps.DocumentsToShow{
			LettreDirecteur: true,
			ListeVetements:  true,
		},
	}
}

// select the last taux used
func lastTaux(camps cps.Camps) ds.IdTaux {
	// there is always the defaut taux present in the DB, by construction
	if len(camps) == 0 {
		return 1
	}
	var last cps.Camp
	for _, camp := range camps {
		if camp.DateDebut.Time().After(last.DateDebut.Time()) {
			last = camp
		}
	}
	return last.IdTaux
}

func (ct *Controller) createCamp(host string) (CampHeader, error) {
	camps, err := cps.SelectAllCamps(ct.db)
	if err != nil {
		return CampHeader{}, utils.SQLError(err)
	}
	idTaux := lastTaux(camps)

	camp, err := defaultCamp(idTaux).Insert(ct.db)
	if err != nil {
		return CampHeader{}, utils.SQLError(err)
	}
	taux, err := ds.SelectTaux(ct.db, camp.IdTaux)
	if err != nil {
		return CampHeader{}, utils.SQLError(err)
	}
	directeurURL := directeurURL(ct.key, host, camp.Id)
	return CampHeader{camp.Ext(), taux, cps.StatistiquesInscrits{}, nil, false, "", directeurURL}, nil
}

func (ct *Controller) CampsUpdate(c echo.Context) error {
	var args cps.Camp
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := ct.updateCamp(args)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) updateCamp(args cps.Camp) (cps.CampExt, error) {
	camp, err := cps.SelectCamp(ct.db, args.Id)
	if err != nil {
		return cps.CampExt{}, utils.SQLError(err)
	}

	if err = args.Check(); err != nil {
		return cps.CampExt{}, err
	}
	if err := checkCurrency(ct.db, camp.IdTaux, args.Prix.Currency); err != nil {
		return cps.CampExt{}, err
	}

	camp.Nom = args.Nom
	camp.DateDebut = args.DateDebut
	camp.Duree = args.Duree
	camp.Lieu = args.Lieu
	camp.Agrement = args.Agrement
	camp.ImageURL = args.ImageURL
	camp.Description = args.Description
	camp.Navette = args.Navette
	camp.Places = args.Places
	camp.AgeMin = args.AgeMin
	camp.AgeMax = args.AgeMax
	camp.Meta = args.Meta
	camp.NeedEquilibreGF = args.NeedEquilibreGF
	camp.WithoutInscription = args.WithoutInscription
	camp.Statut = args.Statut
	camp.Prix = args.Prix
	camp.OptionPrix = args.OptionPrix
	camp.OptionQuotientFamilial = args.OptionQuotientFamilial
	camp.Password = args.Password
	camp, err = camp.Update(ct.db)
	if err != nil {
		return cps.CampExt{}, utils.SQLError(err)
	}

	return camp.Ext(), nil
}

type CampsSetTauxIn struct {
	IdCamp cps.IdCamp
	// If Taux has an [Id] <= 0, it is created first,
	// Otherwise, only its [Id] is used.
	Taux ds.Taux
}

// CampsSetTaux permet à l'utilisateur de changer le Taux
// utilisé par un séjour donné, à condition qu'il n'y ait
// pas encore de participants.
func (ct *Controller) CampsSetTaux(c echo.Context) error {
	var args CampsSetTauxIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := ct.setTaux(args)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) setTaux(args CampsSetTauxIn) (out ds.Taux, _ error) {
	err := utils.InTx(ct.db, func(tx *sql.Tx) error {
		participants, err := cps.SelectParticipantsByIdCamps(tx, args.IdCamp)
		if err != nil {
			return err
		}
		if len(participants) > 0 {
			return errors.New("des participants sont déjà déclarés")
		}

		args.Taux, err = ensureTaux(tx, args.Taux)
		if err != nil {
			return err
		}
		camp, err := cps.SelectCamp(tx, args.IdCamp)
		if err != nil {
			return err
		}
		camp.IdTaux = args.Taux.Id
		camp, err = camp.Update(tx)
		if err != nil {
			return err
		}
		out = args.Taux
		return nil
	})
	return out, err
}

// CampsDelete supprime un camp SANS participants ni équipiers,
// renvoie une erreur sinon.
func (ct *Controller) CampsDelete(c echo.Context) error {
	id, err := utils.QueryParamInt[cps.IdCamp](c, "id")
	if err != nil {
		return err
	}
	err = ct.deleteCamp(id)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) deleteCamp(id cps.IdCamp) error {
	var toDelete []fs.IdFile
	err := utils.InTx(ct.db, func(tx *sql.Tx) error {
		// intégrité sur les participants
		// cascade sur les équipiers

		links, err := fs.DeleteFileCampsByIdCamps(tx, id)
		if err != nil {
			return err
		}

		toDelete, err = fs.DeleteFilesByIDs(tx, links.IdFiles()...)
		if err != nil {
			return err
		}

		// cascade sur les DemandeCamps et Groupes
		_, err = cps.DeleteCampById(tx, id)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	// cleanup the document content
	go func() {
		err = ct.files.Delete(toDelete...)
		if err != nil {
			log.Println(err)
		}
	}()

	return nil
}

type OuvreInscriptionsIn struct {
	Camps []cps.IdCamp
}

// CampsOuvreInscriptions est un raccourci permettant d'ouvrir
// d'un coup plusieurs séjours.
func (ct *Controller) CampsOuvreInscriptions(c echo.Context) error {
	var args OuvreInscriptionsIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.ouvreInscriptions(args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) ouvreInscriptions(args OuvreInscriptionsIn) error {
	camps, err := cps.SelectCamps(ct.db, args.Camps...)
	if err != nil {
		return utils.SQLError(err)
	}
	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		for _, camp := range camps {
			camp.Statut = cps.Ouvert
			_, err = camp.Update(tx)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// Documents

func (ct *Controller) CampsDocuments(c echo.Context) error {
	id, err := utils.QueryParamInt[cps.IdCamp](c, "idCamp")
	if err != nil {
		return err
	}
	out, err := ct.getCampDocument(id)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type FilesCamp struct {
	ToShow cps.DocumentsToShow

	Generated       []fsAPI.GeneratedFile
	ToRead          []logic.PublicFile
	ToUploadModeles []logic.PublicFile
}

func (ct *Controller) getCampDocument(id cps.IdCamp) (FilesCamp, error) {
	camp, err := cps.SelectCamp(ct.db, id)
	if err != nil {
		return FilesCamp{}, utils.SQLError(err)
	}
	links, err := fs.SelectFileCampsByIdCamps(ct.db, id)
	if err != nil {
		return FilesCamp{}, utils.SQLError(err)
	}
	campFiles, err := fs.SelectFiles(ct.db, links.IdFiles()...)
	if err != nil {
		return FilesCamp{}, utils.SQLError(err)
	}

	out := FilesCamp{ToShow: camp.DocumentsToShow}
	// other files
	for _, link := range links {
		out.ToRead = append(out.ToRead, logic.NewPublicFile(ct.key, campFiles[link.IdFile]))
	}

	// generated files
	doc1, err := fsAPI.CampDocument(ct.key, camp, fsAPI.ListeVetements)
	if err != nil {
		return FilesCamp{}, err
	}
	doc2, err := fsAPI.CampDocument(ct.key, camp, fsAPI.ListeParticipants)
	if err != nil {
		return FilesCamp{}, err
	}
	out.Generated = []fsAPI.GeneratedFile{doc1, doc2}

	links2, err := fs.SelectDemandeCampsByIdCamps(ct.db, id)
	if err != nil {
		return FilesCamp{}, utils.SQLError(err)
	}
	demandes, err := fs.SelectDemandes(ct.db, links2.IdDemandes()...)
	if err != nil {
		return FilesCamp{}, utils.SQLError(err)
	}
	demandesFiles, err := fs.SelectFiles(ct.db, demandes.IdFiles()...)
	if err != nil {
		return FilesCamp{}, utils.SQLError(err)
	}
	for _, file := range demandesFiles {
		out.ToUploadModeles = append(out.ToUploadModeles, logic.NewPublicFile(ct.key, file))
	}
	return out, nil
}

type CreateEquipierIn struct {
	IdPersonne pr.IdPersonne
	IdCamp     cps.IdCamp
	Roles      cps.Roles // to select default Demandes
}

func (ct *Controller) CampsCreateEquipier(c echo.Context) error {
	var args CreateEquipierIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := ct.createEquipier(args)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) createEquipier(args CreateEquipierIn) (out cps.Equipier, _ error) {
	err := utils.InTx(ct.db, func(tx *sql.Tx) error {
		var err error
		out, err = cps.Equipier{IdCamp: args.IdCamp, IdPersonne: args.IdPersonne, Roles: args.Roles}.Insert(tx)
		if err != nil {
			return err
		}
		demandes := ct.builtins.Defaut(out.Id, out.Roles)
		err = fs.InsertManyDemandeEquipiers(tx, demandes...)
		if err != nil {
			return err
		}
		return nil
	})
	return out, err
}

func (ct *Controller) CampUpdateEquipier(c echo.Context) error {
	var args cps.Equipier
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.updateEquipier(args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) updateEquipier(args cps.Equipier) error {
	_, err := args.Update(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}

func (ct *Controller) CampsDownloadParticipants(c echo.Context) error {
	year, err := utils.QueryParamInt[int](c, "year")
	if err != nil {
		return err
	}
	content, name, err := ct.exportListeParticipants(year)
	if err != nil {
		return err
	}
	mimeType := fsAPI.SetBlobHeader(c, content, name)
	return c.Blob(200, mimeType, content)
}

// export the [Inscrits] from the camp starting at the given year
func (ct *Controller) exportListeParticipants(year int) ([]byte, string, error) {
	// select the correct [Camp]s
	camps, err := cps.SelectAllCamps(ct.db)
	if err != nil {
		return nil, "", utils.SQLError(err)
	}
	camps.RestrictByYear(year)

	loader, err := cps.LoadCamps(ct.db, camps.IDs())
	if err != nil {
		return nil, "", err
	}
	// build inscrit liste
	var inscrits []cps.ParticipantCamp
	for _, camp := range loader.Camps {
		for _, p := range loader.For(camp.Id).Participants(false) {
			inscrits = append(inscrits, cps.ParticipantCamp{Camp: camp, ParticipantPersonne: p})
		}
	}

	dossiers, err := logic.LoadDossiersFinances(ct.db, loader.IdDossiers()...)
	if err != nil {
		return nil, "", err
	}

	// load equipiers
	equipiers, equipiersP, err := cps.LoadEquipiersByCamps(ct.db, camps.IDs()...)
	if err != nil {
		return nil, "", utils.SQLError(err)
	}

	remisesHintsL := estimeRemises(loader, dossiers.Dossiers, equipiers, equipiersP, ct.asso.RemisesHints, false)
	remisesHints := map[cps.IdParticipant]cps.Remises{}
	for _, p := range remisesHintsL {
		remisesHints[p.IdParticipant] = p.Hint
	}

	showNationnaliteSuisse := ct.asso.AskNationnalite
	content, err := sheets.ListeParticipantsCamps(inscrits, dossiers, remisesHints, showNationnaliteSuisse)
	if err != nil {
		return nil, "", err
	}
	name := fmt.Sprintf("Participants %d.xlsx", year)
	return content, name, nil
}
