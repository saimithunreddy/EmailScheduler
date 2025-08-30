package service

import (
	"encoding/json"
	"net/http"
)

func DecodeScheduleEmailNotification(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(&dst); err != nil {
		return err
	}

	if err := r.Body.Close(); err != nil {
		return err
	}

	return nil
}

func EncodeJSONBody(w http.ResponseWriter, src interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(src)
}
