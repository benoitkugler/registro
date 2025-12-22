package rawdata

import (
	"encoding/json"
	"time"
)

func (t Time) MarshalJSON() ([]byte, error) {
	return t.Time().MarshalJSON()
}

func (t *Time) UnmarshalJSON(data []byte) error {
	return (*time.Time)(t).UnmarshalJSON(data)
}

func (t Date) MarshalJSON() ([]byte, error) {
	s := t.Time().Format("2006-01-02")
	return json.Marshal(s)
}

func (t *Date) UnmarshalJSON(data []byte) error {
	var tmp string
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	parsed, err := time.Parse("2006-01-02", tmp)
	*t = Date(parsed)
	return err
}
