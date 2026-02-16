package logic

import (
	"slices"

	"registro/crypto"
	"registro/logic/search"
	cps "registro/sql/camps"
	"registro/sql/dons"
	ds "registro/sql/dossiers"
	"registro/sql/files"
	fs "registro/sql/files"
	pr "registro/sql/personnes"
	"registro/utils"
)

type CampItem struct {
	Id    cps.IdCamp
	Label string
	IsOld bool // true if the end is passed by 45 jours
}

func NewCampItem(camp cps.Camp) CampItem {
	const deltaOld = 45
	return CampItem{camp.Id, camp.Label(), camp.IsPassedBy(deltaOld)}
}

func NewCampItems(camps cps.Camps) []CampItem {
	list := utils.MapValues(camps)
	slices.SortFunc(list, func(a, b cps.Camp) int { return int(a.Id - b.Id) })
	slices.SortStableFunc(list, func(a, b cps.Camp) int { return -a.DateDebut.Time().Compare(b.DateDebut.Time()) })

	out := make([]CampItem, len(list))
	for i, camp := range list {
		out[i] = NewCampItem(camp)
	}
	return out
}

func LoadCamps(db cps.DB) ([]CampItem, error) {
	camps, err := cps.SelectAllCamps(db)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	return NewCampItems(camps), nil
}

func SelectPersonne(db pr.DB, pattern string, removeTemp bool) ([]search.PersonneHeader, error) {
	const maxCount = 10
	personnes, err := pr.SelectAllPersonnes(db)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	if removeTemp {
		personnes.RemoveTemp()
	}
	out := search.FilterPersonnes(personnes, pattern)
	if len(out) > maxCount {
		out = out[:maxCount]
	}
	return out, nil
}

func SearchSimilaires(db pr.DB, id pr.IdPersonne) ([]search.ScoredPersonne, error) {
	const maxCount = 5
	// can't use [SelectAllFieldsForSimilaires] since it does
	// not returns Temp profiles
	input, err := pr.SelectPersonne(db, id)
	if err != nil {
		return nil, utils.SQLError(err)
	}
	personnes, err := search.SelectAllFieldsForSimilaires(db)
	if err != nil {
		return nil, err
	}

	_, filtered := search.ChercheSimilaires(personnes, search.NewPatternsSimilarite(input.Identite))
	if len(filtered) > maxCount {
		filtered = filtered[:maxCount]
	}
	return filtered, nil
}

// not included, will cascade on delete : Fichesanitaire, Demande
type PersonneReferences struct {
	Participants []cps.IdParticipant
	Equipiers    []cps.IdEquipier
	Dossiers     []ds.IdDossier
	Files        []fs.IdFile
	Dons         []dons.IdDon // TODO
}

func (pr PersonneReferences) Empty() bool {
	return len(pr.Participants) == 0 &&
		len(pr.Equipiers) == 0 &&
		len(pr.Dossiers) == 0 &&
		len(pr.Files) == 0 &&
		len(pr.Dons) == 0
}

// check all the use of [id] in other tables
// wraps error
func CheckPersonneReferences(db pr.DB, id pr.IdPersonne) (out PersonneReferences, _ error) {
	parts, err := cps.SelectParticipantsByIdPersonnes(db, id)
	if err != nil {
		return out, utils.SQLError(err)
	}
	out.Participants = parts.IDs()

	equipiers, err := cps.SelectEquipiersByIdPersonnes(db, id)
	if err != nil {
		return out, utils.SQLError(err)
	}
	out.Equipiers = equipiers.IDs()

	dossiers, err := ds.SelectDossiersByIdResponsables(db, id)
	if err != nil {
		return out, utils.SQLError(err)
	}
	out.Dossiers = dossiers.IDs()

	links1, err := fs.SelectFilePersonnesByIdPersonnes(db, id)
	if err != nil {
		return out, utils.SQLError(err)
	}
	out.Files = links1.IdFiles()

	// links2, err := rd.SelectDonDonateursByIdPersonnes(db, id)
	// if err != nil {
	// 	return out, utils.SQLError(err)
	// }
	// out.Dons = links2

	return out, nil
}

const EndpointEspacePerso = "espace-perso"

func EspacePersoURL(key crypto.Encrypter, host string, dossier ds.IdDossier, queryParams ...utils.QueryParam) string {
	crypted := crypto.EncryptID(key, dossier)
	queryParams = append(queryParams, utils.QP("token", crypted))
	return utils.BuildUrl(host, EndpointEspacePerso, queryParams...)
}

// PublicFile expose un accès protégé à un fichier,
// permettant téléchargement/suppression/modification.
type PublicFile struct {
	Key string // crypted
	files.File
}

func NewPublicFile(key crypto.Encrypter, file files.File) PublicFile {
	return PublicFile{
		Key:  crypto.EncryptID(key, file.Id),
		File: file,
	}
}
