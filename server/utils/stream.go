package utils

import (
	"encoding/json"
	"iter"
	"net/http"
)

func StreamJSON[T any](resp http.ResponseWriter, items iter.Seq2[T, error]) error {
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(resp)
	for item, err := range items {
		if err != nil {
			resp.Write([]byte(err.Error()))
			return err
		}
		if err := enc.Encode(item); err != nil {
			return err
		}
		if flusher, ok := resp.(http.Flusher); ok {
			flusher.Flush()
		}
	}
	return nil
}
