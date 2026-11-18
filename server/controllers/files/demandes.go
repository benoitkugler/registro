package files

import (
	"slices"

	"registro/config"
	cps "registro/sql/camps"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/utils"
)

// DemandesLoader resolves the documents that may be asked to participants.
//
// Resolution happens in two phases:
//  1. [LoadDemandes] resolves the demandes from the association settings
//     and the camps (builtins and camp-specific demandes);
//  2. [DemandesLoader.ForParticipant] filters them against each participant
//     (registration status, age and nationnality).
type DemandesLoader struct {
	settings config.AskedDocs

	demandes fs.Demandes // covers all camps
	byCamp   map[cps.IdCamp][]fs.IdDemande
}

// LoadDemandes resolves the demandes relevant to [idsCamps].
// Builtin demandes are always included, except for CarteId which depends
// on the association settings.
func LoadDemandes(db fs.DB, settings config.AskedDocs, idsCamps ...cps.IdCamp) (DemandesLoader, error) {
	links, err := fs.SelectDemandeCampsByIdCamps(db, idsCamps...)
	if err != nil {
		return DemandesLoader{}, utils.SQLError(err)
	}
	byLinks := links.ByIdCamp()
	custom, err := fs.SelectDemandes(db, links.IdDemandes()...)
	if err != nil {
		return DemandesLoader{}, utils.SQLError(err)
	}
	vaccin, carteId, err := DemandesVaccinEtCarteId(db)
	if err != nil {
		return DemandesLoader{}, err
	}
	custom[vaccin.Id] = vaccin
	custom[carteId.Id] = carteId

	out := DemandesLoader{
		settings: settings,
		demandes: custom,
		byCamp:   make(map[cps.IdCamp][]fs.IdDemande, len(idsCamps)),
	}

	for _, idCamp := range idsCamps {
		ids := []fs.IdDemande{vaccin.Id}
		if settings.CarteId {
			ids = append(ids, carteId.Id)
		}
		ids = append(ids, byLinks[idCamp].IdDemandes()...)
		out.byCamp[idCamp] = ids
	}
	return out, nil
}

// DemandesForCamp returns the demandes resolved for [idCamp].
func (l DemandesLoader) DemandesForCamp(idCamp cps.IdCamp) fs.Demandes {
	ids := l.byCamp[idCamp]
	out := make(fs.Demandes, len(ids))
	for _, id := range ids {
		out[id] = l.demandes[id]
	}
	return out
}

// AllDemandes returns every resolved demande, indexed by ID.
func (l DemandesLoader) AllDemandes() fs.Demandes { return l.demandes }

// DemandesParticipant is the subset of demandes required from one participant.
type DemandesParticipant struct {
	// FicheSanitaire is true when a fiche sanitaire (and vaccins) is required.
	FicheSanitaire bool
	// Demandes preserves the resolution order: vaccins, CarteId, camp demandes
	Demandes []fs.Demande
}

// Has reports whether [id] is required from the participant.
func (dp DemandesParticipant) Has(id fs.IdDemande) bool {
	return slices.ContainsFunc(dp.Demandes, func(demande fs.Demande) bool { return demande.Id == id })
}

// ParticipantDemandesArgs describes a participant to filter demandes for.
// AgeCamp is the camp used by the age rule: the camp itself for directors,
// but the participant's first camp in espaceperso.
type ParticipantDemandesArgs struct {
	IdCamp   cps.IdCamp
	AgeCamp  cps.Camp
	Personne pr.Personne
	Statut   cps.StatutParticipant
}

// ForParticipant filters the resolved demandes against one participant.
func (l DemandesLoader) ForParticipant(args ParticipantDemandesArgs) DemandesParticipant {
	if args.Personne.IsTemp || args.Statut != cps.Inscrit {
		return DemandesParticipant{}
	}
	ficheSanitaire := args.AgeCamp.AgeDebutCamp(args.Personne.DateNaissance) < 18 || l.settings.FicheSanitaireMajeurs
	out := DemandesParticipant{FicheSanitaire: ficheSanitaire}
	for _, id := range l.byCamp[args.IdCamp] {
		demande := l.demandes[id]
		if !demande.Nationnalite.Match(args.Personne.Nationnalite) {
			continue
		}
		if demande.Categorie == fs.Vaccins && !ficheSanitaire {
			continue
		}
		out.Demandes = append(out.Demandes, demande)
	}
	return out
}
