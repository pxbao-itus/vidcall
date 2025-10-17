package common

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"

	"github.com/go-chi/chi/v5"
)

func BindRequest[T any](r *http.Request, t T) error {
	if r.URL.RawQuery != "" {
		q := r.URL.Query()
		v := reflect.ValueOf(t).Elem()
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			if tag, ok := field.Tag.Lookup("query"); ok {
				if values, exists := q[tag]; exists && len(values) > 0 {
					v.Field(i).Set(reflect.ValueOf(values[0]).Convert(field.Type))
				}
			}
		}
	}

	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		return nil
	}

	return nil
}

func GetUserID(r *http.Request) (string, error) {
	// temporary consider user id as client ip
	if userId := r.Header.Get("X-User-ID"); userId != "" {
		return userId, nil
	}

	return r.RemoteAddr, nil
}

func GetParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}
