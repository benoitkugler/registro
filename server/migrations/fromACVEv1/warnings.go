package main

import (
	"fmt"
	"os"
	"sort"

	rd "registro/migrations/fromACVEv1/rawdata"
)

// warningsFromV1 collects the warnings that can be computed from the trimmed v1
// data alone. It is called before the optional v2 connection so that a dry run
// reports the same diagnostics as a commit.
func warningsFromV1(trimmed baseV1, opts migrationOptions) []string {
	var out []string

	camps := make([]rd.Camp, 0, len(trimmed.camps))
	for _, camp := range trimmed.camps {
		camps = append(camps, camp)
	}
	sort.Slice(camps, func(i, j int) bool { return camps[i].Id < camps[j].Id })
	for _, v1 := range camps {
		if _, ok := specialPrixParAgeFromV1(v1.Id, v1.OptionPrix); ok {
			continue // fully migrated by the camp 338 special case
		}
		if _, has := optionPrixCampFromV1(v1.OptionPrix); has {
			out = append(out, fmt.Sprintf("camp %d: %s v1 daily prices are not uniform and were reduced to one v2 price", v1.Id, v1.Label()))
		}
		switch v1.OptionPrix.Active {
		case "", rd.OptionsPrix.JOUR:
		default:
			out = append(out, fmt.Sprintf("camp %d %s: v1 pricing option %q has no v2 equivalent", v1.Id, v1.Label(), v1.OptionPrix.Active))
		}
	}

	messages := make([]rd.Message, 0, len(trimmed.messages))
	for _, message := range trimmed.messages {
		messages = append(messages, message)
	}
	sort.Slice(messages, func(i, j int) bool { return messages[i].Id < messages[j].Id })
	for _, v1 := range messages {
		if _, ok := eventKindFromV1(v1.Kind); !ok {
			out = append(out, fmt.Sprintf("message %d: kind %d has no v2 equivalent", v1.Id, v1.Kind))
			continue
		}
		if v1.Kind == rd.MAccuseReception &&
			len(inscritParticipantsForFacture(trimmed.participants, v1.IdFacture)) == 0 {
			if _, ok := firstParticipantForFacture(trimmed.participants, v1.IdFacture); !ok {
				out = append(out, fmt.Sprintf("message %d: no participant found for statuation", v1.Id))
			}
		}
	}

	if len(trimmed.documents) > 0 {
		if opts.v1FilesDir == "" {
			out = append(out, fmt.Sprintf("%d legacy document contents were not migrated (pass -v1-files-dir to copy them)", len(trimmed.documents)))
		} else {
			missing := 0
			for id := range trimmed.documents {
				if _, err := os.ReadFile(legacyDocumentPath(opts.v1FilesDir, id)); err != nil {
					missing++
				}
			}
			if missing > 0 {
				out = append(out, fmt.Sprintf("%d legacy document contents were missing from %s", missing, opts.v1FilesDir))
			}
		}

		if !opts.skipDocumentContents {
			miniatures := trimmed.contenuDocuments.ByIdDocument()
			missing := 0
			for id := range trimmed.documents {
				if len(miniatures[id].Miniature) == 0 {
					missing++
				}
			}
			if missing > 0 {
				out = append(out, fmt.Sprintf("%d document(s) have no miniature in contenu_documents (v2 will use a placeholder)", missing))
			}
		}
	}

	return out
}

// checkAidesUnicity refuses les données v1 où plusieurs aides concernent le
// même participant : la v2 n'autorise qu'une aide par participant.
// Un avertissement explicite est affiché pour chaque participant concerné.
func checkAidesUnicity(trimmed baseV1) error {
	byParticipant := make(map[int64][]int64)
	for id, aide := range trimmed.aides {
		byParticipant[aide.IdParticipant] = append(byParticipant[aide.IdParticipant], id)
	}

	duplicates := make([]int64, 0)
	for idParticipant, ids := range byParticipant {
		if len(ids) > 1 {
			duplicates = append(duplicates, idParticipant)
		}
	}
	if len(duplicates) == 0 {
		return nil
	}

	sort.Slice(duplicates, func(i, j int) bool { return duplicates[i] < duplicates[j] })
	for _, idParticipant := range duplicates {
		ids := byParticipant[idParticipant]
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
		warningf("participant %d : %d aides v1 (ids %v) alors qu'une seule aide v2 est autorisée", idParticipant, len(ids), ids)
	}
	return fmt.Errorf("migration interrompue : plusieurs aides v1 pour un même participant")
}
