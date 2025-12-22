package rawdata

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	_ "github.com/lib/pq" // SQL driver registration
)

func loadJSON(out interface{}, src interface{}) error {
	if src == nil {
		return nil //zero value out
	}
	bs, ok := src.([]byte)
	if !ok {
		return errors.New("not a []byte")
	}
	return json.Unmarshal(bs, out)
}

func dumpJSON(s interface{}) (driver.Value, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return driver.Value(b), nil
}

func (s *Exemplaires) Scan(src interface{}) error {
	return loadJSON(s, src)
}

func (s Exemplaires) Value() (driver.Value, error) {
	return dumpJSON(s)
}

func (s *OptionsCamp) Scan(src interface{}) error {
	return loadJSON(s, src)
}

func (s OptionsCamp) Value() (driver.Value, error) {
	return dumpJSON(s)
}

func (s *Envois) Scan(src interface{}) error {
	return loadJSON(s, src)
}

func (s Envois) Value() (driver.Value, error) {
	return dumpJSON(s)
}

func (m *Modules) Scan(src interface{}) error {
	return loadJSON(m, src)
}

func (m Modules) Value() (driver.Value, error) {
	return dumpJSON(m)
}

func (s *DestinatairesOptionnels) Scan(src interface{}) error {
	return loadJSON(s, src)
}

func (s DestinatairesOptionnels) Value() (driver.Value, error) {
	return dumpJSON(s)
}

func (s *EtatNegociation) Scan(src interface{}) error {
	return loadJSON(s, src)
}

func (s EtatNegociation) Value() (driver.Value, error) {
	return dumpJSON(s)
}

func (s *OptionsParticipant) Scan(src interface{}) error {
	return loadJSON(s, src)
}

func (s OptionsParticipant) Value() (driver.Value, error) {
	return dumpJSON(s)
}

func (s *OptionPrixParticipant) Scan(src interface{}) error {
	return loadJSON(s, src)
}

func (s OptionPrixParticipant) Value() (driver.Value, error) {
	return dumpJSON(s)
}

func (s *Remises) Scan(src interface{}) error {
	return loadJSON(s, src)
}

func (s Remises) Value() (driver.Value, error) {
	return dumpJSON(s)
}

func (s *FicheSanitaire) Scan(src interface{}) error {
	return loadJSON(s, src)
}

func (s FicheSanitaire) Value() (driver.Value, error) {
	return dumpJSON(s)
}

func (s *OptionPrixCamp) Scan(src interface{}) error {
	return loadJSON(s, src)
}

func (s OptionPrixCamp) Value() (driver.Value, error) {
	return dumpJSON(s)
}

func (s *ListeVetements) Scan(src interface{}) error {
	return loadJSON(s, src)
}

func (s ListeVetements) Value() (driver.Value, error) {
	return dumpJSON(s)
}

func (s *IdentificationId) Scan(src interface{}) error {
	return loadJSON(s, src)
}

func (s IdentificationId) Value() (driver.Value, error) {
	return dumpJSON(s)
}

func (s *OptionnalPlage) Scan(src interface{}) error {
	return loadJSON(s, src)
}

func (s OptionnalPlage) Value() (driver.Value, error) {
	return dumpJSON(s)
}

func (s *InfoDon) Scan(src interface{}) error {
	return loadJSON(s, src)
}

func (s InfoDon) Value() (driver.Value, error) {
	return dumpJSON(s)
}

func (s *ListeAttente) Scan(src interface{}) error {
	return loadJSON(s, src)
}

func (s ListeAttente) Value() (driver.Value, error) {
	return dumpJSON(s)
}

func (s *Plage) Scan(src interface{}) error {
	return loadJSON(s, src)
}

func (s Plage) Value() (driver.Value, error) {
	return dumpJSON(s)
}

func (s *ParticipantInscriptions) Scan(src interface{}) error {
	return loadJSON(s, src)
}

func (s ParticipantInscriptions) Value() (driver.Value, error) {
	return dumpJSON(s)
}

func (s *ResponsableLegal) Scan(src interface{}) error {
	return loadJSON(s, src)
}

func (s ResponsableLegal) Value() (driver.Value, error) {
	return dumpJSON(s)
}

func (s *Coordonnees) Scan(src interface{}) error {
	return loadJSON(s, src)
}

func (s Coordonnees) Value() (driver.Value, error) {
	return dumpJSON(s)
}
