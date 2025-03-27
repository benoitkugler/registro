package directeurs

import (
	"database/sql"
	"errors"

	"registro/controllers/files"
	"registro/controllers/search"
	"registro/crypto"
	cps "registro/sql/camps"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/utils"

	"github.com/labstack/echo/v4"
)

const EndpointEquipier = "/equipier"

func (ct *Controller) EquipiersGet(c echo.Context) error {
	user := JWTUser(c)

	out, err := ct.getEquipiers(c.Request().Host, user)
	if err != nil {
		return err
	}

	return c.JSON(200, out)
}

type EquipierExt struct {
	Equipier cps.Equipier
	Personne string

	FormURL string
}

func newEquipierExt(key crypto.Encrypter, host string, equipier cps.Equipier, personne pr.Personne) EquipierExt {
	token := crypto.EncryptID(key, equipier.Id)
	url := utils.BuildUrl(host, EndpointEquipier, utils.QP("token", token))
	return EquipierExt{equipier, personne.NOMPrenom(), url}
}

func (ct *Controller) getEquipiers(host string, user cps.IdCamp) ([]EquipierExt, error) {
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
		out = append(out, newEquipierExt(ct.key, host, equipier, personnes[equipier.IdPersonne]))
	}

	return out, nil
}

type EquipiersCreateIn struct {
	CreatePersonne bool
	Personne       search.PatternsSimilarite // valid if CreatePersonne is true
	IdPersonne     pr.IdPersonne             // valid if CreatePersonne if false

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
	// check Directeur unicity : this avoid cryptic error messages
	equipiers, err := cps.SelectEquipiersByIdCamps(ct.db, user)
	if err != nil {
		return EquipierExt{}, utils.SQLError(err)
	}
	if _, hasDirecteur := equipiers.Directeur(); args.Roles.Is(cps.Direction) && hasDirecteur {
		return EquipierExt{}, errors.New("Le séjour a déjà un directeur.")
	}

	var (
		equipier cps.Equipier
		personne pr.Personne
	)
	err = utils.InTx(ct.db, func(tx *sql.Tx) error {
		// two modes : create or link Personne
		if args.CreatePersonne {
			// we do not mark as tmp since it would prevent document uploading
			pe, err := pr.Personne{Etatcivil: args.Personne.Personne()}.Insert(tx)
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
		return nil
	})
	if err != nil {
		return EquipierExt{}, err
	}
	return newEquipierExt(ct.key, host, equipier, personne), nil
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
	Files []files.PublicFile
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
			publicFiles := make([]files.PublicFile, len(links))
			for i, link := range links {
				publicFiles[i] = files.NewPublicFile(ct.key, allFiles[link.IdFile])
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
