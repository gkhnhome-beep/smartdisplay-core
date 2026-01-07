package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"smartdisplay-core/internal/auth"
	"smartdisplay-core/internal/logger"
	"strings"
)

// corsDevMiddleware adds CORS headers for local development (localhost:5500)
// This allows frontend on different ports to call the API during development
func corsDevMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle OPTIONS preflight requests
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5500")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-Role, X-Request-ID, X-SmartDisplay-PIN, X-SmartDisplay-Role")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Add CORS headers to all responses
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5500")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-Role, X-Request-ID, X-SmartDisplay-PIN, X-SmartDisplay-Role")

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

// authMiddleware extracts PIN from request and validates it
// FAZ L1: PIN-based authentication middleware
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Development/testing override: allow X-User-Role header to set role locally
		// This is intentionally simple to ease local testing of admin-only endpoints.
		roleHeader := r.Header.Get("X-User-Role")
		var authCtx *auth.AuthContext
		if roleHeader != "" {
			switch strings.ToLower(roleHeader) {
			case "admin":
				authCtx = &auth.AuthContext{Role: auth.Admin, Authenticated: true, PIN: ""}
			case "user":
				authCtx = &auth.AuthContext{Role: auth.User, Authenticated: true, PIN: ""}
			default:
				authCtx = &auth.AuthContext{Role: auth.Guest, Authenticated: false, PIN: ""}
			}
			logger.Info("auth: role override via X-User-Role header: " + roleHeader)
		} else {
			// Extract PIN from header
			pin := r.Header.Get("X-SmartDisplay-PIN")

			// Validate PIN and get auth context
			validatedCtx, err := auth.ValidatePIN(pin)
			if err != nil {
				// Invalid PIN - treat as guest
				logger.Info("auth: invalid PIN provided (treating as guest)")
				authCtx = &auth.AuthContext{
					Role:          auth.Guest,
					Authenticated: false,
					PIN:           "",
				}
			} else {
				authCtx = validatedCtx
			}
		}

		// Never log PIN
		logger.Info("auth: role=" + string(authCtx.Role) + " authenticated=" + boolToString(authCtx.Authenticated))

		// Attach auth context to request
		ctx := context.WithValue(r.Context(), ctxAuthContext, authCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ctxAuthContext is the context key for auth context
const ctxAuthContext contextKey = "auth_context"

// getAuthContext retrieves auth context from request context
func getAuthContext(r *http.Request) *auth.AuthContext {
	authCtx, ok := r.Context().Value(ctxAuthContext).(*auth.AuthContext)
	if !ok {
		// No auth context - return guest
		return &auth.AuthContext{
			Role:          auth.Guest,
			Authenticated: false,
			PIN:           "",
		}
	}
	return authCtx
}

// requireAdmin returns 403 if request is not from admin role
func requireAdmin(r *http.Request) bool {
	authCtx := getAuthContext(r)
	return authCtx.IsAdmin()
}

// boolToString converts bool to string
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
