package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"smartdisplay-core/internal/logger"
)

// corsDevMiddleware adds CORS headers for local development (localhost:5500)
// This allows frontend on different ports to call the API during development
func corsDevMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle OPTIONS preflight requests
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5500")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-Role, X-Request-ID")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Add CORS headers to all responses
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5500")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-Role, X-Request-ID")

		next.ServeHTTP(w, r)
	})
}

// requestIDMiddleware generates and injects request ID into each request
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := generateRequestID()
		ctx := ContextWithRequestID(r.Context(), requestID)
		logger.Info("request: id=" + requestID + " method=" + r.Method + " path=" + r.RequestURI)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// generateRequestID creates a unique request identifier
func generateRequestID() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to simple counter if random fails
		return "req-" + hex.EncodeToString(b)
	}
	return "req-" + hex.EncodeToString(b)
}

// respondError writes a standardized error response
func (s *Server) respondError(w http.ResponseWriter, r *http.Request, code ErrorCode, localizedMsg string) {
	requestID := RequestIDFromContext(r.Context())
	if code == CodeInternalError {
		logger.Error("error: id=" + requestID + " code=" + string(code) + " msg=" + localizedMsg)
	} else {
		logger.Info("error: id=" + requestID + " code=" + string(code) + " msg=" + localizedMsg)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code.StatusCode())

	errEnvelope := NewErrorEnvelope(r.Context(), code, localizedMsg)
	failsafe := map[string]interface{}{
		"active":      s.coord.InFailsafeMode(),
		"explanation": s.coord.FailsafeExplanation(),
	}
	out := map[string]interface{}{
		"error":    errEnvelope,
		"failsafe": failsafe,
	}
	_ = json.NewEncoder(w).Encode(out)
}
