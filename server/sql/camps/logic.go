package camps

import pr "registro/sql/personnes"

type ParticipantExt struct {
	Camp        *Camp
	Participant Participant
	Personne    pr.Personne
}

// StatistiquesInscrits détails le nombre d'inscriptions
// sur un séjour
type StatistiquesInscrits struct {
	Inscriptions                            int // nombre total
	InscriptionsFilles, InscriptionsSuisses int
	Valides, ValidesFilles, ValidesSuisses  int // confirmés
	Refus, AStatuer, Exceptions, Attente    int
}

func (p ParticipantExt) add(stats *StatistiquesInscrits) {
	stats.Inscriptions += 1

	isFille := p.Personne.Sexe == pr.Woman
	isSuisse := p.Personne.Nationnalite == pr.Suisse

	if isFille {
		stats.InscriptionsFilles += 1
	}
	if isSuisse {
		stats.InscriptionsSuisses += 1
	}

	switch p.Participant.Statut {
	case Refuse:
		stats.Refus += 1
	case AStatuer:
		stats.AStatuer += 1
	case AttenteProfilInvalide:
		stats.Exceptions += 1
	case AttenteCampComplet, EnAttenteReponse:
		stats.Attente += 1
	case Inscrit:
		stats.Valides += 1
		if isFille {
			stats.ValidesFilles += 1
		}
		if isSuisse {
			stats.ValidesSuisses += 1
		}
	}
}

// CampLoader permet d'accéder à diverses
// propriété d'un séjour nécessitant la liste des inscrits.
type CampLoader struct {
	Camp         *Camp
	Participants Participants // liste (exacte) des participants du camp
	// Doit contenir au moins les participants
	Personnes pr.Personnes
}

func (cd CampLoader) Stats() StatistiquesInscrits {
	var stats StatistiquesInscrits
	for _, participant := range cd.Participants {
		ext := ParticipantExt{cd.Camp, participant, cd.Personnes[participant.IdPersonne]}
		ext.add(&stats)
	}
	return stats
}
