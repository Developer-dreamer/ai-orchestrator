package helper

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

var ErrInvalidPayload = errors.New("invalid payload")

func FromJSON(r io.Reader, entity any) error {
	err := json.NewDecoder(r).Decode(entity)
	if err != nil {
		return errors.Join(ErrInvalidPayload, err)
	}
	return nil
}

func ToJSON(w io.Writer, entity any) error {
	return json.NewEncoder(w).Encode(entity)
}

func WriteJSONError(w http.ResponseWriter, httpStatus int, message string, err error) {
	response := map[string]string{
		"message": message,
	}
	if err != nil {
		response["error"] = err.Error()
	}
	WriteJSONResponse(w, httpStatus, response)
}

func WriteJSONResponse(w http.ResponseWriter, httpStatus int, entity any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	err := ToJSON(w, entity)
	if err != nil {
		err = ToJSON(w, map[string]string{
			"message": "failed to return object",
		})
		// If writing JSON response fails, don't keep trying, just send status code.
		if err != nil {
			return
		}
	}
}
