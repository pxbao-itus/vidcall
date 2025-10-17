package common

import (
	"encoding/json"
	"net/http"
)

func WriteResponse(w http.ResponseWriter, statusCode int, data any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = w.Write(raw)
	if err != nil {
		return err
	}

	w.WriteHeader(statusCode)
	
	return nil
}
