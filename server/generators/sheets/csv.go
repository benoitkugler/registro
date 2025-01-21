package sheets

import (
	"bytes"
	"encoding/csv"
	"fmt"
)

// CreateCsv creates and returns a .CSV file
func CreateCsv(liste [][]string) ([]byte, error) {
	var b bytes.Buffer
	writer := csv.NewWriter(&b)
	err := writer.WriteAll(liste)
	if err != nil {
		return nil, fmt.Errorf("exporting CSV: %s", err)
	}
	return b.Bytes(), nil
}
