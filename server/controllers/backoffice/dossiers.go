package backoffice

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"registro/controllers/espaceperso"
	"registro/controllers/logic"
	"registro/controllers/search"
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

type QueryReglement uint8

const (
	EmptyQR QueryReglement = iota // Indifférent
	Zero                          // Non commencé
	Partiel                       // En cours
	Total                         // Complété
)

// The zero value defaults to returning everything
type SearchDossierIn struct {
	Pattern   string // Responsable et participants
	IdCamp    evs.OptIdCamp
	Attente   QueryAttente
	Reglement QueryReglement
}

type SearchDossierOut struct {
	Dossiers []DossierHeader // passing the query
	Total    int             // all dossiers in the DB, not just passing the query
}

// DossiersSearch returns a list of [Dossier] headers
// matching the given query, sorted by activity time (defined by the messages)
func (ct *Controller) DossiersSearch(c echo.Context) error {
	var args SearchDossierIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := ct.searchDossiers(args)
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

func newDossierHeader(dossier logic.Dossier) DossierHeader {
	personnes := dossier.Personnes()
	// extract participants
	chunks := make([]string, 0, len(personnes)-1)
	for _, pe := range personnes[1:] {
		chunks = append(chunks, pe.PrenomNOM())
	}
	return DossierHeader{
		Id:           dossier.Dossier.Id,
		Responsable:  personnes[0].PrenomNOM(),
		Participants: strings.Join(chunks, ", "),
		NewMessages:  len(dossier.Events.UnreadMessagesForBackoffice()),
	}
}

const sortByMessagesPattern = "sort:messages"

func (ct *Controller) searchDossiers(query SearchDossierIn) (SearchDossierOut, error) {
	var (
		dossiers ds.Dossiers
		err      error
	)
	// try for the special Virement label
	idDossier, isVirementLabel := OffuscateurVirements.Unmask(strings.TrimSpace(query.Pattern))
	if isVirementLabel {
		dossiers, err = ds.SelectDossiers(ct.db, idDossier)
		if err != nil {
			return SearchDossierOut{}, utils.SQLError(err)
		}
		query = SearchDossierIn{} // reset the query to force the match
	} else {
		dossiers, err = ds.SelectAllDossiers(ct.db)
		if err != nil {
			return SearchDossierOut{}, utils.SQLError(err)
		}
		dossiers.RestrictByValidated(true)
	}
	ids := dossiers.IDs()

	data, err := logic.LoadDossiersFinances(ct.db, ids...)
	if err != nil {
		return SearchDossierOut{}, err
	}

	sortByMessages := false
	if query.Pattern == sortByMessagesPattern {
		sortByMessages = true
		query.Pattern = "" // reset the query to force the match
	}
	queryText := search.NewQuery(query.Pattern)

	var filtered []logic.DossierFinance
	for _, id := range ids {
		dossier := data.For(id)
		if match(dossier, queryText, query.IdCamp, query.Attente, query.Reglement) {
			filtered = append(filtered, dossier)
		}
	}

	// sort by messages time
	slices.SortFunc(filtered, func(a, b logic.DossierFinance) int { return a.Time().Compare(b.Time()) })

	if sortByMessages {
		slices.SortStableFunc(filtered, func(a, b logic.DossierFinance) int {
			return len(b.Events.UnreadMessagesForBackoffice()) - len(a.Events.UnreadMessagesForBackoffice())
		})
	}

	// paginate and return the headers only
	const maxCount = 50
	if len(filtered) > maxCount {
		filtered = filtered[:maxCount]
	}
	out := make([]DossierHeader, len(filtered))
	for i, v := range filtered {
		out[i] = newDossierHeader(v.Dossier)
	}
	return SearchDossierOut{out, len(ids)}, nil
}

func match(dossier logic.DossierFinance,
	text search.Query, idCamp evs.OptIdCamp, attente QueryAttente, reglement QueryReglement,
) bool {
	// critère camp
	if idCamp.Valid {
		_, hasCamp := dossier.Camps()[idCamp.Id]
		if !hasCamp {
			return false
		}
	}

	// critère texte
	matchText := false
	for _, personne := range dossier.Personnes() {
		if search.QueryMatch(text, personne) {
			matchText = true
			break
		}
	}
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
			// ignore les participants en dehors du camp sélectionné
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
			return hasAtLeastOneAttente
		case AvecInscrits:
			return hasAtLeastOneInscrit
		case AvecAttenteOnly:
			return hasAllAttente
		}
	}

	// critère financier
	if reglement != EmptyQR {
		matchStatut := dossier.Bilan().StatutPaiement() == logic.StatutPaiement(reglement)
		if !matchStatut {
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
}

// also marks the message as seen
func (ct *Controller) loadDossier(host string, id ds.IdDossier) (DossierDetails, error) {
	dossier, err := logic.LoadDossiersFinance(ct.db, id)
	if err != nil {
		return DossierDetails{}, err
	}
	url := espaceperso.URLEspacePerso(ct.key, host, id)
	virement := OffuscateurVirements.Mask(id)
	return DossierDetails{dossier.Publish(), url, virement}, nil
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

func (ct *Controller) createDossier(idR pr.IdPersonne) (DossierHeader, error) {
	responsable, err := pr.SelectPersonne(ct.db, idR)
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

type ParticipantsCreateIn struct {
	IdDossier  ds.IdDossier
	IdCamp     cps.IdCamp
	IdPersonne pr.IdPersonne
}

// ParticipantsCreate ajoute un participant au séjour donné,
// en résolvant statut et groupe.
func (ct *Controller) ParticipantsCreate(c echo.Context) error {
	var args ParticipantsCreateIn
	if err := c.Bind(&args); err != nil {
		return err
	}
	out, err := ct.createParticipant(args)
	if err != nil {
		return err
	}
	return c.JSON(200, out)
}

func (ct *Controller) createParticipant(args ParticipantsCreateIn) (cps.Participant, error) {
	dossier, err := ds.SelectDossier(ct.db, args.IdDossier)
	if err != nil {
		return cps.Participant{}, utils.SQLError(err)
	}
	personne, err := pr.SelectPersonne(ct.db, args.IdPersonne)
	if err != nil {
		return cps.Participant{}, utils.SQLError(err)
	}

	// better error message if already present
	_, alreadyHere, err := cps.SelectParticipantByIdCampAndIdPersonne(ct.db, args.IdCamp, args.IdPersonne)
	if err != nil {
		return cps.Participant{}, utils.SQLError(err)
	}
	if alreadyHere {
		return cps.Participant{}, fmt.Errorf("Le profil %s est déjà présent sur ce camp !", personne.PrenomNOM())
	}

	loaders, err := cps.LoadCamps(ct.db, args.IdCamp)
	if err != nil {
		return cps.Participant{}, err
	}
	camp := loaders[0]

	// resolve Groupe...
	groupes, err := cps.SelectGroupesByIdCamps(ct.db, args.IdCamp)
	if err != nil {
		return cps.Participant{}, utils.SQLError(err)
	}
	groupe, hasGroupe := groupes.TrouveGroupe(personne.DateNaissance)

	// ... and Statut
	statut := camp.Status([]pr.Personne{personne})[0]
	participant := cps.Participant{
		IdDossier:  args.IdDossier,
		IdCamp:     args.IdCamp,
		IdPersonne: args.IdPersonne,

		IdTaux: dossier.IdTaux,
		Statut: statut.Hint(),
	}

	// if the dossier is empty (for instance if manually created), we want to allow
	// a different taux (the one of the camp) to be used
	existingP, err := cps.SelectParticipantsByIdDossiers(ct.db, args.IdDossier)
	if err != nil {
		return cps.Participant{}, utils.SQLError(err)
	}

	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		if len(existingP) == 0 {
			// update the dossier ...
			dossier.IdTaux = camp.Camp.IdTaux
			_, err = dossier.Update(tx)
			if err != nil {
				return err
			}
			// ... and use this taux
			participant.IdTaux = camp.Camp.IdTaux
		}

		participant, err = participant.Insert(tx)
		if err != nil {
			return err
		}
		if hasGroupe {
			err = cps.GroupeParticipant{IdGroupe: groupe.Id, IdCamp: groupe.IdCamp, IdParticipant: participant.Id}.Insert(tx)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return participant, err
}

// ParticipantsUpdate modifie les champs d'un participant.
//
// Les champs [IdPersonne], [IdDossier], [IdTaux] et [IdCamp] sont ignorés.
//
// Le statut est modifié sans aucune notification.
func (ct *Controller) ParticipantsUpdate(c echo.Context) error {
	var args cps.Participant
	if err := c.Bind(&args); err != nil {
		return err
	}
	err := ct.updateParticipant(args)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) updateParticipant(args cps.Participant) error {
	current, err := cps.SelectParticipant(ct.db, args.Id)
	if err != nil {
		return utils.SQLError(err)
	}
	current.Statut = args.Statut
	current.Remises = args.Remises
	current.QuotientFamilial = args.QuotientFamilial
	current.OptionPrix = args.OptionPrix
	current.Details = args.Details
	current.Bus = args.Bus
	_, err = current.Update(ct.db)
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}

// ParticipantsDelete supprime le participant donné.
// Si la personne liée est temporaire, elle est aussi supprimée.
func (ct *Controller) ParticipantsDelete(c echo.Context) error {
	id, err := utils.QueryParamInt[cps.IdParticipant](c, "id")
	if err != nil {
		return err
	}
	err = ct.deleteParticipant(id)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

func (ct *Controller) deleteParticipant(id cps.IdParticipant) error {
	// cleanup aides files and temp personne; the other items will cascade
	aides, err := cps.SelectAidesByIdParticipants(ct.db, id)
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

		participant, err := cps.DeleteParticipantById(tx, id)
		if err != nil {
			return err
		}
		personne, err := pr.SelectPersonne(tx, participant.IdPersonne)
		if err != nil {
			return err
		}
		if personne.IsTemp { // cleanup
			_, err = pr.DeletePersonneById(tx, personne.Id)
			if err != nil {
				return err
			}
		}

		return nil
	})
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
	content, name, err := utils.ReadUpload(fileHeader)
	if err != nil {
		return err
	}
	err = ct.uploadAideJustificatif(id, content, name)
	if err != nil {
		return err
	}
	return c.NoContent(200)
}

// uploadJustificatif create or update the link table
func (ct *Controller) uploadAideJustificatif(idAide cps.IdAide, content []byte, filename string) error {
	item, found, err := fs.SelectFileAideByIdAide(ct.db, idAide)
	if err != nil {
		return utils.SQLError(err)
	}
	idFile := item.IdFile

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
		_, err = fs.UploadFile(ct.files, tx, idFile, content, filename)
		if err != nil {
			return err
		}
		return nil
	})

	return err
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
	dossier, err := ds.SelectDossier(ct.db, idDossier)
	if err != nil {
		return ds.Paiement{}, err
	}
	personne, err := pr.SelectPersonne(ct.db, dossier.IdResponsable)
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
			url := espaceperso.URLEspacePerso(ct.key, host, args.To)
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
