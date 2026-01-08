package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

	// Wrap with auth middleware (FAZ L1: PIN-based authentication)
	handler = authMiddleware(handler)

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

	// Frontend static files (serve web directory with no-cache headers)
	// Use absolute path to find web directory reliably
	webDir := filepath.Join(os.Getenv("PWD"), "web")
	if _, err := os.Stat(webDir); err != nil {
		// Fallback to relative path if PWD env var not set
		webDir = "web"
	}

	// Wrap FileServer with cache-control middleware for dynamic files
	fs := http.FileServer(http.Dir(webDir))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set cache control headers for .js, .css, .html files to prevent caching versioned files
		if filepath.Ext(r.URL.Path) == ".js" || filepath.Ext(r.URL.Path) == ".css" || filepath.Ext(r.URL.Path) == ".html" {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, max-age=0")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			w.Header().Del("ETag")
			w.Header().Del("Last-Modified")
		}
		fs.ServeHTTP(w, r)
	}))

	// Home Endpoints (D2)
	mux.HandleFunc("/api/ui/home/state", s.handleHomeState)
	mux.HandleFunc("/api/ui/home/summary", s.handleHomeSummary)

	// Alarm Endpoints (D3)
	mux.HandleFunc("/api/ui/alarm/state", s.handleAlarmState)
	mux.HandleFunc("/api/ui/alarm/summary", s.handleAlarmSummary)
	mux.HandleFunc("/api/ui/alarm/action", s.handleAlarmAction) // A4: Controlled write operations

	// Alarmo Monitoring (read-only)
	mux.HandleFunc("/api/ui/alarmo/status", s.handleAlarmoStatus)
	mux.HandleFunc("/api/ui/alarmo/sensors", s.handleAlarmoSensors)
	mux.HandleFunc("/api/ui/alarmo/events", s.handleAlarmoEvents)

	// Alarmo Control (write operations)
	mux.HandleFunc("/api/ui/alarmo/arm", s.handleAlarmoArm)
	mux.HandleFunc("/api/ui/alarmo/disarm", s.handleAlarmoDisarm)

	// Guest Endpoints (D4)
	mux.HandleFunc("/api/ui/guest/state", s.handleGuestState)
	mux.HandleFunc("/api/ui/guest/summary", s.handleGuestSummary)
	mux.HandleFunc("/api/ui/guest/request", s.handleGuestRequest)
	mux.HandleFunc("/api/ui/guest/request/", s.handleGuestRequestStatus) // FAZ L2: Status check with dynamic request_id
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

	// Settings - Home Assistant Integration (FAZ S2, admin-only)
	mux.HandleFunc("/api/settings/homeassistant", s.handleHASettingsSave)
	mux.HandleFunc("/api/settings/homeassistant/status", s.handleHASettingsStatus)

	// Settings - Home Assistant Connection Test (FAZ S3, admin-only)
	mux.HandleFunc("/api/settings/homeassistant/test", s.handleHASettingsTest)

	// Settings - Home Assistant Initial Sync (FAZ S5, admin-only)
	mux.HandleFunc("/api/settings/homeassistant/sync", s.handleHAInitialSync)

	// Devices - Lighting (read: all roles, write: user/admin)
	mux.HandleFunc("/api/devices/lights", s.handleDevicesLights)
	mux.HandleFunc("/api/devices/lights/toggle", s.handleDevicesLightsToggle)
	mux.HandleFunc("/api/devices/lights/set", s.handleDevicesLightsSet)

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
