package directeurs

import (
	"database/sql"
	"errors"

	"registro/controllers/files"
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
		token := crypto.EncryptID(ct.key, equipier.Id)
		url := utils.BuildUrl(host, EndpointEquipier, utils.QP("token", token))
		personne := personnes[equipier.IdPersonne].NOMPrenom()
		out = append(out, EquipierExt{equipier, personne, url})
	}

	return out, nil
}

type DemandeState uint8

const (
	NonDemande  DemandeState = iota // -
	Optionnelle                     // En option
	Obligatoire                     // Requis
)

type EquipierDemandes struct {
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
	Equipiers []EquipierDemandes
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
	var list []EquipierDemandes
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
			list = append(list, EquipierDemandes{key, states[key], publicFiles})
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
