package directeurs

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"registro/controllers/files"
	"registro/crypto"
	"registro/logic"
	"registro/mails"
	cps "registro/sql/camps"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/sql/shared"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

const (
	EndpointDirecteur = "/directeurs"
	EndpointEquipier  = "/equipier"
)

func (ct *Controller) EquipiersGet(c echo.Context) error {
	user := JWTUser(c)

	out, err := ct.getEquipiers(c.Request().Host, user)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

func equipierURL(key crypto.Encrypter, host string, id cps.IdEquipier) string {
	token := crypto.EncryptID(key, id)
	return utils.BuildUrl(host, EndpointEquipier, utils.QP("token", token))
}

type EquipierExt struct {
	Equipier              cps.Equipier
	Personne              string
	HasBirthday           bool
	IsAnimateur, IsAuPair bool
	FormURL               string
}

func newEquipierExt(key crypto.Encrypter, host string, camp cps.Camp, equipier cps.Equipier, personne pr.Personne) EquipierExt {
	return EquipierExt{
		equipier, personne.NOMPrenom(), camp.Plage().HasBirthday(personne.DateNaissance),
		equipier.Roles.Is(cps.Animation) || equipier.Roles.Is(cps.AideAnimation),
		equipier.Roles.IsAuPair(),
		equipierURL(key, host, equipier.Id),
	}
}

func (ct *Controller) getEquipiers(host string, user cps.IdCamp) ([]EquipierExt, error) {
	camp, err := cps.SelectCamp(ct.db, user)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	equipiers, err := cps.SelectEquipiersByIdCamps(ct.db, user)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	personnes, err := pr.SelectPersonnes(ct.db, equipiers.IdPersonnes()...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	out := make([]EquipierExt, 0, len(equipiers))
	for _, equipier := range equipiers {
		out = append(out, newEquipierExt(ct.key, host, camp, equipier, personnes[equipier.IdPersonne]))
	}

	return out, nil
}

type EquipiersCreateIn struct {
	CreatePersonne bool

	Nom, Prenom, Mail string        // valid if CreatePersonne is true
	IdPersonne        pr.IdPersonne // valid if CreatePersonne if false

	Roles cps.Roles
}

func (ct *Controller) EquipiersCreate(c echo.Context) error {
	user := JWTUser(c)

	var args EquipiersCreateIn
	if err := c.Bind(&args); err != nil {
		return err
	}

	out, err := ct.createEquipier(c.Request().Host, args, user)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

func (ct *Controller) createEquipier(host string, args EquipiersCreateIn, user cps.IdCamp) (EquipierExt, error) {
	camp, err := cps.SelectCamp(ct.db, user)
	if err != nil {
		return EquipierExt{}, utils.SQLError(err)
	}
	// check Directeur unicity : this avoid cryptic error messages
	equipiers, err := cps.SelectEquipiersByIdCamps(ct.db, user)
	if err != nil {
		return EquipierExt{}, utils.SQLError(err)
	}
	if _, hasDirecteur := equipiers.Directeur(); args.Roles.Is(cps.Direction) && hasDirecteur {
		return EquipierExt{}, errors.New("Le séjour a déjà un directeur.")
	}
	// also check the personne is not already in this camp
	if _, is := equipiers.ByIdPersonne()[args.IdPersonne]; !args.CreatePersonne && is {
		return EquipierExt{}, errors.New("Ce profil est déjà dans la liste des équipiers du séjour.")
	}

	var (
		equipier cps.Equipier
		personne pr.Personne
	)
	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		// two modes : create or link Personne
		if args.CreatePersonne {
			// we do not mark as tmp since it would prevent document uploading
			pe, err := pr.Personne{Etatcivil: pr.Etatcivil{
				Nom: args.Nom, Prenom: args.Prenom, Mail: args.Mail,
			}}.Insert(tx)
			if err != nil {
				return err
			}
			personne = pe
		} else {
			personne, err = pr.SelectPersonne(tx, args.IdPersonne)
			if err != nil {
				return err
			}
		}
		equipier, err = cps.Equipier{IdPersonne: personne.Id, IdCamp: user, Roles: args.Roles}.Insert(tx)
		if err != nil {
			return err
		}

		demandes := ct.builtins.Defaut(equipier)
		err = fs.InsertManyDemandeEquipiers(tx, demandes...)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return EquipierExt{}, err
	}
	return newEquipierExt(ct.key, host, camp, equipier, personne), nil
}

// EquipiersInviteIn encodes 2 alternatives:
//   - only one equipier
//   - everyone, except the ones having already answered
type EquipiersInviteIn struct {
	OnlyOne shared.OptID[cps.IdEquipier]
}

// EquipiersInvite invite un ou plusieurs équipiers
// à remplir le formulaire, par email.
func (ct *Controller) EquipiersInvite(c echo.Context) error {
	user := JWTUser(c)

	var args EquipiersInviteIn
	if err := c.Bind(&args); err != nil {
		return err
	}

	err := ct.inviteEquipiers(c.Request().Host, args, user)
	if err != nil {
		return err
	}

	out, err := ct.getEquipiers(c.Request().Host, user)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

func (ct *Controller) inviteEquipiers(host string, args EquipiersInviteIn, user cps.IdCamp) error {
	camp, err := cps.SelectCamp(ct.db, user)
	if err != nil {
		return utils.SQLError(err)
	}
	equipiers, err := cps.SelectEquipiersByIdCamps(ct.db, user)
	if err != nil {
		return utils.SQLError(err)
	}
	personnes, err := pr.SelectPersonnes(ct.db, equipiers.IdPersonnes()...)
	if err != nil {
		return utils.SQLError(err)
	}

	var (
		replyTo   mails.ReplyTo
		directeur string
	)
	if eq, ok := equipiers.Directeur(); ok {
		per := personnes[eq.IdPersonne]
		replyTo = mails.CustomReplyTo(per.Mail)
		directeur = per.FPrenom()
	}

	// select equipiers
	var ids []cps.IdEquipier
	if args.OnlyOne.Valid {
		ids = []cps.IdEquipier{args.OnlyOne.Id}
	} else {
		for _, equipier := range equipiers {
			if equipier.FormStatus != cps.Answered {
				ids = append(ids, equipier.Id)
			}
		}
	}

	pool, err := mails.NewPool(ct.smtp, ct.asso.MailsSettings, nil)
	if err != nil {
		return err
	}
	defer pool.Close()

	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		// update DB ...
		for _, id := range ids {
			equipier := equipiers[id]
			equipier.FormStatus = cps.Pending
			_, err = equipier.Update(tx)
			if err != nil {
				return err
			}
		}
		// ... and notify
		for _, id := range ids {
			equipier := equipiers[id]
			personne := personnes[equipier.IdPersonne]
			url := equipierURL(ct.key, host, id)

			html, err := mails.InviteEquipier(ct.asso, camp.Label(), directeur, personne.Etatcivil, url)
			if err != nil {
				return err
			}
			err = pool.SendMail(personne.Mail, fmt.Sprintf("Equipier %s", camp.Label()),
				html, nil, replyTo)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (ct *Controller) EquipiersDelete(c echo.Context) error {
	user := JWTUser(c)

	id, err := utils.QueryParamInt[cps.IdEquipier](c, "id")
	if err != nil {
		return err
	}

	err = ct.deleteEquipier(id, user)
	if err != nil {
		return err
	}

	return c.NoContent(200)
}

func (ct *Controller) deleteEquipier(id cps.IdEquipier, user cps.IdCamp) error {
	equipier, err := cps.SelectEquipier(ct.db, id)
	if err != nil {
		return utils.SQLError(err)
	}
	if equipier.IdCamp != user {
		return errors.New("access forbidden")
	}

	// Demandes will cascade
	_, err = cps.DeleteEquipierById(ct.db, id)
	if err != nil {
		return utils.SQLError(err)
	}
	return nil
}

type DemandeState uint8

const (
	NonDemande  DemandeState = iota // -
	Optionnelle                     // En option
	Obligatoire                     // Requis
)

type EquipierDemande struct {
	Key DemandeKey

	State DemandeState
	Files []logic.PublicFile
}

type DemandeKey struct {
	IdEquipier cps.IdEquipier
	IdDemande  fs.IdDemande
}

type DemandesOut struct {
	Demandes  []fs.Demande // all possible, builtins, sorted
	Equipiers []EquipierDemande
}

func (ct *Controller) EquipiersDemandesGet(c echo.Context) error {
	user := JWTUser(c)

	out, err := ct.getDemandesEquipiers(user)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

func (ct *Controller) getDemandesEquipiers(user cps.IdCamp) (DemandesOut, error) {
	equipiers, err := cps.SelectEquipiersByIdCamps(ct.db, user)
	if err != nil {
		return DemandesOut{}, utils.SQLError(err)
	}
	tmp, err := fs.SelectFilePersonnesByIdPersonnes(ct.db, equipiers.IdPersonnes()...)
	if err != nil {
		return DemandesOut{}, utils.SQLError(err)
	}
	filesByPersonne := tmp.ByIdPersonne()
	allFiles, err := fs.SelectFiles(ct.db, tmp.IdFiles()...)
	if err != nil {
		return DemandesOut{}, utils.SQLError(err)
	}

	currentDemandes, err := fs.SelectDemandeEquipiersByIdEquipiers(ct.db, equipiers.IDs()...)
	if err != nil {
		return DemandesOut{}, utils.SQLError(err)
	}

	states := make(map[DemandeKey]DemandeState)
	// fill the demande first ...
	for _, link := range currentDemandes {
		key := DemandeKey{link.IdEquipier, link.IdDemande}
		state := Obligatoire
		if link.Optionnelle {
			state = Optionnelle
		}
		states[key] = state
	}
	// ... then the current files
	demandes := ct.builtins.List()
	var list []EquipierDemande
	for _, equipier := range equipiers {
		current := filesByPersonne[equipier.IdPersonne].ByIdDemande()
		for _, demande := range demandes {
			// publish files
			links := current[demande.Id]
			publicFiles := make([]logic.PublicFile, len(links))
			for i, link := range links {
				publicFiles[i] = logic.NewPublicFile(ct.key, allFiles[link.IdFile])
			}

			key := DemandeKey{equipier.Id, demande.Id}
			list = append(list, EquipierDemande{key, states[key], publicFiles})
		}
	}

	return DemandesOut{demandes, list}, nil
}

type EquipiersDemandeSetIn struct {
	DemandeKey
	State DemandeState
}

func (ct *Controller) EquipiersDemandeSet(c echo.Context) error {
	user := JWTUser(c)

	var args EquipiersDemandeSetIn
	if err := c.Bind(&args); err != nil {
		return err
	}

	err := ct.setDemandeEquipier(args, user)
	if err != nil {
		return err
	}

	return c.NoContent(200)
}

func (ct *Controller) setDemandeEquipier(args EquipiersDemandeSetIn, user cps.IdCamp) error {
	equipier, err := cps.SelectEquipier(ct.db, args.IdEquipier)
	if err != nil {
		return utils.SQLError(err)
	}
	if equipier.IdCamp != user {
		return errors.New("access forbidden")
	}

	return utils.InTx(ct.db, func(tx *sql.Tx) error {
		_, err = fs.DeleteDemandeEquipiersByIdEquipierAndIdDemande(tx, args.IdEquipier, args.IdDemande)
		if err != nil {
			return err
		}
		// if required, add the Demande back
		if args.State != NonDemande {
			err = fs.DemandeEquipier{
				IdEquipier:  args.IdEquipier,
				IdDemande:   args.IdDemande,
				Optionnelle: args.State == Optionnelle,
			}.Insert(tx)
		}
		if err != nil {
			return err
		}
		return nil
	})
}

// EquipiersDownloadFiles renvoie une archive contenant
// tous les documents demandés aux equipiers, en "stremant" la réponse.
func (ct *Controller) EquipiersDownloadFiles(c echo.Context) error {
	user := JWTUser(c)
	return ct.streamFilesEquipiers(user, c.Response())
}

type fileAndPrefix struct {
	id       fs.IdFile
	fullName string // including metadata from demande and equipier
}

func (ct *Controller) compileFilesEquipiers(user cps.IdCamp) ([]fileAndPrefix, error) {
	equipiers, err := cps.SelectEquipiersByIdCamps(ct.db, user)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	personnes, err := pr.SelectPersonnes(ct.db, equipiers.IdPersonnes()...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	tmp1, err := fs.SelectDemandeEquipiersByIdEquipiers(ct.db, equipiers.IDs()...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	demandesByEquipier := tmp1.ByIdEquipier()
	demandes, err := fs.SelectDemandes(ct.db, tmp1.IdDemandes()...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	tmp2, err := fs.SelectFilePersonnesByIdPersonnes(ct.db, personnes.IDs()...)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	filesByPersonne := tmp2.ByIdPersonne()
	files, err := fs.SelectFiles(ct.db, tmp2.IdFiles()...)
	if err != nil {
		return nil, utils.SQLError(err)
	}

	var out []fileAndPrefix
	for _, equipier := range equipiers {
		personne := personnes[equipier.IdPersonne]
		fileLinks := filesByPersonne[equipier.IdPersonne]
		demandeLinks := demandesByEquipier[equipier.Id].ByIdDemande()
		// restrict to [Demande]s
		for _, fileLink := range fileLinks {
			if _, has := demandeLinks[fileLink.IdDemande]; !has {
				continue // ignore this file
			}
			file := files[fileLink.IdFile]
			demande := demandes[fileLink.IdDemande]
			out = append(out, fileAndPrefix{
				id:       fileLink.IdFile,
				fullName: fmt.Sprintf("%s %s %s", demande.Categorie, personne.NOMPrenom(), file.NomClient),
			})
		}
	}

	slices.SortFunc(out, func(a, b fileAndPrefix) int { return strings.Compare(a.fullName, b.fullName) })

	return out, nil
}

func (ct *Controller) streamFilesEquipiers(user cps.IdCamp, resp http.ResponseWriter) error {
	camp, err := cps.SelectCamp(ct.db, user)
	if err != nil {
		return utils.SQLError(err)
	}

	toZip, err := ct.compileFilesEquipiers(user)
	if err != nil {
		return err
	}

	return files.StreamZip(resp, fmt.Sprintf("Documents Equipe %s.zip", camp.Label()), func(yield func(files.ZipItem, error) bool) {
		for _, file := range toZip {
			content, err := ct.files.Load(file.id, false)
			if err != nil {
				yield(files.ZipItem{}, err)
				return
			}
			if !yield(files.ZipItem{Name: file.fullName, Content: content}, nil) {
				return
			}
		}
	})
}
