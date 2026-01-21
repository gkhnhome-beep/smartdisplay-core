package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"smartdisplay-core/internal/health"
	"smartdisplay-core/internal/logger"
	"strconv"
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
		Addr:           ":" + strconv.Itoa(port),
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
	// Kullanıcı yönetimi
	mux.HandleFunc("/api/users/list", s.HandleUserList)
	mux.HandleFunc("/api/users/add", s.HandleUserAdd)
	mux.HandleFunc("/api/users/update", s.HandleUserUpdate)
	mux.HandleFunc("/api/users/delete", s.HandleUserDelete)
	// Sağlık
	mux.HandleFunc("/api/health", health.HealthHandler)
	mux.HandleFunc("/health", health.HealthHandler)
	// Login
	mux.HandleFunc("/api/login", s.handleLogin)
	// Ana ve alt endpointler
	mux.HandleFunc("/api/ui/home/state", s.handleHomeState)
	mux.HandleFunc("/api/ui/home/summary", s.handleHomeSummary)
	mux.HandleFunc("/api/ui/alarm/state", s.handleAlarmState)
	mux.HandleFunc("/api/ui/alarm/summary", s.handleAlarmSummary)
	mux.HandleFunc("/api/ui/alarm/action", s.handleAlarmAction)
	mux.HandleFunc("/api/ui/alarmo/status", s.handleAlarmoStatus)
	mux.HandleFunc("/api/ui/alarmo/sensors", s.handleAlarmoSensors)
	mux.HandleFunc("/api/ui/alarmo/events", s.handleAlarmoEvents)
	mux.HandleFunc("/api/ui/alarmo/arm", s.handleAlarmoArm)
	mux.HandleFunc("/api/ui/alarmo/disarm", s.handleAlarmoDisarm)
	mux.HandleFunc("/api/ui/guest/state", s.handleGuestState)
	mux.HandleFunc("/api/ui/guest/summary", s.handleGuestSummary)
	mux.HandleFunc("/api/ui/guest/request", s.handleGuestRequest)
	mux.HandleFunc("/api/ui/guest/request/", s.handleGuestRequestStatus)
	mux.HandleFunc("/api/ui/guest/exit", s.handleGuestExit)
	mux.HandleFunc("/api/ui/menu", s.handleMenu)
	mux.HandleFunc("/api/ui/logbook", s.handleLogbook)
	mux.HandleFunc("/api/ui/logbook/summary", s.handleLogbookSummary)
	mux.HandleFunc("/api/ui/settings", s.handleSettings)
	mux.HandleFunc("/api/ui/settings/action", s.handleSettingsAction)
	mux.HandleFunc("/api/ui/accessibility", s.handleAccessibility)
	mux.HandleFunc("/api/ui/voice", s.handleVoice)
	mux.HandleFunc("/api/ai/daily", s.handleAIDaily)
	mux.HandleFunc("/api/ai/anomalies", s.handleAIAnomalies)
	mux.HandleFunc("/api/ai/morning", s.handleAIMorning)
	mux.HandleFunc("/api/ai/insight", s.handleAIInsight)
	mux.HandleFunc("/api/ai/insight/explain", s.handleAIExplain)
	mux.HandleFunc("/api/ai/history", s.handleAIHistory)
	mux.HandleFunc("/api/overview", s.handleOverview)
	mux.HandleFunc("/api/alarm/arm", s.handleAlarmArm)
	mux.HandleFunc("/api/alarm/disarm", s.handleAlarmDisarm)
	mux.HandleFunc("/api/guest/approve", s.handleGuestApprove)
	mux.HandleFunc("/api/guest/deny", s.handleGuestDeny)
	mux.HandleFunc("/api/failsafe", s.handleFailsafe)
	mux.HandleFunc("/api/logbook", s.handleLogbook)
	mux.HandleFunc("/api/setup/firstboot/status", s.handleFirstBootStatus)
	mux.HandleFunc("/api/setup/firstboot/next", s.handleFirstBootNext)
	mux.HandleFunc("/api/setup/firstboot/back", s.handleFirstBootBack)
	mux.HandleFunc("/api/setup/firstboot/complete", s.handleFirstBootComplete)
	mux.HandleFunc("/api/ui/help", s.handleUIHelp)
	mux.HandleFunc("/api/ui/scorecard", s.handleUIScorecard)
	// Admin ve ayar endpointleri
	mux.HandleFunc("/api/admin/smoke", s.handleAdminSmoke)
	mux.HandleFunc("/api/admin/restart", s.handleAdminRestart)
	mux.HandleFunc("/api/admin/backup", s.handleAdminBackup)
	mux.HandleFunc("/api/admin/restore", s.handleAdminRestore)
	mux.HandleFunc("/api/admin/telemetry/summary", s.handleTelemetrySummary)
	mux.HandleFunc("/api/admin/telemetry/optin", s.handleTelemetryOptIn)
	mux.HandleFunc("/api/admin/update/status", s.handleUpdateStatus)
	mux.HandleFunc("/api/admin/update/stage", s.handleUpdateStage)
	mux.HandleFunc("/api/settings/homeassistant", s.handleHASettingsSave)
	mux.HandleFunc("/api/settings/homeassistant/status", s.handleHASettingsStatus)
	mux.HandleFunc("/api/settings/homeassistant/test", s.handleHASettingsTest)
	mux.HandleFunc("/api/settings/homeassistant/sync", s.handleHAInitialSync)
	mux.HandleFunc("/api/devices/lights", s.handleDevicesLights)
	mux.HandleFunc("/api/devices/lights/toggle", s.handleDevicesLightsToggle)
	mux.HandleFunc("/api/devices/lights/set", s.handleDevicesLightsSet)

	// Static file handler (EN SONDA ve sadece bir kez)
	webDir := filepath.Join(os.Getenv("PWD"), "web")
	if _, err := os.Stat(webDir); err != nil {
		webDir = "web"
	}
	fs := http.FileServer(http.Dir(webDir))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if filepath.Ext(r.URL.Path) == ".js" || filepath.Ext(r.URL.Path) == ".css" || filepath.Ext(r.URL.Path) == ".html" {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, max-age=0")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			w.Header().Del("ETag")
			w.Header().Del("Last-Modified")
		}
		fs.ServeHTTP(w, r)
	}))

	return mux
}
