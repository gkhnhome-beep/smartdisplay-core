package api

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// ErrorCode represents a standard error code
type ErrorCode string

const (
	// HTTP error codes
	CodeBadRequest         ErrorCode = "bad_request"         // 400
	CodeUnauthorized       ErrorCode = "unauthorized"        // 401
	CodeForbidden          ErrorCode = "forbidden"           // 403
	CodeNotFound           ErrorCode = "not_found"           // 404
	CodeMethodNotAllowed   ErrorCode = "method_not_allowed"  // 405
	CodeConflict           ErrorCode = "conflict"            // 409
	CodeInternalError      ErrorCode = "internal_error"      // 500
	CodeUpstreamError      ErrorCode = "upstream_error"      // 502
	CodeServiceUnavailable ErrorCode = "service_unavailable" // 503
)

// ErrorEnvelope is the standard error response format
type ErrorEnvelope struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Timestamp int64  `json:"timestamp"`
}

// HTTPError maps error codes to status codes and localization keys
func (e ErrorCode) StatusCode() int {
	switch e {
	case CodeBadRequest:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeMethodNotAllowed:
		return http.StatusMethodNotAllowed
	case CodeConflict:
		return http.StatusConflict
	case CodeUpstreamError:
		return http.StatusBadGateway
	case CodeInternalError:
		return http.StatusInternalServerError
	case CodeServiceUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// LocalizationKey returns the i18n key for error message
func (e ErrorCode) LocalizationKey() string {
	return fmt.Sprintf("error.%s", e)
}

// NewErrorEnvelope creates a standardized error response
func NewErrorEnvelope(ctx context.Context, code ErrorCode, localizedMessage string) *ErrorEnvelope {
	requestID := ""
	if id := ctx.Value(ctxRequestID); id != nil {
		requestID = id.(string)
	}
	return &ErrorEnvelope{
		Code:      string(code),
		Message:   localizedMessage,
		RequestID: requestID,
		Timestamp: time.Now().Unix(),
	}
}

// ctxRequestID is the context key for request ID
type contextKey string

const ctxRequestID contextKey = "request_id"

// RequestIDFromContext extracts the request ID from context
func RequestIDFromContext(ctx context.Context) string {
	if id := ctx.Value(ctxRequestID); id != nil {
		return id.(string)
	}
	return ""
}

// ContextWithRequestID returns a new context with request ID
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ctxRequestID, requestID)
}
