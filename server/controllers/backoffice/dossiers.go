package backoffice

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	filesAPI "registro/controllers/files"
	"registro/logic"
	"registro/logic/search"
	"registro/mails"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	evs "registro/sql/events"
	"registro/sql/files"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

var OffuscateurVirements = newOffuscateur[ds.IdDossier]("IN", 8, 3)

type offuscateur[T ~int64] struct {
	prefix  string
	m, a, b T // m * b > a
}

// newOffuscateur panics if m < 2 or b < 2 or if prefix is empty
func newOffuscateur[T ~int64](prefix string, m, b T) offuscateur[T] {
	if m < 2 || b < 2 || prefix == "" {
		panic("invalid Offuscateur parameters")
	}
	a := m*b - 1
	return offuscateur[T]{prefix, m, a, b}
}

func (o offuscateur[T]) Mask(id T) string {
	v := (id+o.b)*o.m - o.a
	return fmt.Sprintf("%s%04d", o.prefix, v)
}

func (o offuscateur[T]) Unmask(code string) (id T, ok bool) {
	noPrefix := strings.TrimPrefix(code, o.prefix)
	if len(noPrefix) == len(code) {
		return 0, false
	}
	entry, err := strconv.ParseInt(noPrefix, 10, 64)
	if err != nil {
		return 0, false
	}
	a := T(entry) + o.a
	if a%o.m != 0 {
		return 0, false
	}
	return a/o.m - o.b, true
}

type QueryAttente uint8

const (
	EmptyQA         QueryAttente = iota // Indifférent
	AvecAttente                         // Avec liste d'attente
	AvecInscrits                        // Avec inscrits
	AvecAttenteOnly                     // Seulement avec liste d'attente
)

// Flag
type QueryReglement uint8

const (
	EmptyQR QueryReglement = 0               // Indifférent
	Zero    QueryReglement = 1 << (iota - 1) // Non commencé
	Partiel                                  // En cours
	Total                                    // Complété
)

func (qr QueryReglement) match(statut logic.StatutPaiement) bool {
	var flag QueryReglement = 1 << (statut - 1)
	return qr&flag != 0
}

// The zero value defaults to returning everything
type SearchDossierIn struct {
	Pattern           string // Responsable et participants
	IdCamp            evs.OptIdCamp
	Attente           QueryAttente
	Reglement         QueryReglement
	SortByNewMessages bool
	OnlyFondSoutien   bool
}

type SearchDossierOut struct {
	Dossiers []DossierHeader // passing the query
	Total    int             // all dossiers in the DB, not just passing the query
}

// DossiersSearch returns a list of [Dossier] headers
// matching the given query, sorted by activity time (defined by the messages)
func (ct *Controller) DossiersSearch(c echo.Context) error {
	_, isFondSoutien := JWTUser(c)
	var args SearchDossierIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := ct.searchDossiers(args, isFondSoutien)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type DossierHeader struct {
	Id           ds.IdDossier
	Responsable  string
	Participants string
	NewMessages  int
}

func newDossierHeader(dossier logic.Dossier, isFondSoutien bool) DossierHeader {
	return DossierHeader{
		Id:           dossier.Dossier.Id,
		Responsable:  dossier.Responsable().PrenomNOM(),
		Participants: dossier.ParticipantsLabels(),
		NewMessages:  len(dossier.Events.UnreadMessagesFor(isFondSoutien)),
	}
}

const byIdPrefix = "id:"

// isIdQuery tries for the special Virement label or an Id pattern
func isIdQuery(pattern string) (ds.IdDossier, bool) {
	pattern = strings.TrimSpace(pattern)
	idDossier, isVirementLabel := OffuscateurVirements.Unmask(pattern)
	if isVirementLabel {
		return idDossier, true
	}
	if after, ok := strings.CutPrefix(strings.ToLower(pattern), byIdPrefix); ok {
		id, err := strconv.ParseInt(after, 10, 64)
		if err == nil {
			return ds.IdDossier(id), true
		}
	}
	return 0, false
}

func loadAndFilter(db ds.DB, query SearchDossierIn) ([]logic.DossierFinance, int, error) {
	dossiers, err := ds.SelectAllDossiers(db)
	if err != nil {
		return nil, 0, utils.SQLError(err)
	}
	allDossiersCount := len(dossiers)

	if idDossier, isId := isIdQuery(query.Pattern); isId {
		if dossier, has := dossiers[idDossier]; has {
			dossiers = ds.Dossiers{idDossier: dossier}
			query = SearchDossierIn{} // reset the query to force the match
		}
	}

	ids := dossiers.IDs()

	data, err := logic.LoadDossiersFinances(db, ids...)
	if err != nil {
		return nil, 0, err
	}

	queryText := search.NewQuery(query.Pattern)

	var filtered []logic.DossierFinance
	for _, id := range ids {
		dossier := data.For(id)
		if match(dossier, queryText, query.IdCamp, query.Attente, query.Reglement, query.OnlyFondSoutien) {
			filtered = append(filtered, dossier)
		}
	}

	return filtered, allDossiersCount, nil
}

func (ct *Controller) searchDossiers(query SearchDossierIn, isFondSoutien bool) (SearchDossierOut, error) {
	filtered, totalCount, err := loadAndFilter(ct.db, query)
	if err != nil {
		return SearchDossierOut{}, err
	}

	// sort by messages time
	slices.SortFunc(filtered, func(a, b logic.DossierFinance) int { return a.LastEventTime().Compare(b.LastEventTime()) })

	if query.SortByNewMessages {
		slices.SortStableFunc(filtered, func(a, b logic.DossierFinance) int {
			return len(b.Events.UnreadMessagesFor(isFondSoutien)) - len(a.Events.UnreadMessagesFor(isFondSoutien))
		})
	}

	// paginate and return the headers only
	const maxCount = 40
	if len(filtered) > maxCount {
		filtered = filtered[:maxCount]
	}
	out := make([]DossierHeader, len(filtered))
	for i, v := range filtered {
		out[i] = newDossierHeader(v.Dossier, isFondSoutien)
	}
	return SearchDossierOut{out, totalCount}, nil
}

func match(dossier logic.DossierFinance,
	text search.Query, idCamp evs.OptIdCamp, attente QueryAttente, reglement QueryReglement,
	onlyFondSoutien bool,
) bool {
	// critère fonds de soutien
	if onlyFondSoutien {
		if !dossier.Dossier.Dossier.DemandeFondSoutien {
			return false
		}
	}

	// critère camp
	if idCamp.Valid {
		_, hasCamp := dossier.Camps()[idCamp.Id]
		if !hasCamp {
			return false
		}
	}

	// critère texte
	matchText := slices.ContainsFunc(dossier.Personnes(), text.Match)
	if !matchText {
		return false
	}

	// critère liste d'attente
	if attente != EmptyQA {
		var (
			hasAtLeastOneAttente, hasAtLeastOneInscrit = false, false
			hasAllAttente                              = true
		)
		for _, part := range dossier.Participants {
			// ignore les participants en dehors du séjour sélectionné
			if idCamp.Valid && idCamp.Id != part.IdCamp {
				continue
			}
			if part.Statut == cps.Inscrit {
				hasAtLeastOneInscrit = true
				hasAllAttente = false
			} else {
				hasAtLeastOneAttente = true
			}
		}
		switch attente {
		case AvecAttente:
			if !hasAtLeastOneAttente {
				return false
			}
		case AvecInscrits:
			if !hasAtLeastOneInscrit {
				return false
			}
		case AvecAttenteOnly:
			if !hasAllAttente {
				return false
			}
		}
	}

	// critère financier
	if reglement != EmptyQR {
		statut := dossier.Bilan().StatutPaiement()
		if !reglement.match(statut) {
			return false
		}
	}

	// we have a match !
	return true
}

func (ct *Controller) DossiersLoad(c echo.Context) error {
	id, err := utils.QueryParamInt[ds.IdDossier](c, "id")
	if err != nil {
		return err
	}
	out, err := ct.loadDossier(c.Request().Host, id)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

type DossierDetails struct {
	Dossier        logic.DossierExt
	EspacepersoURL string
	VirementCode   string
	// also displayed in espace perso
	// name, IBAN
	BankAccounts [][2]string
}

func (ct *Controller) loadDossier(host string, id ds.IdDossier) (DossierDetails, error) {
	dossier, err := logic.LoadDossiersFinance(ct.db, id)
	if err != nil {
		return DossierDetails{}, err
	}
	url := logic.EspacePersoURL(ct.key, host, id)
	virement := OffuscateurVirements.Mask(id)
	accounts := ct.asso.BankAccounts()
	return DossierDetails{dossier.Publish(ct.key), url, virement, accounts}, nil
}

func (ct *Controller) DossiersCreate(c echo.Context) error {
	idR, err := utils.QueryParamInt[pr.IdPersonne](c, "idResponsable")
	if err != nil {
		return err
	}
	out, err := ct.createDossier(idR)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) createDossier(idResponsable pr.IdPersonne) (DossierHeader, error) {
	responsable, err := pr.SelectPersonne(ct.db, idResponsable)
	if err != nil {
		return DossierHeader{}, err
	}
	dossier, err := ds.Dossier{
		IdTaux: 1, IdResponsable: responsable.Id, MomentInscription: time.Now().Truncate(time.Second),
		IsValidated: true, PartageAdressesOK: true,
	}.Insert(ct.db)
	if err != nil {
		return DossierHeader{}, err
	}
	return DossierHeader{Id: dossier.Id, Responsable: responsable.PrenomNOM()}, nil
}

func (ct *Controller) DossiersUpdate(c echo.Context) error {
	var args ds.Dossier
	if err := c.Bind(&args); err != nil {
		return err
	}
	type DossiersUpdateOut struct {
		Responsable string
	}
	out, err := ct.updateDossier(args)
	if err != nil {
		return err
	}
	return c.JSON(200, DossiersUpdateOut{out})
}

// returns the Responsable
func (ct *Controller) updateDossier(args ds.Dossier) (string, error) {
	current, err := ds.SelectDossier(ct.db, args.Id)
	if err != nil {
		return "", utils.SQLError(err)
	}
	responsable, err := pr.SelectPersonne(ct.db, args.IdResponsable)
	if err != nil {
		return "", utils.SQLError(err)
	}

	current.IdResponsable = args.IdResponsable
	current.CopiesMails = args.CopiesMails
	current.PartageAdressesOK = args.PartageAdressesOK
	current.DemandeFondSoutien = args.DemandeFondSoutien
	_, err = current.Update(ct.db)
	if err != nil {
		return "", utils.SQLError(err)
	}

	return responsable.PrenomNOM(), nil
}

func (ct *Controller) DossiersDelete(c echo.Context) error {
	id, err := utils.QueryParamInt[ds.IdDossier](c, "id")
	if err != nil {
		return err
	}
	err = ct.deleteDossier(id)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) deleteDossier(id ds.IdDossier) error {
	data, err := logic.LoadDossier(ct.db, id)
	if err != nil {
		return err
	}
	// garbage collect temporary personnes
	toDelete := make(utils.Set[pr.IdPersonne])
	for _, pers := range data.Personnes() {
		if pers.IsTemp {
			toDelete.Add(pers.Id)
		}
	}
	// also cleanup aides files; the other items will cascade
	aides, err := cps.SelectAidesByIdParticipants(ct.db, data.Participants.IDs()...)
	if err != nil {
		return utils.SQLError(err)
	}
	links, err := files.SelectFileAidesByIdAides(ct.db, aides.IDs()...)
	if err != nil {
		return utils.SQLError(err)
	}

	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		deleted, err := files.DeleteFilesByIDs(tx, links.IdFiles()...)
		if err != nil {
			return err
		}
		err = ct.files.Delete(deleted...)
		if err != nil {
			return err
		}

		_, err = ds.DeleteDossierById(tx, id)
		if err != nil {
			return err
		}
		for id := range toDelete {
			_, err = pr.DeletePersonneById(tx, id)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// gestion des aides

func (ct *Controller) StructureaidesGet(c echo.Context) error {
	out, err := cps.SelectAllStructureaides(ct.db)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) StructureaideCreate(c echo.Context) error {
	out, err := cps.Structureaide{}.Insert(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return c.JSON(200, out)
}

func (ct *Controller) StructureaideUpdate(c echo.Context) error {
	var args cps.Structureaide
	if err := c.Bind(&args); err != nil {
		return err
	}
	_, err := args.Update(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return c.NoContent(200)
}

// return an error if aide are already declared
func (ct *Controller) StructureaideDelete(c echo.Context) error {
	id, err := utils.QueryParamInt[cps.IdStructureaide](c, "id")
	if err != nil {
		return err
	}
	err = ct.deleteStructureaide(id)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) deleteStructureaide(id cps.IdStructureaide) error {
	aides, err := cps.SelectAidesByIdStructureaides(ct.db, id)
	if err != nil {
		return utils.SQLError(err)
	}
	if len(aides) != 0 {
		return errors.New("Des aides sont encore rattachées à cette structure.")
	}
	_, err = cps.DeleteStructureaideById(ct.db, id)
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}

type AidesCreateIn struct {
	IdParticipant cps.IdParticipant
	IdStructure   cps.IdStructureaide
}

func (ct *Controller) AidesCreate(c echo.Context) error {
	var args AidesCreateIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := ct.createAide(args)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) createAide(args AidesCreateIn) (cps.Aide, error) {
	// Considère l'aide valide car venant du backoffice
	out, err := cps.Aide{IdParticipant: args.IdParticipant, IdStructureaide: args.IdStructure, Valide: true}.Insert(ct.db)
	if err != nil {
		return out, utils.SQLError(err)
	}
	return out, nil
}

func (ct *Controller) AidesUpdate(c echo.Context) error {
	var args cps.Aide
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.updateAide(args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) updateAide(args cps.Aide) error {
	current, err := cps.SelectAide(ct.db, args.Id)
	if err != nil {
		return utils.SQLError(err)
	}

	participant, err := cps.SelectParticipant(ct.db, current.IdParticipant)
	if err != nil {
		return utils.SQLError(err)
	}
	if err := checkCurrency(ct.db, participant.IdTaux, args.Valeur.Currency); err != nil {
		return err
	}

	current.IdStructureaide = args.IdStructureaide
	current.Valide = args.Valide
	current.Valeur = args.Valeur
	current.ParJour = args.ParJour
	current.NbJoursMax = args.NbJoursMax
	_, err = current.Update(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}

func (ct *Controller) AidesJustificatifUpload(c echo.Context) error {
	id, err := utils.QueryParamInt[cps.IdAide](c, "idAide")
	if err != nil {
		return err
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return err
	}
	content, name, err := filesAPI.ReadUpload(fileHeader)
	if err != nil {
		return err
	}
	out, err := ct.uploadAideJustificatif(id, content, name)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

// uploadJustificatif create or update the link table
func (ct *Controller) uploadAideJustificatif(idAide cps.IdAide, content []byte, filename string) (logic.PublicFile, error) {
	item, found, err := fs.SelectFileAideByIdAide(ct.db, idAide)
	if err != nil {
		return logic.PublicFile{}, utils.SQLError(err)
	}
	idFile := item.IdFile

	var out logic.PublicFile
	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		if !found { // create one file and a link
			file, err := fs.File{}.Insert(tx)
			if err != nil {
				return err
			}
			err = fs.FileAide{IdFile: file.Id, IdAide: idAide}.Insert(tx)
			if err != nil {
				return err
			}
			idFile = file.Id
		}
		file, err := fs.UploadFile(ct.files, tx, idFile, content, filename)
		if err != nil {
			return err
		}
		out = logic.NewPublicFile(ct.key, file)
		return nil
	})

	return out, err
}

func (ct *Controller) AidesJustificatifDelete(c echo.Context) error {
	id, err := utils.QueryParamInt[cps.IdAide](c, "idAide")
	if err != nil {
		return err
	}
	err = ct.deleteAideJustificatif(id)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) deleteAideJustificatif(id cps.IdAide) error {
	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		links, err := fs.DeleteFileAidesByIdAides(tx, id)
		if err != nil {
			return err
		}
		deleted, err := fs.DeleteFilesByIDs(tx, links.IdFiles()...)
		if err != nil {
			return err
		}
		err = ct.files.Delete(deleted...)
		if err != nil {
			return err
		}

		return nil
	})
}

func (ct *Controller) AidesDelete(c echo.Context) error {
	id, err := utils.QueryParamInt[cps.IdAide](c, "id")
	if err != nil {
		return err
	}
	err = ct.deleteAide(id)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

// returns the dossier the [Aide] was linked
func (ct *Controller) deleteAide(id cps.IdAide) error {
	var files []fs.IdFile
	err := utils.InTx(ct.db, func(tx *sql.Tx) error {
		// remove associated documents
		links, err := fs.DeleteFileAidesByIdAides(tx, id)
		if err != nil {
			return err
		}
		files, err = fs.DeleteFilesByIDs(tx, links.IdFiles()...)
		if err != nil {
			return err
		}
		_, err = cps.DeleteAideById(tx, id)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = ct.files.Delete(files...)
	if err != nil {
		return err
	}
	return nil
}

func (ct *Controller) PaiementsCreate(c echo.Context) error {
	idDossier, err := utils.QueryParamInt[ds.IdDossier](c, "idDossier")
	if err != nil {
		return err
	}
	out, err := ct.createPaiement(idDossier)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) createPaiement(idDossier ds.IdDossier) (ds.Paiement, error) {
	// by default, fill with the responsable
	_, personne, err := dossierAndResp(ct.db, idDossier)
	if err != nil {
		return ds.Paiement{}, err
	}

	out, err := ds.Paiement{
		IdDossier: idDossier,
		Time:      time.Now().Truncate(time.Second),
		Mode:      ds.Cheque,
		Payeur:    personne.NOMPrenom(),
	}.Insert(ct.db)
	if err != nil {
		return out, utils.SQLError(err)
	}
	return out, nil
}

func (ct *Controller) PaiementsUpdate(c echo.Context) error {
	var args ds.Paiement
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.updatePaiement(args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

// ensure the currency is supported by the taux
func checkCurrency(db ds.DB, idTaux ds.IdTaux, currency ds.Currency) error {
	taux, err := ds.SelectTaux(db, idTaux)
	if err != nil {
		return utils.SQLError(err)
	}
	if !taux.Has(currency) {
		return errors.New("Montant.Currency unsupported by Taux")
	}
	return nil
}

func (ct *Controller) updatePaiement(args ds.Paiement) error {
	current, err := ds.SelectPaiement(ct.db, args.Id)
	if err != nil {
		return utils.SQLError(err)
	}

	// check that currency is acceptable
	dossier, err := ds.SelectDossier(ct.db, current.IdDossier)
	if err != nil {
		return utils.SQLError(err)
	}
	if err := checkCurrency(ct.db, dossier.IdTaux, args.Montant.Currency); err != nil {
		return err
	}

	// enforce private fields
	args.IdDossier = current.IdDossier
	_, err = args.Update(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}

func (ct *Controller) PaiementsDelete(c echo.Context) error {
	id, err := utils.QueryParamInt[ds.IdPaiement](c, "id")
	if err != nil {
		return err
	}
	_, err = ds.DeletePaiementById(ct.db, id)
	if err != nil {
		return utils.SQLError(err)
	}
	return c.NoContent(200)
}

type DossiersMergeIn struct {
	From    ds.IdDossier // dossier à fusionner
	To      ds.IdDossier // destination
	Notifie bool         // si oui, notifie par mail du changement d'espace perso
}

// DossiersMerge redirige les participants, paiements et messages
// d'un dossier vers un autre, avant de supprimer le dossier
// maintenant vide.
// Un mail de notification au responsable du dossier supprimé peut être envoyé.
func (ct *Controller) DossiersMerge(c echo.Context) error {
	var args DossiersMergeIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.mergeDossier(c.Request().Host, args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) mergeDossier(host string, args DossiersMergeIn) error {
	from, err := ds.SelectDossier(ct.db, args.From)
	if err != nil {
		return utils.SQLError(err)
	}
	fromResp, err := pr.SelectPersonne(ct.db, from.IdResponsable)
	if err != nil {
		return utils.SQLError(err)
	}

	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		err = ds.SwitchPaiementDossier(tx, args.To, from.Id)
		if err != nil {
			return err
		}
		err = evs.SwitchEventMessageDossier(tx, args.To, from.Id)
		if err != nil {
			return err
		}
		err = cps.SwitchParticipantDossier(tx, args.To, from.Id)
		if err != nil {
			return err
		}
		// now delete the empty dossier
		_, err = ds.DeleteDossierById(tx, from.Id)
		if err != nil {
			return err
		}

		if args.Notifie {
			url := logic.EspacePersoURL(ct.key, host, args.To)
			html, err := mails.NotifieFusionDossier(ct.asso, mails.NewContact(&fromResp), url)
			if err != nil {
				return err
			}
			err = mails.NewMailer(ct.smtp, ct.asso.MailsSettings).SendMail(fromResp.Mail,
				"Fusion de dossier", html, from.CopiesMails, nil)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

type PreviewRelance struct {
	Id               ds.IdDossier
	Responsable      string
	Bilan            logic.BilanFinancesPub
	LastEventFacture time.Time // maybe empty
}

func (ct *Controller) EventsSendRelancePaiementPreview(c echo.Context) error {
	idCamp, err := utils.QueryParamInt[cps.IdCamp](c, "idCamp")
	if err != nil {
		return err
	}
	out, err := ct.previewRelancePaiement(idCamp)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) previewRelancePaiement(idCamp cps.IdCamp) ([]PreviewRelance, error) {
	filtered, _, err := loadAndFilter(ct.db, SearchDossierIn{
		IdCamp:    idCamp.Opt(),
		Reglement: Partiel | Zero,
	})
	if err != nil {
		return nil, err
	}

	out := make([]PreviewRelance, len(filtered))
	for i, dossier := range filtered {

		dossier.LastEventTime()
		out[i] = PreviewRelance{
			dossier.Dossier.Dossier.Id,
			dossier.Responsable().NOMPrenom(),
			dossier.Publish(ct.key).Bilan,
			dossier.Events.LastBy(evs.Facture).Created,
		}
	}
	return out, nil
}
