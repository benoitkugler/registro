package logic

import (
	"database/sql"
	"errors"
	"slices"
	"strings"
	"time"

	"registro/config"
	"registro/crypto"
	"registro/logic/search"
	"registro/mails"
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
}

func newInscription(de Dossier) Inscription {
	var chunks []string
	// collect the messages
	for event := range IterEventsBy[Message](de.Events) {
		content := event.Content.Message
		if content.Origine == evs.Espaceperso {
			chunks = append(chunks, content.Contenu)
		}
	}
	message := strings.Join(chunks, "\n\n")

	return Inscription{
		Dossier:      de.Dossier,
		Responsable:  de.Responsable(),
		Participants: de.ParticipantsExt(),
		Message:      message,
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

type StatutHints = map[cps.IdParticipant]StatutExt

// StatutHints renvoie le statut qu'il faudrait appliquer
// au participant du dossier.
func (dossier Dossier) StatutHints(db ds.DB, bypass StatutBypassRights) (StatutHints, error) {
	// le status est calculé camp par camp
	partsByCamp := dossier.Participants.ByIdCamp()

	// on calcule le statut des participants (requiert les participants et personnes déjà inscrites)
	camps, err := cps.LoadCampsPersonnes(db, dossier.Camps().IDs()...)
	if err != nil {
		return nil, err
	}

	out := make(StatutHints)
	for _, camp := range camps {
		incommingPa := utils.MapValues(partsByCamp[camp.Camp.Id])
		incommingPe := dossier.PersonnesFor(incommingPa)

		for index, status := range camp.Status(incommingPe) {
			pa := incommingPa[index]
			out[pa.Id] = bypass.resolve(status, pa.Statut)
		}
	}

	return out, err
}

// StatutBypassRights grants the rights to validate a participant,
// and override the default (computed) hint.
type StatutBypassRights struct {
	ProfilInvalide bool
	CampComplet    bool
	Inscrit        bool
}

type StatutExt struct {
	Causes cps.StatutCauses
	Statut cps.StatutParticipant

	AllowedChanges []cps.StatutParticipant // empty for readonly
	// if false, no update will be done
	// it is always false for participnt already
	// validated
	Validable bool
}

// IsAllowed returns 'true' if the bypass rights allow the given statut to be
// applied.
func (st StatutExt) IsAllowed(statut cps.StatutParticipant) bool {
	return statut == st.Statut || slices.Contains(st.AllowedChanges, statut)
}

func (bp StatutBypassRights) resolve(st cps.StatutCauses, currentStatut cps.StatutParticipant) StatutExt {
	out := StatutExt{Causes: st, Statut: st.Hint()}
	switch out.Statut {
	case cps.AttenteProfilInvalide:
		if bp.ProfilInvalide {
			out.AllowedChanges = []cps.StatutParticipant{cps.Inscrit}
			out.Validable = true
		}
	case cps.AttenteCampComplet:
		if bp.CampComplet {
			out.AllowedChanges = []cps.StatutParticipant{cps.Inscrit}
			out.Validable = true
		}
	case cps.Inscrit:
		out.Validable = true
		if bp.Inscrit {
			out.AllowedChanges = []cps.StatutParticipant{cps.AttenteProfilInvalide, cps.AttenteCampComplet}
		}
	default: // should not happen
	}
	if currentStatut != cps.AStatuer {
		out.Validable = false
	}
	return out
}

func allValidated(ps cps.Participants) bool {
	for _, part := range ps {
		if part.Statut == cps.AStatuer {
			return false
		}
	}
	return true
}

// InscriptionsValideIn indique le statut des participants
// à appliquer.
type InscriptionsValideIn struct {
	IdDossier ds.IdDossier
	// choosen by the clients
	Statuts  map[cps.IdParticipant]cps.StatutParticipant
	SendMail bool
}

// ValideInscription met à jour le statut des participants et
// envoie un mail d'accusé de réception.
//
// Le dossier est validé si aucun participant n'est encore [AStatuer]
func ValideInscription(db *sql.DB, key crypto.Encrypter, smtp config.SMTP, asso config.Asso,
	host string, args InscriptionsValideIn, bypass StatutBypassRights, idCamp cps.OptIdCamp,
) error {
	loader, err := LoadDossier(db, args.IdDossier)
	if err != nil {
		return err
	}
	dossier := loader.Dossier

	hints, err := loader.StatutHints(db, bypass)
	if err != nil {
		return err
	}

	// on s'assure qu'aucune personne n'est temporaire
	for _, pe := range loader.Personnes() {
		if pe.IsTemp {
			return errors.New("internal error: Personne should not be temporary")
		}
	}

	err = utils.InTx(db, func(tx *sql.Tx) error {
		var inscrits, attente, astatuer []mails.Participant
		for _, pExt := range loader.ParticipantsExt() {
			participant := pExt.Participant
			mPart := mails.Participant{Personne: pExt.Personne.PrenomNOM(), Camp: pExt.Camp.Label()}
			hint := hints[participant.Id]
			// ignore participant not validable (already validated or restricte for directors)
			// or for other camps
			if !hint.Validable || (idCamp.Valid && !idCamp.Is(participant.IdCamp)) {
				if participant.Statut == cps.AStatuer {
					astatuer = append(astatuer, mPart)
				}
				continue
			}

			// check the new status is present and allowed
			newStatut, _ := args.Statuts[participant.Id]
			if newStatut == 0 {
				return errors.New("internal error: missing participant in InscriptionsValideIn.Statuts")
			}
			if !hint.IsAllowed(newStatut) {
				return errors.New("internal error: statut not allowed")
			}

			participant.Statut = newStatut
			_, err = participant.Update(tx)
			if err != nil {
				return err
			}
			// update loader, used below
			loader.Participants[participant.Id] = participant

			if newStatut == cps.Inscrit {
				inscrits = append(inscrits, mPart)
			} else {
				attente = append(attente, mPart)
			}
		}

		if allValidated(loader.Participants) {
			dossier.IsValidated = true
			_, err = dossier.Update(tx)
			if err != nil {
				return err
			}
		}

		// mark the validation ...
		ev, err := evs.Event{IdDossier: dossier.Id, Kind: evs.Validation, Created: time.Now()}.Insert(tx)
		if err != nil {
			return err
		}
		err = evs.EventValidation{IdEvent: ev.Id, IdCamp: idCamp}.Insert(tx)
		if err != nil {
			return err
		}

		// ... and notify if required
		if args.SendMail {
			resp := loader.Responsable()
			url := URLEspacePerso(key, host, dossier.Id, utils.QPInt("idEvent", ev.Id))
			html, err := mails.NotifieValidationInscription(asso, mails.NewContact(&resp), url, inscrits, attente, astatuer)
			if err != nil {
				return err
			}
			err = mails.NewMailer(smtp, asso.MailsSettings).SendMail(resp.Mail, "Inscription reçue", html, dossier.CopiesMails, nil)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}
