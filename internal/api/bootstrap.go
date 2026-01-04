package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"smartdisplay-core/internal/health"
	"smartdisplay-core/internal/logger"
	"time"
)

// panicRecovery is HTTP middleware that recovers from panics and returns 500
func panicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(fmt.Sprintf("panic recovered: %v", err))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				// Standard error envelope
				fmt.Fprintf(w, `{"ok":false,"error":"internal server error"}`)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// startHTTPServer creates, configures, and starts the HTTP server
func (s *Server) startHTTPServer(port int) error {
	// Register all routes (centralized, deterministic order)
	mux := s.registerRoutes()

	// Wrap with CORS middleware (dev: localhost:5500)
	handler := corsDevMiddleware(mux)

	// Wrap with request ID middleware
	handler = requestIDMiddleware(handler)

	// Wrap with panic recovery middleware
	handler = panicRecovery(handler)

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:           ":" + itoa(port),
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start server in goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("api server error:", err)
		}
	}()

	logger.Info(fmt.Sprintf("http server started on port %d", port))
	return nil
}

// registerRoutes sets up all HTTP routes in deterministic order
func (s *Server) registerRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Health endpoint (first - critical for monitoring)
	mux.HandleFunc("/health", health.HealthHandler)
	mux.HandleFunc("/api/health", health.HealthHandler)

	// Home Endpoints (D2)
	mux.HandleFunc("/api/ui/home/state", s.handleHomeState)
	mux.HandleFunc("/api/ui/home/summary", s.handleHomeSummary)

	// Alarm Endpoints (D3)
	mux.HandleFunc("/api/ui/alarm/state", s.handleAlarmState)
	mux.HandleFunc("/api/ui/alarm/summary", s.handleAlarmSummary)

	// Guest Endpoints (D4)
	mux.HandleFunc("/api/ui/guest/state", s.handleGuestState)
	mux.HandleFunc("/api/ui/guest/summary", s.handleGuestSummary)
	mux.HandleFunc("/api/ui/guest/request", s.handleGuestRequest)
	mux.HandleFunc("/api/ui/guest/exit", s.handleGuestExit)

	// Menu Endpoints (D5)
	mux.HandleFunc("/api/ui/menu", s.handleMenu)

	// Logbook Endpoints (D6)
	mux.HandleFunc("/api/ui/logbook", s.handleLogbook)
	mux.HandleFunc("/api/ui/logbook/summary", s.handleLogbookSummary)

	// Settings Endpoints (D7)
	mux.HandleFunc("/api/ui/settings", s.handleSettings)
	mux.HandleFunc("/api/ui/settings/action", s.handleSettingsAction)

	// Accessibility & Voice
	mux.HandleFunc("/api/ui/accessibility", s.handleAccessibility)
	mux.HandleFunc("/api/ui/voice", s.handleVoice)

	// AI Endpoints
	mux.HandleFunc("/api/ai/daily", s.handleAIDaily)
	mux.HandleFunc("/api/ai/anomalies", s.handleAIAnomalies)
	mux.HandleFunc("/api/ai/morning", s.handleAIMorning)
	mux.HandleFunc("/api/ai/insight", s.handleAIInsight)
	mux.HandleFunc("/api/ai/insight/explain", s.handleAIExplain)
	mux.HandleFunc("/api/ai/history", s.handleAIHistory)

	// Legacy endpoints (backward compatibility)
	mux.HandleFunc("/api/overview", s.handleOverview)
	mux.HandleFunc("/api/alarm/arm", s.handleAlarmArm)
	mux.HandleFunc("/api/alarm/disarm", s.handleAlarmDisarm)
	mux.HandleFunc("/api/guest/approve", s.handleGuestApprove)
	mux.HandleFunc("/api/guest/deny", s.handleGuestDeny)
	mux.HandleFunc("/api/guest/request", s.handleGuestRequest)
	mux.HandleFunc("/api/guest/exit", s.handleGuestExit)
	mux.HandleFunc("/api/failsafe", s.handleFailsafe)
	mux.HandleFunc("/api/logbook", s.handleLogbook)

	// First-boot Setup Endpoints (D0)
	mux.HandleFunc("/api/setup/firstboot/status", s.handleFirstBootStatus)
	mux.HandleFunc("/api/setup/firstboot/next", s.handleFirstBootNext)
	mux.HandleFunc("/api/setup/firstboot/back", s.handleFirstBootBack)
	mux.HandleFunc("/api/setup/firstboot/complete", s.handleFirstBootComplete)

	// UI Help & Scorecard
	mux.HandleFunc("/api/ui/help", s.handleUIHelp)
	mux.HandleFunc("/api/ui/scorecard", s.handleUIScorecard)

	// Admin Endpoints (must come after user endpoints to avoid shadowing)
	mux.HandleFunc("/api/admin/smoke", s.handleAdminSmoke)
	mux.HandleFunc("/api/admin/restart", s.handleAdminRestart)
	mux.HandleFunc("/api/admin/backup", s.handleAdminBackup)
	mux.HandleFunc("/api/admin/restore", s.handleAdminRestore)
	mux.HandleFunc("/api/admin/telemetry/summary", s.handleTelemetrySummary)
	mux.HandleFunc("/api/admin/telemetry/optin", s.handleTelemetryOptIn)
	mux.HandleFunc("/api/admin/update/status", s.handleUpdateStatus)
	mux.HandleFunc("/api/admin/update/stage", s.handleUpdateStage)

	return mux
}

// ShutdownCtx gracefully shuts down the server with timeout
func (s *Server) ShutdownCtx(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	// Cancel the shutdown context to signal handlers
	s.mu.Lock()
	if s.shutdownCxl != nil {
		s.shutdownCxl()
	}
	s.mu.Unlock()

	// Perform HTTP server shutdown with provided timeout context
	return s.httpServer.Shutdown(ctx)
}
