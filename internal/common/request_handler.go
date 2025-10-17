package common

import (
	"encoding/json"
	"net/http"
)

func BindRequest[T any](r *http.Request, t T) error {
	return json.NewDecoder(r.Body).Decode(&t)
}
