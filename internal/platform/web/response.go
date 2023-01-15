package web

import (
	"context"
	"encoding/json"
	"go.opentelemetry.io/otel/attribute"
	"net/http"
)

//var (
//	// ErrNotFound is abstracting the mgo not found error.
//	ErrNotFound = errors.New("Entity not found")
//
//	// ErrInvalidID occurs when an ID is not in a valid form.
//	ErrInvalidID = errors.New("ID is not in it's proper form")
//)
//
//// JSONError is the response for errors that occur within the API.
//type JSONError struct {
//	Error string `json:"error"`
//}
//
//// Error handles all error responses for the API.
//func Error(cxt context.Context, w http.ResponseWriter, err error) {
//	switch errors.Cause(err) {
//	case ErrNotFound:
//		RespondError(cxt, w, err, http.StatusNotFound)
//		return
//
//	case ErrInvalidID:
//		RespondError(cxt, w, err, http.StatusBadRequest)
//		return
//	}
//
//	RespondError(cxt, w, err, http.StatusInternalServerError)
//}
//
//// RespondError sends JSON describing the error
//func RespondError(ctx context.Context, w http.ResponseWriter, err error, code int) {
//	Respond(ctx, w, JSONError{Error: err.Error()}, code)
//}


// Respond converts a Go value to JSON and sends it to the client.
func Respond(ctx context.Context, w http.ResponseWriter, data any, statusCode int) error {
	ctx, span := AddSpan(ctx, "foundation.web.response", attribute.Int("status", statusCode))
	defer span.End()

	SetStatusCode(ctx, statusCode)

	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	return nil
}
