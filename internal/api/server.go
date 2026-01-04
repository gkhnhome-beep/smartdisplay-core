package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"smartdisplay-core/internal/alarm"
	"smartdisplay-core/internal/audit"
	"smartdisplay-core/internal/auth"
	"smartdisplay-core/internal/config"
	"smartdisplay-core/internal/contexthelp"
	"smartdisplay-core/internal/firstboot"
	"smartdisplay-core/internal/logbook"
	"smartdisplay-core/internal/logger"
	"smartdisplay-core/internal/settings"
	"smartdisplay-core/internal/system"
	"smartdisplay-core/internal/telemetry"
	"smartdisplay-core/internal/update"
	"strings"
	"sync"
	"time"
)

// handleAIAnomalies returns grouped anomaly packets for UI
func (s *Server) handleAIAnomalies(w http.ResponseWriter, r *http.Request) {
	packets := s.coord.AI.GroupAnomalies()
	s.respond(w, true, packets, "", 200)
}

// handleAIDaily returns the daily AI summary for UI card
func (s *Server) handleAIDaily(w http.ResponseWriter, r *http.Request) {
	summary := s.coord.AI.GetDailySummary()
	s.respond(w, true, map[string]string{"summary": summary}, "", 200)
}

// getRole extracts the role from the X-User-Role header, defaults to guest if missing/invalid
func getRole(r *http.Request) auth.Role {
	role := strings.ToLower(r.Header.Get("X-User-Role"))
	switch role {
	case "admin":
		return auth.Admin
	case "user":
		return auth.User
	case "guest", "":
		return auth.Guest
	default:
		return auth.Guest
	}
}

// checkPerm enforces permission, logs/audits the decision, and returns role/allowed
func (s *Server) checkPerm(w http.ResponseWriter, r *http.Request, perm auth.Permission) (auth.Role, bool) {
	role := getRole(r)
	allowed := auth.HasPermission(role, perm)
	msg := fmt.Sprintf("role=%s perm=%s allowed=%v", role, perm, allowed)
	log.Println("perm decision:", msg)
	audit.Record("perm_check", msg)
	if !allowed {
		s.respondError(w, r, CodeForbidden, "insufficient permissions")
	}
	return role, allowed
}

type Server struct {
	coord       *system.Coordinator
	httpServer  *http.Server
	mu          sync.Mutex
	telemetry   *telemetry.Collector
	updateMgr   *update.Manager
	shutdownCtx context.Context
	shutdownCxl context.CancelFunc
}

type envelope struct {
	Ok    bool        `json:"ok"`
	Data  interface{} `json:"data,omitempty"`
	Error *string     `json:"error"`
}

func NewServer(coord *system.Coordinator) *Server {
	// Initialize telemetry with data directory
	tel := telemetry.New("data")
	// Try to load existing telemetry state
	_ = tel.LoadState()

	// Initialize update manager
	auditLogger := &UpdateAuditLogger{}
	updateMgr := update.New("1.0.0", "data/staging", auditLogger)

	// Create shutdown context (will be cancelled on graceful shutdown)
	ctx, cancel := context.WithCancel(context.Background())

	return &Server{
		coord:       coord,
		telemetry:   tel,
		updateMgr:   updateMgr,
		shutdownCtx: ctx,
		shutdownCxl: cancel,
	}
}

func (s *Server) Start(port int) error {
	return s.startHTTPServer(port)
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.httpServer != nil {
		return s.httpServer.Close()
	}
	return nil
}

func (s *Server) respond(w http.ResponseWriter, ok bool, data interface{}, errMsg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	var errPtr *string
	if errMsg != "" {
		errPtr = &errMsg
	}
	// Always include failsafe state in envelope
	failsafe := map[string]interface{}{
		"active":      s.coord.InFailsafeMode(),
		"explanation": s.coord.FailsafeExplanation(),
	}
	resp := envelope{Ok: ok, Data: data, Error: errPtr}
	// Wrap in outer object with failsafe
	out := map[string]interface{}{
		"response": resp,
		"failsafe": failsafe,
	}
	_ = json.NewEncoder(w).Encode(out)
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.coord.UpdateFailsafeState()
	res := s.coord.SelfCheck()
	s.respond(w, true, res, "", 200)
}

func (s *Server) handleOverview(w http.ResponseWriter, r *http.Request) {
	_, allowed := s.checkPerm(w, r, auth.PermAlarm)
	if !allowed {
		return
	}
	s.coord.UpdateFailsafeState()

	// Load accessibility and voice preferences
	runtimeCfg, err := config.LoadRuntimeConfig()
	var a11y map[string]bool
	var voiceCfg map[string]bool
	if err == nil {
		a11y = map[string]bool{
			"high_contrast":  runtimeCfg.HighContrast,
			"large_text":     runtimeCfg.LargeText,
			"reduced_motion": runtimeCfg.ReducedMotion,
		}
		voiceCfg = map[string]bool{
			"voice_enabled": runtimeCfg.VoiceEnabled,
		}
	} else {
		a11y = map[string]bool{
			"high_contrast":  false,
			"large_text":     false,
			"reduced_motion": false,
		}
		voiceCfg = map[string]bool{
			"voice_enabled": false,
		}
	}

	overview := map[string]interface{}{
		"alarm":         s.coord.Alarm.CurrentState(),
		"guest":         s.coord.Guest.CurrentState(),
		"ha":            s.coord.HA.IsConnected(),
		"ai":            s.coord.GetCurrentInsight(),
		"accessibility": a11y,
		"voice":         voiceCfg,
	}
	s.respond(w, true, overview, "", 200)
}

// Dedicated failsafe endpoint
func (s *Server) handleFailsafe(w http.ResponseWriter, r *http.Request) {
	s.coord.UpdateFailsafeState()
	state := map[string]interface{}{
		"active":      s.coord.InFailsafeMode(),
		"explanation": s.coord.FailsafeExplanation(),
	}
	s.respond(w, true, state, "", 200)
}

func (s *Server) handleAlarmArm(w http.ResponseWriter, r *http.Request) {
	_, allowed := s.checkPerm(w, r, auth.PermAlarm)
	if !allowed {
		return
	}
	if r.Method != "POST" {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
		return
	}
	err := s.coord.Alarm.Handle("ARM_REQUEST")
	if err != nil {
		s.respondError(w, r, CodeInternalError, "alarm arm error")
		return
	}
	s.respond(w, true, map[string]string{"result": "ok"}, "", 200)
}

func (s *Server) handleAlarmDisarm(w http.ResponseWriter, r *http.Request) {
	_, allowed := s.checkPerm(w, r, auth.PermAlarm)
	if !allowed {
		return
	}
	if r.Method != "POST" {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
		return
	}
	err := s.coord.Alarm.Handle("DISARM_REQUEST")
	if err != nil {
		s.respondError(w, r, CodeInternalError, "alarm disarm error")
		return
	}
	s.respond(w, true, map[string]string{"result": "ok"}, "", 200)
}

func (s *Server) handleGuestApprove(w http.ResponseWriter, r *http.Request) {
	role, allowed := s.checkPerm(w, r, auth.PermGuest)
	if !allowed || role != auth.Admin {
		if allowed {
			s.respondError(w, r, CodeForbidden, "admin required")
		}
		return
	}
	if r.Method != "POST" {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
		return
	}
	err := s.coord.Guest.Handle("APPROVE")
	if err != nil {
		s.respondError(w, r, CodeInternalError, "guest approve error")
		return
	}
	s.respond(w, true, map[string]string{"result": "ok"}, "", 200)
}

func (s *Server) handleGuestDeny(w http.ResponseWriter, r *http.Request) {
	role, allowed := s.checkPerm(w, r, auth.PermGuest)
	if !allowed || role != auth.Admin {
		if allowed {
			s.respondError(w, r, CodeForbidden, "admin required")
		}
		return
	}
	if r.Method != "POST" {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
		return
	}
	err := s.coord.Guest.Handle("DENY")
	if err != nil {
		s.respondError(w, r, CodeInternalError, "guest deny error")
		return
	}
	s.respond(w, true, map[string]string{"result": "ok"}, "", 200)
}

func (s *Server) handleMenu(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.respondError(w, r, CodeMethodNotAllowed, "GET required")
		return
	}
	if s.coord.Menu == nil {
		s.respondError(w, r, CodeInternalError, "menu manager not initialized")
		return
	}
	userID := r.Header.Get("X-User-ID")
	menuResp := s.coord.Menu.ResolveMenu(userID)
	s.respond(w, true, menuResp, "", 200)
}

func (s *Server) handleAIInsight(w http.ResponseWriter, r *http.Request) {
	insight := s.coord.GetCurrentInsight()
	s.respond(w, true, insight, "", 200)
}

func (s *Server) handleAIExplain(w http.ResponseWriter, r *http.Request) {
	explanation := s.coord.ExplainInsight()
	s.respond(w, true, map[string]string{"explanation": explanation}, "", 200)
}

func (s *Server) handleAIHistory(w http.ResponseWriter, r *http.Request) {
	history := s.coord.AI.GetInsightHistory()
	s.respond(w, true, history, "", 200)
}

// handleUIHelp returns deterministic, action-oriented help for the current UI screen and state
func (s *Server) handleUIHelp(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.respondError(w, r, CodeMethodNotAllowed, "GET required")
		return
	}
	screen := r.URL.Query().Get("screen")
	// Optionally, parse state from query or body (for now, just screen)
	req := struct {
		Screen string                 `json:"screen"`
		State  map[string]interface{} `json:"state"`
	}{Screen: screen, State: map[string]interface{}{}}
	// Optionally, parse state from JSON body if POST is ever supported
	// For now, only screen is used
	help := contexthelp.GenerateHelp(contexthelp.HelpRequest{
		Screen: req.Screen,
		State:  req.State,
	})
	s.respond(w, true, help, "", 200)
}

// handleAIMorning returns a single useful morning message (briefing) for the user
func (s *Server) handleAIMorning(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.respondError(w, r, CodeMethodNotAllowed, "GET required")
		return
	}
	// TODO: Implement morning briefing handler with proper dependency methods
	// Requires: Guest.ExpectedToday(), Coordinator.HasIssues()
	s.respondError(w, r, CodeInternalError, "not yet implemented")
}

// handleUIScorecard returns a simple system quality scorecard for UI display
func (s *Server) handleUIScorecard(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.respondError(w, r, CodeMethodNotAllowed, "GET required")
		return
	}
	// TODO: Implement scorecard handler with proper dependency methods
	// Requires: Alarm.IsArmed(), Coordinator.HasIssues()
	s.respondError(w, r, CodeInternalError, "not yet implemented")
}

// handleTelemetrySummary returns aggregated telemetry summary (admin-only)
// GET /api/admin/telemetry/summary
// Returns: aggregated feature usage, error categories, and performance buckets (no personal data)
func (s *Server) handleTelemetrySummary(w http.ResponseWriter, r *http.Request) {
	role := getRole(r)
	if role != auth.Admin {
		s.respondError(w, r, CodeForbidden, "admin required")
		return
	}
	if r.Method != "GET" {
		s.respondError(w, r, CodeMethodNotAllowed, "GET required")
		return
	}
	summary := s.telemetry.GetSummary()
	s.respond(w, true, summary, "", 200)
}

// handleTelemetryOptIn enables or disables telemetry opt-in (admin-only)
// POST /api/admin/telemetry/optin
// Request body: {"enabled": bool}
// Opt-in is disabled by default and must be explicitly enabled by admin.
func (s *Server) handleTelemetryOptIn(w http.ResponseWriter, r *http.Request) {
	role := getRole(r)
	if role != auth.Admin {
		s.respondError(w, r, CodeForbidden, "admin required")
		return
	}
	if r.Method != "POST" {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
		return
	}
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, r, CodeBadRequest, "invalid json")
		return
	}
	s.telemetry.SetOptIn(req.Enabled)
	// Persist the state
	if err := s.telemetry.Flush(); err != nil {
		audit.Record("telemetry_error", "failed to persist telemetry state: "+err.Error())
	}
	status := "disabled"
	if req.Enabled {
		status = "enabled"
	}
	audit.Record("telemetry_optin", "admin set telemetry to "+status)
	s.respond(w, true, map[string]interface{}{
		"enabled": req.Enabled,
		"message": "telemetry " + status,
	}, "", 200)
}

// UpdateAuditLogger logs update actions to the audit trail.
type UpdateAuditLogger struct{}

func (u *UpdateAuditLogger) Record(action string, detail string) {
	audit.Record("update_"+action, detail)
}

// handleUpdateStatus returns current update system status (admin-only)
// GET /api/admin/update/status
// Returns: current version, available updates, staged updates, pending reboot state
func (s *Server) handleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	role := getRole(r)
	if role != auth.Admin {
		s.respondError(w, r, CodeForbidden, "admin required")
		return
	}
	if r.Method != "GET" {
		s.respondError(w, r, CodeMethodNotAllowed, "GET required")
		return
	}
	status := s.updateMgr.GetStatus()
	audit.Record("update_status_check", fmt.Sprintf("admin checked update status (version=%s, pending=%v)", status.CurrentVersion, status.PendingReboot))
	s.respond(w, true, status, "", 200)
}

// handleUpdateStage stages an update package for later activation (admin-only, NO-OP STUB)
// POST /api/admin/update/stage
// Request body: {"package": {...}, "data": "base64-encoded-package-data"}
// NOTE: This is currently a STUB. Package data is not actually downloaded or written.
// In future phases, this will accept pre-validated packages and stage them to disk.
func (s *Server) handleUpdateStage(w http.ResponseWriter, r *http.Request) {
	role := getRole(r)
	if role != auth.Admin {
		s.respondError(w, r, CodeForbidden, "admin required")
		return
	}
	if r.Method != "POST" {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
		return
	}

	var req struct {
		Package struct {
			Version  string `json:"version"`
			BuildID  string `json:"build_id"`
			Checksum string `json:"checksum"`
			Size     int64  `json:"size"`
		} `json:"package"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, r, CodeBadRequest, "invalid json")
		audit.Record("update_stage_error", "invalid request body")
		return
	}

	// STUB: No actual staging implemented yet
	// In future phases:
	// 1. Download package from server
	// 2. Validate checksum
	// 3. Write to staging directory
	// 4. Return staged path

	audit.Record("update_stage_stub", fmt.Sprintf("admin requested staging of version %s (no-op stub)", req.Package.Version))
	s.respond(w, true, map[string]interface{}{
		"message": "update staging is a no-op stub in this phase",
		"version": req.Package.Version,
		"status":  "stub_response",
	}, "", 200)
}

// handleAccessibility handles GET/POST for accessibility preferences (FAZ 80)
// GET: Returns current accessibility preferences
// POST: Updates accessibility preferences with validation
func (s *Server) handleAccessibility(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.handleAccessibilityGet(w, r)
		return
	} else if r.Method == http.MethodPost {
		s.handleAccessibilityPost(w, r)
		return
	}
	s.respondError(w, r, CodeMethodNotAllowed, "GET or POST required")
}

// handleAccessibilityGet returns current accessibility preferences
func (s *Server) handleAccessibilityGet(w http.ResponseWriter, r *http.Request) {
	runtimeCfg, err := config.LoadRuntimeConfig()
	if err != nil {
		logger.Error("failed to load runtime config: " + err.Error())
		s.respondError(w, r, CodeInternalError, "failed to load preferences")
		return
	}

	prefs := map[string]interface{}{
		"high_contrast":  runtimeCfg.HighContrast,
		"large_text":     runtimeCfg.LargeText,
		"reduced_motion": runtimeCfg.ReducedMotion,
	}

	s.respond(w, true, prefs, "", 200)
}

// handleAccessibilityPost updates accessibility preferences
func (s *Server) handleAccessibilityPost(w http.ResponseWriter, r *http.Request) {
	var req struct {
		HighContrast  *bool `json:"high_contrast"`
		LargeText     *bool `json:"large_text"`
		ReducedMotion *bool `json:"reduced_motion"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, r, CodeBadRequest, "invalid json")
		return
	}

	// Load current config
	runtimeCfg, err := config.LoadRuntimeConfig()
	if err != nil {
		logger.Error("failed to load runtime config: " + err.Error())
		s.respondError(w, r, CodeInternalError, "failed to load preferences")
		return
	}

	// Track changes for logging
	changes := []string{}

	// Update only provided fields (atomic update)
	if req.HighContrast != nil && *req.HighContrast != runtimeCfg.HighContrast {
		runtimeCfg.HighContrast = *req.HighContrast
		changes = append(changes, fmt.Sprintf("high_contrast=%v", *req.HighContrast))
	}
	if req.LargeText != nil && *req.LargeText != runtimeCfg.LargeText {
		runtimeCfg.LargeText = *req.LargeText
		changes = append(changes, fmt.Sprintf("large_text=%v", *req.LargeText))
	}
	if req.ReducedMotion != nil && *req.ReducedMotion != runtimeCfg.ReducedMotion {
		runtimeCfg.ReducedMotion = *req.ReducedMotion
		changes = append(changes, fmt.Sprintf("reduced_motion=%v", *req.ReducedMotion))
	}

	// Save updated config
	if err := config.SaveRuntimeConfig(runtimeCfg); err != nil {
		logger.Error("failed to save runtime config: " + err.Error())
		s.respondError(w, r, CodeInternalError, "failed to save preferences")
		return
	}

	// Log changes
	if len(changes) > 0 {
		logger.Info("accessibility preferences updated: " + strings.Join(changes, ", "))
		audit.Record("accessibility", "preferences updated: "+strings.Join(changes, ", "))
	}

	// TODO: Update coordinator's config if it has accessibility awareness
	// (For reduced_motion to affect AI behavior)
	// Requires: Coordinator.UpdateAccessibilityPreferences() method

	s.respond(w, true, map[string]interface{}{
		"high_contrast":  runtimeCfg.HighContrast,
		"large_text":     runtimeCfg.LargeText,
		"reduced_motion": runtimeCfg.ReducedMotion,
	}, "", 200)
}

// handleVoice handles GET/POST for voice feedback preferences (FAZ 81)
// GET: Returns current voice feedback state
// POST: Updates voice_enabled setting
func (s *Server) handleVoice(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.handleVoiceGet(w, r)
	} else if r.Method == http.MethodPost {
		s.handleVoicePost(w, r)
	} else {
		s.respondError(w, r, CodeMethodNotAllowed, "GET or POST required")
	}
}

// handleVoiceGet returns current voice feedback state
func (s *Server) handleVoiceGet(w http.ResponseWriter, r *http.Request) {
	runtimeCfg, err := config.LoadRuntimeConfig()
	if err != nil {
		s.respondError(w, r, CodeInternalError, "failed to load config")
		return
	}

	s.respond(w, true, map[string]interface{}{
		"voice_enabled": runtimeCfg.VoiceEnabled,
	}, "", 200)
}

// handleVoicePost updates voice feedback preferences
func (s *Server) handleVoicePost(w http.ResponseWriter, r *http.Request) {
	var req struct {
		VoiceEnabled *bool `json:"voice_enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, r, CodeBadRequest, "invalid request")
		return
	}

	// Load current config
	runtimeCfg, err := config.LoadRuntimeConfig()
	if err != nil {
		s.respondError(w, r, CodeInternalError, "failed to load config")
		return
	}

	// Atomic update: only change provided fields
	if req.VoiceEnabled != nil {
		if *req.VoiceEnabled != runtimeCfg.VoiceEnabled {
			runtimeCfg.VoiceEnabled = *req.VoiceEnabled
			logger.Info("voice preferences updated: voice_enabled=" + fmt.Sprintf("%v", *req.VoiceEnabled))
		}
	}

	// Save updated config
	if err := config.SaveRuntimeConfig(runtimeCfg); err != nil {
		s.respondError(w, r, CodeInternalError, "failed to save config")
		return
	}

	// TODO: Apply to coordinator's voice hook
	// Requires: Coordinator.Voice field and SetEnabled() method

	s.respond(w, true, map[string]interface{}{
		"voice_enabled": runtimeCfg.VoiceEnabled,
	}, "", 200)
}

// handleFirstBootStatus returns the current first-boot status and step info (D0)
func (s *Server) handleFirstBootStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, r, CodeMethodNotAllowed, "GET required")
		return
	}

	if s.coord.FirstBoot == nil {
		s.respondError(w, r, CodeInternalError, "first-boot manager not initialized")
		return
	}

	s.respond(w, true, s.coord.FirstBoot.AllStepsStatus(), "", 200)
}

// handleFirstBootNext advances to the next step (D0)
func (s *Server) handleFirstBootNext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
		return
	}

	if s.coord.FirstBoot == nil {
		s.respondError(w, r, CodeInternalError, "first-boot manager not initialized")
		return
	}

	success, err := s.coord.FirstBoot.Next()
	if !success {
		s.respondError(w, r, CodeBadRequest, err.Error())
		return
	}

	s.respond(w, true, s.coord.FirstBoot.AllStepsStatus(), "", 200)
}

// handleFirstBootBack returns to the previous step (D0)
func (s *Server) handleFirstBootBack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
		return
	}

	if s.coord.FirstBoot == nil {
		s.respondError(w, r, CodeInternalError, "first-boot manager not initialized")
		return
	}

	success, err := s.coord.FirstBoot.Back()
	if !success {
		s.respondError(w, r, CodeBadRequest, err.Error())
		return
	}

	s.respond(w, true, s.coord.FirstBoot.AllStepsStatus(), "", 200)
}

// handleFirstBootComplete marks first-boot as complete and updates config (D0)
func (s *Server) handleFirstBootComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
		return
	}

	if s.coord.FirstBoot == nil {
		s.respondError(w, r, CodeInternalError, "first-boot manager not initialized")
		return
	}

	success, err := s.coord.FirstBoot.Complete()
	if !success {
		s.respondError(w, r, CodeBadRequest, err.Error())
		return
	}

	// Persist completion to config
	if err := firstboot.SaveCompletion(true); err != nil {
		logger.Error("failed to save first-boot completion: " + err.Error())
		s.respondError(w, r, CodeInternalError, "failed to save completion")
		return
	}

	s.respond(w, true, map[string]interface{}{
		"wizard_completed": true,
		"status":           s.coord.FirstBoot.AllStepsStatus(),
	}, "", 200)
}

// === HOME SCREEN ENDPOINTS (D2) ===

// handleHomeState returns full home screen state with all contextual data (D2)
func (s *Server) handleHomeState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, r, CodeMethodNotAllowed, "GET required")
		return
	}

	if s.coord.Home == nil {
		s.respondError(w, r, CodeInternalError, "home manager not initialized")
		return
	}

	resp := s.coord.Home.GetStateResponse()
	s.respond(w, true, resp, "", 200)
}

// handleHomeSummary returns lightweight summary-only data (D2)
func (s *Server) handleHomeSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, r, CodeMethodNotAllowed, "GET required")
		return
	}

	if s.coord.Home == nil {
		s.respondError(w, r, CodeInternalError, "home manager not initialized")
		return
	}

	summary := s.coord.Home.GetSummaryResponse()
	s.respond(w, true, summary, "", 200)
}

// === ALARM SCREEN ENDPOINTS (D3) ===

// handleAlarmState returns full alarm screen state with all contextual data (D3)
// A2: Uses Alarmo as source of truth instead of internal alarm state
func (s *Server) handleAlarmState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, r, CodeMethodNotAllowed, "GET required")
		return
	}

	if s.coord.AlarmScreen == nil {
		s.respondError(w, r, CodeInternalError, "alarm screen manager not initialized")
		return
	}

	// A2: Use Alarmo state as source of truth
	s.coord.AlarmoMu.RLock()
	alarmoState := s.coord.AlarmoState
	s.coord.AlarmoMu.RUnlock()

	// Get screen state (will use Alarmo-derived mode if available)
	state := s.coord.AlarmScreen.GetScreenState()

	// A2: Override mode from Alarmo for accuracy
	if alarmoState.Mode != "" {
		switch alarmoState.Mode {
		case "disarmed":
			state.Mode = alarm.ModeDisarmed
		case "arming":
			state.Mode = alarm.ModeArming
		case "armed":
			state.Mode = alarm.ModeArmed
		case "triggered":
			state.Mode = alarm.ModeTriggered
		}
	}

	lastUpdated := ""
	if !alarmoState.LastChanged.IsZero() {
		lastUpdated = alarmoState.LastChanged.Format(time.RFC3339)
	}

	// A3: Expose Alarmo countdown/trigger metadata alongside screen state
	payload := struct {
		*alarm.ScreenState
		Alarmo map[string]interface{} `json:"alarmo"`
	}{
		ScreenState: state,
		Alarmo: map[string]interface{}{
			"state":      alarmoState.Mode,
			"armed_mode": alarmoState.ArmedMode,
			"triggered":  alarmoState.Triggered,
			"delay": map[string]interface{}{
				"type":      alarmoState.DelayType,
				"remaining": alarmoState.DelayRemaining,
			},
			"last_updated": lastUpdated,
		},
	}

	s.respond(w, true, payload, "", 200)
}

// handleAlarmSummary returns lightweight alarm summary for frequent polling (D3)
// A2: Uses Alarmo as source of truth instead of internal alarm state
func (s *Server) handleAlarmSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, r, CodeMethodNotAllowed, "GET required")
		return
	}

	if s.coord.AlarmScreen == nil {
		s.respondError(w, r, CodeInternalError, "alarm screen manager not initialized")
		return
	}

	// A2: Use Alarmo state as source of truth
	s.coord.AlarmoMu.RLock()
	alarmoState := s.coord.AlarmoState
	s.coord.AlarmoMu.RUnlock()

	summary := s.coord.AlarmScreen.GetSummaryState()

	// A2: Override mode from Alarmo for accuracy
	if alarmoState.Mode != "" {
		switch alarmoState.Mode {
		case "disarmed":
			summary.Mode = alarm.ModeDisarmed
		case "arming":
			summary.Mode = alarm.ModeArming
		case "armed":
			summary.Mode = alarm.ModeArmed
		case "triggered":
			summary.Mode = alarm.ModeTriggered
		}
	}

	s.respond(w, true, summary, "", 200)
}

// handleAlarmAction handles controlled arm/disarm requests to Alarmo (A4)
// POST /api/ui/alarm/action
// Request: {"action": "arm_home | arm_away | arm_night | disarm"}
// This is the FIRST write operation - fully controlled and audited
func (s *Server) handleAlarmAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
		return
	}

	// Parse request body
	var req struct {
		Action string `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, r, CodeBadRequest, "invalid JSON")
		return
	}

	// Validate action
	validActions := map[string]bool{
		"arm_home":  true,
		"arm_away":  true,
		"arm_night": true,
		"disarm":    true,
	}

	if !validActions[req.Action] {
		s.respondError(w, r, CodeBadRequest, "invalid action")
		return
	}

	// Send request to coordinator (does NOT modify local state)
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	err := s.coord.RequestAlarmAction(ctx, req.Action)
	if err != nil {
		// Check error type for appropriate status code
		errMsg := err.Error()
		if strings.Contains(errMsg, "unreachable") {
			s.respondError(w, r, CodeServiceUnavailable, "alarmo unreachable")
			return
		}
		if strings.Contains(errMsg, "triggered") {
			s.respondError(w, r, CodeConflict, "action blocked: system triggered")
			return
		}
		s.respondError(w, r, CodeInternalError, "action request failed")
		return
	}

	// Success - but state change will appear via polling
	s.respond(w, true, map[string]string{
		"status":  "requested",
		"message": "action sent to alarmo, state will update via polling",
	}, "", 200)
}

// === GUEST ACCESS ENDPOINTS (D4) ===

// handleGuestState returns full guest access state with all contextual data (D4)
func (s *Server) handleGuestState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, r, CodeMethodNotAllowed, "GET required")
		return
	}

	if s.coord.GuestScreen == nil {
		s.respondError(w, r, CodeInternalError, "guest screen manager not initialized")
		return
	}

	state := s.coord.GuestScreen.GetScreenState()
	s.respond(w, true, state, "", 200)
}

// handleGuestSummary returns lightweight guest summary for frequent polling (D4)
func (s *Server) handleGuestSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, r, CodeMethodNotAllowed, "GET required")
		return
	}

	if s.coord.GuestScreen == nil {
		s.respondError(w, r, CodeInternalError, "guest screen manager not initialized")
		return
	}

	summary := s.coord.GuestScreen.GetSummaryState()
	s.respond(w, true, summary, "", 200)
}

// handleGuestRequest initiates a new guest access request (D4)
func (s *Server) handleGuestRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
		return
	}

	if s.coord.GuestScreen == nil {
		s.respondError(w, r, CodeInternalError, "guest screen manager not initialized")
		return
	}

	// Generate guest ID (in real implementation, from request context)
	guestID := "guest_unknown"

	s.coord.GuestScreen.OnRequestInitiated(guestID)
	state := s.coord.GuestScreen.GetScreenState()
	s.respond(w, true, state, "", 200)
}

// handleGuestExit processes guest exit and alarm re-arming (D4)
func (s *Server) handleGuestExit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
		return
	}

	if s.coord.GuestScreen == nil {
		s.respondError(w, r, CodeInternalError, "guest screen manager not initialized")
		return
	}

	s.coord.GuestScreen.OnExit()
	state := s.coord.GuestScreen.GetScreenState()
	s.respond(w, true, state, "", 200)
}

func (s *Server) handleLogbook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.respondError(w, r, CodeMethodNotAllowed, "GET required")
		return
	}
	if s.coord.Logbook == nil {
		s.respondError(w, r, CodeInternalError, "logbook manager not initialized")
		return
	}

	// Parse query parameters
	limit := 20
	offset := 0
	categoryFilter := ""

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		fmt.Sscanf(offsetStr, "%d", &offset)
	}
	if cat := r.URL.Query().Get("category"); cat != "" {
		categoryFilter = cat
	}

	userID := r.Header.Get("X-User-ID")
	userRole := r.Header.Get("X-User-Role")

	// Map user role string to logbook role type
	var logbookRole logbook.UserRole
	if userRole == "admin" {
		logbookRole = logbook.RoleAdmin
	} else if userRole == "user" {
		logbookRole = logbook.RoleUser
	} else {
		logbookRole = logbook.RoleGuest
	}

	response := s.coord.Logbook.GetEntries(logbookRole, limit, offset, categoryFilter)
	response.UserID = userID
	response.Role = logbookRole

	s.respond(w, true, response, "", 200)
}

func (s *Server) handleLogbookSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.respondError(w, r, CodeMethodNotAllowed, "GET required")
		return
	}
	if s.coord.Logbook == nil {
		s.respondError(w, r, CodeInternalError, "logbook manager not initialized")
		return
	}

	// Parse query parameters
	limit := 5
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}

	userRole := r.Header.Get("X-User-Role")

	// Map user role string to logbook role type
	var logbookRole logbook.UserRole
	if userRole == "admin" {
		logbookRole = logbook.RoleAdmin
	} else if userRole == "user" {
		logbookRole = logbook.RoleUser
	} else {
		logbookRole = logbook.RoleGuest
	}

	response := s.coord.Logbook.GetSummary(logbookRole, limit)
	s.respond(w, true, response, "", 200)
}

// === SETTINGS HANDLERS (D7) ===

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.respondError(w, r, CodeMethodNotAllowed, "GET required")
		return
	}
	if s.coord.Settings == nil {
		s.respondError(w, r, CodeInternalError, "settings manager not initialized")
		return
	}

	userRole := r.Header.Get("X-User-Role")

	// Map user role string to settings role type
	var settingsRole string
	if userRole == "admin" {
		settingsRole = "admin"
	} else if userRole == "user" {
		settingsRole = "user"
	} else {
		settingsRole = "guest"
	}

	// Only admin can access settings
	if settingsRole != "admin" {
		s.respondError(w, r, CodeForbidden, "admin required")
		return
	}

	response, err := s.coord.Settings.GetSettings()
	if err != nil {
		s.respondError(w, r, CodeBadRequest, err.Error())
		return
	}

	s.respond(w, true, response, "", 200)
}

func (s *Server) handleSettingsAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
		return
	}
	if s.coord.Settings == nil {
		s.respondError(w, r, CodeInternalError, "settings manager not initialized")
		return
	}

	userRole := r.Header.Get("X-User-Role")

	// Only admin can execute settings actions
	if userRole != "admin" {
		s.respondError(w, r, CodeForbidden, "admin required")
		return
	}

	// Parse request body
	var reqBody map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		s.respondError(w, r, CodeBadRequest, "invalid request body")
		return
	}

	action, ok := reqBody["action"].(string)
	if !ok {
		s.respondError(w, r, CodeBadRequest, "missing action field")
		return
	}

	// Handle field changes
	if action == "field_change" {
		fieldID, ok := reqBody["field_id"].(string)
		if !ok {
			s.respondError(w, r, CodeBadRequest, "missing field_id")
			return
		}

		newValue := reqBody["new_value"]
		confirm := false
		if confirmVal, ok := reqBody["confirm"].(bool); ok {
			confirm = confirmVal
		}

		req := &settings.FieldChangeRequest{
			Action:   settings.ActionFieldChange,
			FieldID:  fieldID,
			NewValue: newValue,
			Confirm:  confirm,
		}

		response, err := s.coord.Settings.ApplyFieldChange(req)
		if err != nil {
			s.respondError(w, r, CodeBadRequest, err.Error())
			return
		}

		s.respond(w, true, response, "", 200)
		return
	}

	// Handle actions (restart, backup, factory reset)
	actionReq := &settings.ActionRequest{
		Action:  settings.ActionType(action),
		Confirm: false,
	}

	if confirmVal, ok := reqBody["confirm"].(bool); ok {
		actionReq.Confirm = confirmVal
	}

	if backupID, ok := reqBody["backup_id"].(string); ok {
		actionReq.BackupID = backupID
	}

	if confirmType, ok := reqBody["confirm_type"].(string); ok {
		actionReq.ConfirmType = settings.ConfirmationType(confirmType)
	}

	if confirmText, ok := reqBody["confirm_text"].(string); ok {
		actionReq.ConfirmText = confirmText
	}

	response, err := s.coord.Settings.ApplyAction(actionReq)
	if err != nil {
		s.respondError(w, r, CodeBadRequest, err.Error())
		return
	}

	s.respond(w, true, response, "", 200)
}
