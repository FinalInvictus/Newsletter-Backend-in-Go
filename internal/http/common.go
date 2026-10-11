package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Handler holds the single database handle shared by all route handlers.
// This is intentionally not split into controller/service/repository layers.
type Handler struct{ DB *gorm.DB }

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if status != http.StatusNoContent {
		if err := json.NewEncoder(w).Encode(value); err != nil {
			log.Printf("encode response: %v", err)
		}
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MiB request limit.
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return false
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "request body must contain exactly one JSON value")
		return false
	}
	return true
}

func parseUUID(value string) (uuid.UUID, bool) {
	parsed, err := uuid.Parse(value)
	return parsed, err == nil
}

func requireText(value, field string) string {
	if strings.TrimSpace(value) == "" {
		return field + " is required"
	}
	return ""
}

// writeDBError translatees any GORM error to stable HTTP requests without leaking any SQL data
func writeDBError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		writeError(w, http.StatusNotFound, "resource not found")
	case errors.Is(err, gorm.ErrDuplicatedKey):
		writeError(w, http.StatusConflict, "resource already exists")
	case errors.Is(err, gorm.ErrForeignKeyViolated):
		writeError(w, http.StatusBadRequest, "referenced resource does not exist")
	default:
		log.Printf("database operation failed: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
