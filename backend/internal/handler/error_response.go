package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"backend/internal/domain"
	"backend/pkg/api"
)

const (
	contentTypeJSON        = "application/json"
	causeInvalidRequest    = "invalid_request"
	causeMethodNotAllowed  = "method_not_allowed"
	causeInvalidCredential = "invalid_credential"
	causeUnauthorized      = "unauthorized"
	causeDuplicate         = "duplicate"
	causeInternalError     = "internal_error"
)

func respondInvalidJSON(w http.ResponseWriter) {
	respondAPIError(w, http.StatusBadRequest, causeInvalidRequest, "body", "request body must be valid JSON")
}

func respondInvalidField(w http.ResponseWriter, field string) {
	respondAPIError(w, http.StatusBadRequest, causeInvalidRequest, field, fmt.Sprintf("%s is invalid", field))
}

func respondMethodNotAllowed(w http.ResponseWriter, allowed string) {
	w.Header().Set("Allow", allowed)
	respondAPIError(w, http.StatusMethodNotAllowed, causeMethodNotAllowed, "method", fmt.Sprintf("use %s", allowed))
}

func respondDuplicateField(w http.ResponseWriter, field string) {
	respondAPIError(w, http.StatusConflict, causeDuplicate, field, fmt.Sprintf("%s already exists", field))
}

func respondInvalidCredential(w http.ResponseWriter, status int) {
	respondAPIError(w, status, causeInvalidCredential, "credential", "credential mismatch")
}

func respondUnauthorizedSession(w http.ResponseWriter) {
	respondAPIError(w, http.StatusUnauthorized, causeUnauthorized, "session", "invalid or expired session")
}

func respondInternalServerError(w http.ResponseWriter) {
	respondAPIError(w, http.StatusInternalServerError, causeInternalError, "server", "internal server error")
}

func respondAPIError(w http.ResponseWriter, status int, cause, field, message string) {
	apiErr, err := domain.NewAPIError(cause, field, message)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}

	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(api.NewErrorResponse(apiErr))
}
