package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/example/ledgerly/internal/model"
)

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	switch {
	case err == model.ErrNotFound:
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	case err == model.ErrAlreadyExists:
		writeJSON(w, http.StatusConflict, map[string]string{"error": "already exists"})
	case err == model.ErrConflict:
		writeJSON(w, http.StatusConflict, map[string]string{"error": "conflict"})
	case err == model.ErrUnauthorized:
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	default:
		if _, ok := err.(model.ValidationErrors); ok {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "validation failed", "details": err.Error()})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

// decodeJSON decodes a JSON request body.
func decodeJSON(r *http.Request, v any, maxBody int64) error {
	if r.ContentLength > maxBody {
		return model.ErrInvalidInput
	}
	return json.NewDecoder(r.Body).Decode(v)
}

// parseDate parses a date string in YYYY-MM-DD format.
func parseDate(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, _ := time.Parse("2006-01-02", s)
	return t
}

// pathSegment extracts a segment from a URL path (1-based index).
func pathSegment(path string, index int) string {
	parts := splitPath(path)
	if index < 1 || index > len(parts) {
		return ""
	}
	return parts[index-1]
}

func splitPath(path string) []string {
	var parts []string
	var b []byte
	for i := 0; i <= len(path); i++ {
		c := byte('/')
		if i < len(path) {
			c = path[i]
		}
		if c == '/' {
			if len(b) > 0 {
				parts = append(parts, string(b))
				b = nil
			}
		} else {
			b = append(b, c)
		}
	}
	return parts
}
