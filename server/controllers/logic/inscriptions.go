package logic

import (
	"database/sql"
	"errors"
	"slices"
	"strings"

	"registro/controllers/search"
	cps "registro/sql/camps"
	ds "registro/sql/dossiers"
	evs "registro/sql/events"
	"registro/sql/files"
	pr "registro/sql/personnes"
	"registro/utils"
)

type Inscription struct {
	Dossier      ds.Dossier
	Message      string // le message (optionnel) du formulaire d'inscription
	Responsable  pr.Personne
	Participants []cps.ParticipantCamp
	// ValidatedBy stores the camp which have validated
	// this inscription, computed using the participants status.
	// This field is ignored in backoffice, but used in directeurs
	// to handle inscriptions with mixed camps.
	ValidatedBy []cps.IdCamp
}

func newInscription(de Dossier) Inscription {
	var chunks []string
	// collect the messages
	for _, event := range de.Events.By(evs.Message) {
		content := event.Content.(Message).Message
		if content.Origine == evs.FromEspaceperso {
			chunks = append(chunks, content.Contenu)
		}
	}
	message := strings.Join(chunks, "\n\n")

	var validatedBy []cps.IdCamp
	for idCamp, l := range de.Participants.ByIdCamp() {
		validated := true
		for _, p := range l {
			if p.Statut == cps.AStatuer {
				validated = false
				break
			}
		}
		if validated {
			validatedBy = append(validatedBy, idCamp)
		}
	}

	return Inscription{
		Dossier:      de.Dossier,
		Responsable:  de.Responsable(),
		Participants: de.ParticipantsExt(),
		Message:      message,
		ValidatedBy:  validatedBy,
	}
}

// LoadInscriptions sorts by time
func LoadInscriptions(db ds.DB, ids ...ds.IdDossier) ([]Inscription, error) {
	loader, err := LoadDossiers(db, ids...)
	if err != nil {
		return nil, err
	}

	out := make([]Inscription, len(ids))
	for i, id := range ids {
		out[i] = newInscription(loader.For(id))
	}

	// sort by time
	slices.SortFunc(out, func(a, b Inscription) int {
		return a.Dossier.MomentInscription.Compare(b.Dossier.MomentInscription)
	})

	return out, nil
}

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

// PrepareValideInscription renvoie les participants modifiés,
// à persister via une requête SQL.
func (loader Dossier) PrepareValideInscription(db ds.DB) (cps.Participants, error) {
	// on s'assure qu'aucune personne n'est temporaire
	for _, pe := range loader.Personnes() {
		if pe.IsTemp {
			return nil, errors.New("internal error: personne should not be temporary")
		}
	}

	// le status est calculé camp par camp
	dossierByCamp := loader.Participants.ByIdCamp()

	// on calcule le statut des participants (requiert les participants et personnes déjà inscrites)
	camps, err := cps.LoadCamps(db, loader.Camps().IDs()...)
	if err != nil {
		return nil, err
	}

	out := make(cps.Participants)
	for _, camp := range camps {
		incommingPa := utils.MapValues(dossierByCamp[camp.Camp.Id])
		incommingPe := loader.PersonnesFor(incommingPa)

		for index, status := range camp.Status(incommingPe) {
			listeAttente := status.Hint()
			// update the participant
			pa := incommingPa[index]
			pa.Statut = listeAttente
			out[pa.Id] = pa
		}
	}

	return out, err
}
