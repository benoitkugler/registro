package logic

import (
	"database/sql"
	"errors"

	"registro/controllers/search"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	"registro/sql/files"
	pr "registro/sql/personnes"
	"registro/utils"
)

// IdentTarget indique comment identifier une personne temporaire.
//
// Si `Rattache` vaut false, la personne est simplement marquée comme non temporaire.
//
// Sinon, le profil [IdTemporaire] est supprimé et toutes ses occurences sont remplacées
// par [RattacheTo]. [RattacheTo] est mis à jour pour prendre en compte le contenu de [IdTemporaire],
// en utilisant [search.Merge]
type IdentTarget struct {
	IdTemporaire pr.IdPersonne // le profil à rattacher

	Rattache   bool
	RattacheTo pr.IdPersonne // only valid if [Rattache] is true
}

func IdentifiePersonne(db *sql.DB, args IdentTarget) error {
	temporaire, err := pr.SelectPersonne(db, args.IdTemporaire)
	if err != nil {
		return utils.SQLError(err)
	}

	if !args.Rattache {
		// on marque simplement la personne 'entrante' comme non temporaire
		temporaire.IsTemp = false
		_, err = temporaire.Update(db)
		if err != nil {
			return utils.SQLError(err)
		}
		return nil
	}

	if args.IdTemporaire == args.RattacheTo {
		return errors.New("internal error: same target and origin profil")
	}
	err = utils.InTx(db, func(tx *sql.Tx) error {
		existant, err := pr.SelectPersonne(tx, args.RattacheTo)
		if err != nil {
			return err
		}
		if existant.IsTemp {
			return errors.New("internal error: target is temporary")
		}

		// 1) on applique les modifications de la fusion
		existant.Etatcivil, _ = search.Merge(temporaire.Etatcivil, existant.Etatcivil)
		_, err = existant.Update(tx)
		if err != nil {
			return err
		}

		// 2) redirige les occurrences de [IdTemporaire]
		if err = cps.SwitchParticipantPersonne(tx, existant.Id, temporaire.Id); err != nil {
			return err
		}
		if err = cps.SwitchEquipierPersonne(tx, existant.Id, temporaire.Id); err != nil {
			return err
		}
		if err = ds.SwitchDossierPersonne(tx, existant.Id, temporaire.Id); err != nil {
			return err
		}
		if err = files.SwitchDemandePersonne(tx, existant.Id.Opt(), temporaire.Id.Opt()); err != nil {
			return err
		}
		if err = files.SwitchFilePersonnePersonne(tx, existant.Id, temporaire.Id); err != nil {
			return err
		}

		// 3) supprime la personne temporaire désormais inutile
		_, err = pr.DeletePersonneById(tx, temporaire.Id)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}
