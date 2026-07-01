package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/ndmt1at21/devlog/backend/internal/apierr"
)

// envelope is the uniform response shape for every JSON API endpoint:
//
//	{ "code": 0, "message": "OK", "traceId": "…", "data": … }
//
// code is 0 on success or a stable apierr code the frontend maps to a localized
// message; traceId is unique per request (also echoed in the X-Trace-Id header).
type envelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"traceId"`
	Data    any    `json:"data"`
}

// writeJSON writes a success envelope (code 0) with the given data and status.
func writeJSON(w http.ResponseWriter, r *http.Request, status int, data any) {
	writeEnvelope(w, r, status, apierr.CodeOK, "OK", data)
}

// writeError writes an error envelope. Any *apierr.Error is honored for its
// Code/Status/Message; anything else becomes a generic 500 (ErrInternal).
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	var e *apierr.Error
	if !errors.As(err, &e) {
		e = apierr.ErrInternal
	}
	writeEnvelope(w, r, e.Status, e.Code, e.Message, nil)
}

func writeEnvelope(w http.ResponseWriter, r *http.Request, status, code int, message string, data any) {
	traceID := traceIDFrom(r.Context())
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if traceID != "" {
		w.Header().Set("X-Trace-Id", traceID)
	}
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(envelope{
		Code:    code,
		Message: message,
		TraceID: traceID,
		Data:    data,
	}); err != nil {
		log.Printf("encode response: %v", err)
	}
}

// decodeJSON reads a JSON body into dst, writing a 400 envelope and returning
// false on failure. Bodies are capped to guard against abuse.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, r, apierr.ErrBadRequest)
		return false
	}
	return true
}
