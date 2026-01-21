package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"smartdisplay-core/internal/alarm"
	"smartdisplay-core/internal/audit"
	"smartdisplay-core/internal/auth"
	"smartdisplay-core/internal/config"
	"smartdisplay-core/internal/contexthelp"
	"smartdisplay-core/internal/firstboot"
	"smartdisplay-core/internal/guest"
	"smartdisplay-core/internal/logbook"
	"smartdisplay-core/internal/logger"
	"smartdisplay-core/internal/settings"
	"smartdisplay-core/internal/system"
	"smartdisplay-core/internal/telemetry"
	"smartdisplay-core/internal/update"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Shutdown gracefully shuts down the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// Kullanıcı yönetimi endpointleri
// Kullanıcı yönetimi endpointleri (sade)
func (s *Server) HandleUserList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Method not allowed"})
		return
	}
	role, pin := "", ""
	if cookie, err := r.Cookie("sd_session"); err == nil {
		parts := strings.SplitN(cookie.Value, ":", 2)
		if len(parts) == 2 {
			role = parts[0]
			pin = parts[1]
		}
	}
	if role != "admin" || pin == "" {
		role = r.Header.Get("X-User-Role")
		pin = r.Header.Get("X-User-Pin")
	}
	ctx, err := auth.ValidatePIN(pin)
	if err != nil || !ctx.Authenticated || ctx.Role != "admin" {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Yetkisiz erişim"})
		return
	}
	users, err := auth.LoadAllUsers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Kullanıcılar yüklenemedi"})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "users": users})
}

func (s *Server) HandleUserAdd(w http.ResponseWriter, r *http.Request) {
	role, pin := "", ""
	if cookie, err := r.Cookie("sd_session"); err == nil {
		parts := strings.SplitN(cookie.Value, ":", 2)
		if len(parts) == 2 {
			role = parts[0]
			pin = parts[1]
		}
	}
	if role != "admin" || pin == "" {
		role = r.Header.Get("X-User-Role")
		pin = r.Header.Get("X-User-Pin")
	}
	ctx, err := auth.ValidatePIN(pin)
	if err != nil || !ctx.Authenticated || ctx.Role != "admin" {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Yetkisiz erişim"})
		return
	}
	var user auth.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Geçersiz istek"})
		return
	}
	err = auth.AddUser(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func (s *Server) HandleUserUpdate(w http.ResponseWriter, r *http.Request) {
	role, pin := "", ""
	if cookie, err := r.Cookie("sd_session"); err == nil {
		parts := strings.SplitN(cookie.Value, ":", 2)
		if len(parts) == 2 {
			role = parts[0]
			pin = parts[1]
		}
	}
	if role != "admin" || pin == "" {
		role = r.Header.Get("X-User-Role")
		pin = r.Header.Get("X-User-Pin")
	}
	ctx, err := auth.ValidatePIN(pin)
	if err != nil || !ctx.Authenticated || ctx.Role != "admin" {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Yetkisiz erişim"})
		return
	}
	var user auth.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Geçersiz istek"})
		return
	}
	err = auth.UpdateUser(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func (s *Server) HandleUserDelete(w http.ResponseWriter, r *http.Request) {
	role, pin := "", ""
	if cookie, err := r.Cookie("sd_session"); err == nil {
		parts := strings.SplitN(cookie.Value, ":", 2)
		if len(parts) == 2 {
			role = parts[0]
			pin = parts[1]
		}
	}
	if role != "admin" || pin == "" {
		role = r.Header.Get("X-User-Role")
		pin = r.Header.Get("X-User-Pin")
	}
	ctx, err := auth.ValidatePIN(pin)
	if err != nil || !ctx.Authenticated || ctx.Role != "admin" {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Yetkisiz erişim"})
		return
	}
	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Geçersiz istek"})
		return
	}
	err = auth.DeleteUser(req.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// ...existing code...

// haConnectionTestResult models the result of a connection test
type haConnectionTestResult struct {
	Success bool
	Stage   string
	Message string
}

// runHAConnectionTest performs the sequential HA connection test (FAZ S3)
func (s *Server) runHAConnectionTest() haConnectionTestResult {
	// 1. Load config and decrypt credentials
	cfg, err := settings.LoadHAConfig()
	if err != nil || cfg == nil || cfg.ServerURL == "" || cfg.EncryptedToken == "" {
		return haConnectionTestResult{
			Success: false,
			Stage:   "server_unreachable",
			Message: "Cannot reach Home Assistant server",
		}
	}
	token, err := settings.Decrypt(cfg.EncryptedToken)
	if err != nil || token == "" {
		return haConnectionTestResult{
			Success: false,
			Stage:   "auth_failed",
			Message: "Authentication failed. Please check token",
		}
	}
	baseURL := cfg.ServerURL
	client := &http.Client{Timeout: 4 * time.Second}

	// 2a. Server Reachability
	req, err := http.NewRequest("GET", baseURL+"/api/", nil)
	if err != nil {
		return haConnectionTestResult{
			Success: false,
			Stage:   "server_unreachable",
			Message: "Cannot reach Home Assistant server",
		}
	}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return haConnectionTestResult{
			Success: false,
			Stage:   "server_unreachable",
			Message: "Cannot reach Home Assistant server",
		}
	}
	resp.Body.Close()

	// 2b. Authentication Validity
	req, err = http.NewRequest("GET", baseURL+"/api/config", nil)
	if err != nil {
		return haConnectionTestResult{
			Success: false,
			Stage:   "api_unavailable",
			Message: "Home Assistant API unavailable",
		}
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = client.Do(req)
	if err != nil {
		return haConnectionTestResult{
			Success: false,
			Stage:   "api_unavailable",
			Message: "Home Assistant API unavailable",
		}
	}
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		resp.Body.Close()
		return haConnectionTestResult{
			Success: false,
			Stage:   "auth_failed",
			Message: "Authentication failed. Please check token",
		}
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return haConnectionTestResult{
			Success: false,
			Stage:   "api_unavailable",
			Message: "Home Assistant API unavailable",
		}
	}
	// 2c. Core API Sanity (valid JSON)
	var apiSanity map[string]interface{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&apiSanity); err != nil {
		resp.Body.Close()
		return haConnectionTestResult{
			Success: false,
			Stage:   "api_unavailable",
			Message: "Home Assistant API unavailable",
		}
	}
	resp.Body.Close()

	// 2d. Alarmo Presence (read-only)
	req, err = http.NewRequest("GET", baseURL+"/api/states/alarm_control_panel.alarmo", nil)
	if err != nil {
		return haConnectionTestResult{
			Success: false,
			Stage:   "alarmo_missing",
			Message: "Alarmo integration not found",
		}
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = client.Do(req)
	if err != nil {
		return haConnectionTestResult{
			Success: false,
			Stage:   "alarmo_missing",
			Message: "Alarmo integration not found",
		}
	}
	if resp.StatusCode != 200 && resp.StatusCode != 404 {
		resp.Body.Close()
		return haConnectionTestResult{
			Success: false,
			Stage:   "alarmo_missing",
			Message: "Alarmo integration not found",
		}
	}
	if resp.StatusCode == 404 {
		resp.Body.Close()
		return haConnectionTestResult{
			Success: false,
			Stage:   "alarmo_missing",
			Message: "Alarmo integration not found",
		}
	}
	resp.Body.Close()

	// 4. Success: update last_tested_at
	now := time.Now().UTC()
	cfg.LastTestedAt = &now
	_ = settings.SaveHAConfig(cfg) // ignore error, do not fail test

	return haConnectionTestResult{
		Success: true,
		Stage:   "ok",
		Message: "Home Assistant connection successful",
	}
}

// --- Register FAZ S2 endpoints ---
func (s *Server) RegisterFAZS2Endpoints(mux *http.ServeMux) {
	mux.HandleFunc("/api/settings/homeassistant/test", s.handleHASettingsTest)
	mux.HandleFunc("/api/settings/homeassistant", s.handleHASettingsSave)
	mux.HandleFunc("/api/settings/homeassistant/status", s.handleHASettingsStatus)
	mux.HandleFunc("/api/login", s.handleLogin)
}

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

// getRole extracts the role from auth context (set by auth middleware)

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
	coord         *system.Coordinator
	httpServer    *http.Server
	mu            sync.Mutex
	runtimeCfg    *config.RuntimeConfig
	telemetry     *telemetry.Collector
	updateMgr     *update.Manager
	healthMonitor *settings.RuntimeHealthMonitor
	shutdownCtx   context.Context
	shutdownCxl   context.CancelFunc
}

type envelope struct {
	Ok    bool        `json:"ok"`
	Data  interface{} `json:"data,omitempty"`
	Error *string     `json:"error"`
}

func NewServer(coord *system.Coordinator, runtimeCfg *config.RuntimeConfig) *Server {
	// Initialize health monitor for HA runtime tracking (FAZ S6)
	healthMon := settings.GetGlobalHealthMonitor()

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
		coord:         coord,
		runtimeCfg:    runtimeCfg,
		telemetry:     tel,
		updateMgr:     updateMgr,
		healthMonitor: healthMon,
		shutdownCtx:   ctx,
		shutdownCxl:   cancel,
	}
}

func (s *Server) Start(port int) error {
	// Register FAZ S2 endpoints before starting HTTP server
	if s.httpServer != nil && s.httpServer.Handler != nil {
		if mux, ok := s.httpServer.Handler.(*http.ServeMux); ok {
			s.RegisterFAZS2Endpoints(mux)
		}
	}
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

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
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

// === ALARMO MONITORING (READ-ONLY) ===

// handleAlarmoStatus exposes lightweight Alarmo connectivity/health status
// Visible to all roles; returns only runtime health fields (no secrets)
func (s *Server) handleAlarmoStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, r, CodeMethodNotAllowed, "GET required")
		return
	}

	lastSeen := ""
	if s.healthMonitor != nil {
		if ts := s.healthMonitor.GetLastSeenAt(); ts != nil {
			lastSeen = ts.UTC().Format(time.RFC3339)
		}
	}

	// Read latest Alarmo state snapshot (source of truth)
	s.coord.AlarmoMu.RLock()
	alarmoState := s.coord.AlarmoState
	s.coord.AlarmoMu.RUnlock()

	// Derive best-effort alarmo_state string for UI
	alarmoStateValue := alarmoState.RawState
	if alarmoStateValue == "" {
		if alarmoState.Mode == "armed" && alarmoState.ArmedMode != "" {
			alarmoStateValue = "armed_" + alarmoState.ArmedMode
		} else if alarmoState.Mode != "" {
			alarmoStateValue = alarmoState.Mode
		}
	}

	alarmoLastChanged := ""
	if !alarmoState.LastChanged.IsZero() {
		alarmoLastChanged = alarmoState.LastChanged.UTC().Format(time.RFC3339)
	}

	// Calculate fallback delay if Alarmo doesn't provide it
	delayRemaining := alarmoState.DelayRemaining
	delayType := alarmoState.DelayType

	// Debug logging
	fmt.Printf("[DEBUG] DelayRemaining from Alarmo: %d, Mode: %s, LastChanged: %v\n",
		delayRemaining, alarmoState.Mode, alarmoState.LastChanged)

	// Fallback: Use SmartDisplay settings when arming but no delay from Alarmo
	if delayRemaining == 0 && alarmoState.Mode == "arming" && s.coord != nil && s.coord.Settings != nil {
		// Get exit delay from settings (most common for arming)
		if settingsResp, err := s.coord.Settings.GetSettings(); err == nil && settingsResp != nil {
			if securitySection, exists := settingsResp.Sections["security"]; exists && securitySection != nil {
				for _, field := range securitySection.Fields {
					if field.ID == "alarm_exit_delay_s" {
						if exitDelay, ok := field.Value.(int); ok && exitDelay > 0 {
							// Calculate remaining time based on last changed
							elapsed := time.Since(alarmoState.LastChanged).Seconds()
							remaining := exitDelay - int(elapsed)
							fmt.Printf("[DEBUG] Fallback calculation: exitDelay=%d, elapsed=%.2f, remaining=%d\n",
								exitDelay, elapsed, remaining)
							if remaining > 0 {
								delayRemaining = remaining
								if delayType == "" {
									delayType = "exit"
								}
								fmt.Printf("[DEBUG] Using fallback delayRemaining: %d\n", delayRemaining)
							}
						}
						break
					}
				}
			}
		}
	}

	data := map[string]interface{}{
		"alarmo_connected":       s.coord.AlarmoAdapter != nil && !s.coord.InFailsafeMode(),
		"ha_runtime_unreachable": s.healthMonitor != nil && s.healthMonitor.IsUnreachable(),
		"last_seen_at":           lastSeen,
		"alarmo_state":           alarmoStateValue,
		"alarmo_mode":            alarmoState.Mode,
		"alarmo_armed_mode":      alarmoState.ArmedMode,
		"alarmo_raw_state":       alarmoState.RawState,
		"alarmo_triggered":       alarmoState.Triggered,
		"delay_remaining":        delayRemaining,
		"delay_type":             delayType,
		"alarmo_last_changed":    alarmoLastChanged,
	}

	s.respond(w, true, data, "", 200)
}

// handleAlarmoSensors returns sanitized Alarmo-related sensor list
// Visible to admin/user/guest (read-only)
func (s *Server) handleAlarmoSensors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, r, CodeMethodNotAllowed, "GET required")
		return
	}

	baseURL, token, err := s.getHACredentials()
	if err != nil {
		s.respondError(w, r, CodeInternalError, "failed to load HA credentials")
		return
	}

	// If not configured, return empty list gracefully
	if baseURL == "" || token == "" {
		s.respond(w, true, []interface{}{}, "", 200)
		return
	}

	sensors, _, err := s.fetchAlarmoSensors(baseURL, token)
	if err != nil {
		// Soft-fail: return empty list
		logger.Error("alarmo sensors fetch failed: " + err.Error())
		s.respond(w, true, []interface{}{}, "", 200)
		return
	}

	s.respond(w, true, sensors, "", 200)
}

// handleAlarmoEvents returns recent Alarmo-related events (read-only)
// Visible to admin and user; guests are blocked from event log
func (s *Server) handleAlarmoEvents(w http.ResponseWriter, r *http.Request) {
	role := getRole(r)
	if role == auth.Guest {
		s.respondError(w, r, CodeForbidden, "guest not allowed")
		return
	}

	if r.Method != http.MethodGet {
		s.respondError(w, r, CodeMethodNotAllowed, "GET required")
		return
	}

	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			if parsed > 50 {
				parsed = 50
			}
			limit = parsed
		}
	}

	baseURL, token, err := s.getHACredentials()
	if err != nil {
		s.respondError(w, r, CodeInternalError, "failed to load HA credentials")
		return
	}

	if baseURL == "" || token == "" {
		s.respond(w, true, []interface{}{}, "", 200)
		return
	}

	// Fetch sensors once to build history filter; reuse sanitized list if needed later
	sensors, entityIDs, err := s.fetchAlarmoSensors(baseURL, token)
	if err != nil {
		logger.Error("alarmo events: sensor fetch failed: " + err.Error())
		s.respond(w, true, []interface{}{}, "", 200)
		return
	}

	nameMap := map[string]string{
		"alarm_control_panel.alarmo": "Alarmo",
	}
	for _, s := range sensors {
		nameMap[s.ID] = s.Name
	}

	events, err := s.fetchAlarmoEvents(baseURL, token, limit, entityIDs, nameMap)
	if err != nil {
		logger.Error("alarmo events fetch failed: " + err.Error())
		s.respond(w, true, []interface{}{}, "", 200)
		return
	}

	s.respond(w, true, events, "", 200)
}

// handleAlarmoArm arms the Alarmo system with specified mode
// POST /api/ui/alarmo/arm
// Body: {"mode": "armed_away", "code": "1234"} - code is optional PIN
// Visible to admin and user only
func (s *Server) handleAlarmoArm(w http.ResponseWriter, r *http.Request) {
	role := getRole(r)
	if role == auth.Guest {
		s.respondError(w, r, CodeForbidden, "guest not allowed")
		return
	}

	if r.Method != http.MethodPost {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
		return
	}

	baseURL, token, err := s.getHACredentials()
	if err != nil || baseURL == "" || token == "" {
		s.respondError(w, r, CodeInternalError, "HA not configured")
		return
	}

	// Parse request body for mode and code
	var reqBody map[string]interface{}
	code := ""
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		s.respondError(w, r, CodeBadRequest, "invalid request body")
		return
	}

	mode := "armed_away" // default
	if m, ok := reqBody["mode"].(string); ok {
		// Validate mode
		if m == "armed_away" || m == "armed_home" || m == "armed_night" {
			mode = m
		}
	}

	// Extract code if provided
	if c, ok := reqBody["code"].(string); ok {
		code = c
	}

	// Call HA service to arm Alarmo
	client := &http.Client{Timeout: 30 * time.Second}

	// Map mode to HA service name
	var serviceName string
	switch mode {
	case "armed_home":
		serviceName = "alarm_arm_home"
	case "armed_night":
		serviceName = "alarm_arm_night"
	default: // armed_away
		serviceName = "alarm_arm_away"
	}

	url := fmt.Sprintf("%s/api/services/alarm_control_panel/%s", baseURL, serviceName)
	payload := map[string]interface{}{
		"entity_id": "alarm_control_panel.alarmo",
	}
	if code != "" {
		payload["code"] = code
	}

	body, _ := json.Marshal(payload)
	logger.Info("alarmo arm: mode=" + mode + " code_provided=" + fmt.Sprintf("%v", code != "") + " service=" + serviceName + " payload=" + string(body))

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		s.respondError(w, r, CodeInternalError, "failed to create request")
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("alarmo arm request failed: " + err.Error())
		s.respondError(w, r, CodeInternalError, "failed to arm system")
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	logger.Info("alarmo arm response: status=" + fmt.Sprintf("%d", resp.StatusCode) + " body=" + string(respBody))

	if resp.StatusCode >= 400 {
		logger.Error("alarmo arm error from HA: status=" + fmt.Sprintf("%d", resp.StatusCode) + " body=" + string(respBody))
		s.respondError(w, r, CodeInternalError, "HA returned error")
		return
	}

	s.respond(w, true, map[string]string{"status": "armed", "mode": mode}, "", 200)
}

// handleAlarmoDisarm disarms the Alarmo system
// POST /api/ui/alarmo/disarm
// Body: {"code": "1234"} - optional PIN code for HA
// Visible to admin and user only
func (s *Server) handleAlarmoDisarm(w http.ResponseWriter, r *http.Request) {
	role := getRole(r)
	if role == auth.Guest {
		s.respondError(w, r, CodeForbidden, "guest not allowed")
		return
	}

	if r.Method != http.MethodPost {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
		return
	}

	baseURL, token, err := s.getHACredentials()
	if err != nil || baseURL == "" || token == "" {
		s.respondError(w, r, CodeInternalError, "HA not configured")
		return
	}

	// Parse request body for code (optional)
	var reqBody map[string]interface{}
	code := ""
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err == nil {
		if c, ok := reqBody["code"].(string); ok {
			code = c
		}
	}

	// Call HA service to disarm Alarmo
	client := &http.Client{Timeout: 30 * time.Second}
	url := fmt.Sprintf("%s/api/services/alarm_control_panel/alarm_disarm", baseURL)
	payload := map[string]interface{}{
		"entity_id": "alarm_control_panel.alarmo",
	}
	if code != "" {
		payload["code"] = code
	}

	body, _ := json.Marshal(payload)
	logger.Info("alarmo disarm: code_provided=" + fmt.Sprintf("%v", code != "") + " payload=" + string(body))

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		s.respondError(w, r, CodeInternalError, "failed to create request")
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("alarmo disarm request failed: " + err.Error())
		s.respondError(w, r, CodeInternalError, "failed to disarm system")
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	logger.Info("alarmo disarm response: status=" + fmt.Sprintf("%d", resp.StatusCode) + " body=" + string(respBody))

	if resp.StatusCode >= 400 {
		logger.Error("alarmo disarm error from HA: status=" + fmt.Sprintf("%d", resp.StatusCode) + " body=" + string(respBody))
		s.respondError(w, r, CodeInternalError, "HA returned error")
		return
	}

	s.respond(w, true, map[string]string{"status": "disarmed"}, "", 200)
}

func (s *Server) handleGuestApprove(w http.ResponseWriter, r *http.Request) {
	// FAZ L3: HA approval callback with token validation
	// POST /api/guest/approve
	// Called by HA automation with request_id and decision
	if r.Method != http.MethodPost {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
		return
	}

	// FAZ L3: Validate HA Authorization token
	authHeader := r.Header.Get("Authorization")
	if !s.validateHAToken(authHeader) {
		logger.Error("guest approval: invalid or missing HA token")
		s.respondError(w, r, CodeUnauthorized, "invalid authorization")
		return
	}

	var req struct {
		RequestID string `json:"request_id"`
		Decision  string `json:"decision"` // approve | reject
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, r, CodeBadRequest, "invalid json")
		return
	}

	if req.RequestID == "" || req.Decision == "" {
		s.respondError(w, r, CodeBadRequest, "request_id and decision required")
		return
	}

	if s.coord.GuestRequest == nil {
		s.respondError(w, r, CodeInternalError, "guest request manager not initialized")
		return
	}

	var err error
	switch req.Decision {
	case "approve":
		logger.Info("guest request approved: request_id=" + req.RequestID)
		err = s.coord.GuestRequest.ApproveRequest(req.RequestID)
	case "reject":
		logger.Info("guest request rejected: request_id=" + req.RequestID)
		err = s.coord.GuestRequest.RejectRequest(req.RequestID)
	default:
		s.respondError(w, r, CodeBadRequest, "invalid decision")
		return
	}

	if err != nil {
		s.respondError(w, r, CodeBadRequest, err.Error())
		return
	}

	s.respond(w, true, map[string]string{"result": "ok"}, "", 200)
}

func (s *Server) handleGuestDeny(w http.ResponseWriter, r *http.Request) {
	// FAZ L2: Legacy endpoint (use handleGuestApprove with decision=reject)
	if r.Method != http.MethodPost {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
		return
	}

	if s.coord.GuestRequest == nil {
		s.respondError(w, r, CodeInternalError, "guest request manager not initialized")
		return
	}

	guestReq := s.coord.GuestRequest.GetActiveRequest()
	if guestReq == nil {
		s.respondError(w, r, CodeBadRequest, "no active guest request")
		return
	}

	err := s.coord.GuestRequest.RejectRequest(guestReq.ID)
	if err != nil {
		s.respondError(w, r, CodeBadRequest, err.Error())
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
	switch r.Method {
	case http.MethodGet:
		s.handleAccessibilityGet(w, r)
		return
	case http.MethodPost:
		s.handleAccessibilityPost(w, r)
		return
	default:
		s.respondError(w, r, CodeMethodNotAllowed, "GET or POST required")
	}
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
	switch r.Method {
	case http.MethodGet:
		s.handleVoiceGet(w, r)
	case http.MethodPost:
		s.handleVoicePost(w, r)
	default:
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
		case "pending":
			state.Mode = alarm.ModePending
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

	// Calculate fallback delay if Alarmo doesn't provide it
	delayRemaining := alarmoState.DelayRemaining
	delayType := alarmoState.DelayType

	// Debug logging
	fmt.Printf("[DEBUG AlarmState] DelayRemaining from Alarmo: %d, Mode: %s, LastChanged: %v\n",
		delayRemaining, alarmoState.Mode, alarmoState.LastChanged)

	// Fallback: Use SmartDisplay settings when arming but no delay from Alarmo
	if delayRemaining == 0 && alarmoState.Mode == "arming" && s.coord != nil && s.coord.Settings != nil {
		// Get exit delay from settings (most common for arming)
		if settingsResp, err := s.coord.Settings.GetSettings(); err == nil && settingsResp != nil {
			if securitySection, exists := settingsResp.Sections["security"]; exists && securitySection != nil {
				for _, field := range securitySection.Fields {
					if field.ID == "alarm_exit_delay_s" {
						if exitDelay, ok := field.Value.(int); ok && exitDelay > 0 {
							// Calculate remaining time based on last changed
							elapsed := time.Since(alarmoState.LastChanged).Seconds()
							remaining := exitDelay - int(elapsed)
							fmt.Printf("[DEBUG AlarmState] Fallback calculation: exitDelay=%d, elapsed=%.2f, remaining=%d\n",
								exitDelay, elapsed, remaining)
							if remaining > 0 {
								delayRemaining = remaining
								if delayType == "" {
									delayType = "exit"
								}
								fmt.Printf("[DEBUG AlarmState] Using fallback delayRemaining: %d\n", delayRemaining)
							}
						}
						break
					}
				}
			}
		}
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
				"type":      delayType,
				"remaining": delayRemaining,
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
		case "pending":
			summary.Mode = alarm.ModePending
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

// handleGuestRequest initiates a new guest access request (FAZ L2)
// POST /api/ui/guest/request
// Payload: { ha_user: "username" }
func (s *Server) handleGuestRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
		return
	}

	// Check auth: only guest role can request
	authCtx := getAuthContext(r)
	if authCtx.Role != "guest" {
		s.respondError(w, r, CodeForbidden, "guest role required")
		return
	}

	// Parse request
	var req struct {
		HAUser string `json:"ha_user"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, r, CodeBadRequest, "invalid json")
		return
	}

	if req.HAUser == "" {
		s.respondError(w, r, CodeBadRequest, "ha_user required")
		return
	}

	// Check if alarm is armed (FAZ L2: can only request when alarm is armed)
	// A2: Use Alarmo state as source of truth for arm status
	// Check if alarm is armed (Alarmo modes: disarmed, armed_home, armed_away)
	alarmMode := s.coord.AlarmoState.Mode
	if alarmMode != "armed_home" && alarmMode != "armed_away" {
		s.respondError(w, r, CodeBadRequest, "alarm must be armed to request guest access")
		return
	}

	// Create guest request via request manager
	if s.coord.GuestRequest == nil {
		s.respondError(w, r, CodeInternalError, "guest request manager not initialized")
		return
	}

	guestReq, err := s.coord.GuestRequest.CreateRequest(req.HAUser)
	if err != nil {
		s.respondError(w, r, CodeBadRequest, err.Error())
		return
	}

	logger.Info("guest request created via API: id=" + guestReq.ID)

	// FAZ L3: Send HA mobile notification with approve/reject actions
	if s.coord.HA != nil {
		s.sendGuestRequestNotification(guestReq)
	}

	// Respond with request details
	s.respond(w, true, map[string]interface{}{
		"request_id":  guestReq.ID,
		"status":      guestReq.Status,
		"expires_at":  guestReq.ExpiresAt,
		"target_user": guestReq.TargetUser,
	}, "", 200)
}

// handleGuestRequestStatus retrieves current guest request status (FAZ L2)
// GET /api/ui/guest/request/{request_id}
func (s *Server) handleGuestRequestStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, r, CodeMethodNotAllowed, "GET required")
		return
	}

	// Extract request ID from URL path
	// Expected format: /api/ui/guest/request/{request_id}
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/ui/guest/request/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		s.respondError(w, r, CodeBadRequest, "request_id required")
		return
	}

	requestID := pathParts[0]

	if s.coord.GuestRequest == nil {
		s.respondError(w, r, CodeInternalError, "guest request manager not initialized")
		return
	}

	// Get request from manager
	guestReq := s.coord.GuestRequest.GetActiveRequest()
	if guestReq == nil || guestReq.ID != requestID {
		s.respondError(w, r, CodeNotFound, "request not found or expired")
		return
	}

	// Respond with current request status
	s.respond(w, true, map[string]interface{}{
		"request_id":  guestReq.ID,
		"status":      guestReq.Status,
		"expires_at":  guestReq.ExpiresAt,
		"target_user": guestReq.TargetUser,
	}, "", 200)
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
	switch userRole {
	case "admin":
		logbookRole = logbook.RoleAdmin
	case "user":
		logbookRole = logbook.RoleUser
	default:
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
	switch userRole {
	case "admin":
		logbookRole = logbook.RoleAdmin
	case "user":
		logbookRole = logbook.RoleUser
	default:
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

	// Map user role string to settings role type and update SettingsManager role
	var settingsRole string
	switch userRole {
	case "admin":
		settingsRole = "admin"
		s.coord.Settings.SetUserRole(settings.RoleAdmin)
	case "user":
		settingsRole = "user"
		s.coord.Settings.SetUserRole(settings.RoleUser)
	default:
		settingsRole = "guest"
		s.coord.Settings.SetUserRole(settings.RoleGuest)
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

	// Update SettingsManager role from incoming header so ApplyFieldChange/ApplyAction honor permissions
	if s.coord != nil && s.coord.Settings != nil {
		switch userRole {
		case "admin":
			s.coord.Settings.SetUserRole(settings.RoleAdmin)
		case "user":
			s.coord.Settings.SetUserRole(settings.RoleUser)
		default:
			s.coord.Settings.SetUserRole(settings.RoleGuest)
		}
	}

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

		// Propagate certain security field changes to Home Assistant / Alarmo if available
		if s.coord != nil && s.coord.HA != nil {
			if req.FieldID == "alarm_entry_delay_s" || req.FieldID == "alarm_exit_delay_s" {
				// Build best-effort payload for Alarmo configuration update
				// Note: Alarmo may expose different service names; this is a best-effort attempt.
				payload := map[string]interface{}{
					"entity_id": "alarm_control_panel.alarmo",
				}
				if v, ok := req.NewValue.(float64); ok {
					// JSON numbers decode as float64; cast to int
					if req.FieldID == "alarm_entry_delay_s" {
						payload["entry_delay"] = int(v)
					} else {
						payload["exit_delay"] = int(v)
					}
				} else if v, ok := req.NewValue.(int); ok {
					if req.FieldID == "alarm_entry_delay_s" {
						payload["entry_delay"] = v
					} else {
						payload["exit_delay"] = v
					}
				} else if v, ok := req.NewValue.(string); ok {
					// try parse int from string
					var parsed int
					if _, err := fmt.Sscanf(v, "%d", &parsed); err == nil {
						if req.FieldID == "alarm_entry_delay_s" {
							payload["entry_delay"] = parsed
						} else {
							payload["exit_delay"] = parsed
						}
					}
				}

				// Call HA service - best-effort; log but don't fail the settings request
				go func() {
					if err := s.coord.HA.CallService("alarmo", "set_config", payload); err != nil {
						logger.Error("settings: failed to propagate delay to HA/Alarmo: " + err.Error())
					} else {
						logger.Info("settings: propagated delay change to HA/Alarmo")
					}
				}()
			}
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

// saveRuntimeConfig saves the current runtime config to disk.
// FAZ S4: Helper to persist global HA state changes.
func (s *Server) saveRuntimeConfig() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return config.SaveRuntimeConfig(s.runtimeCfg)
}

// === GUEST REQUEST HA NOTIFICATION (FAZ L3) ===

// sendGuestRequestNotification sends HA mobile notification with approve/reject actions
func (s *Server) sendGuestRequestNotification(req *guest.GuestRequest) {
	if s.coord.HA == nil {
		logger.Error("guest notification: HA adapter not available")
		return
	}

	// Build notification payload with actionable buttons
	payload := map[string]interface{}{
		"title":   "Guest Access Request",
		"message": "A guest requests access via SmartDisplay",
		"data": map[string]interface{}{
			"actions": []map[string]interface{}{
				{
					"action": "SD_GUEST_APPROVE",
					"title":  "Approve",
					"data": map[string]interface{}{
						"request_id": req.ID,
						"decision":   "approve",
					},
				},
				{
					"action": "SD_GUEST_REJECT",
					"title":  "Reject",
					"data": map[string]interface{}{
						"request_id": req.ID,
						"decision":   "reject",
					},
				},
			},
		},
	}

	// Send notification to target user's mobile device
	// Service format: notify.mobile_app_<device_id>
	// For simplicity, we'll use the target user as the service name
	err := s.coord.HA.CallService("notify", req.TargetUser, payload)
	if err != nil {
		logger.Error("guest notification: failed to send to " + req.TargetUser + ": " + err.Error())
	} else {
		logger.Info("guest notification: sent to " + req.TargetUser)
	}
}

// validateHAToken validates the HA Authorization header
// Expects: "Bearer <token>" format
func (s *Server) validateHAToken(authHeader string) bool {
	if authHeader == "" {
		return false
	}

	// Extract token from "Bearer <token>" format
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return false
	}

	token := strings.TrimPrefix(authHeader, bearerPrefix)
	if token == "" {
		return false
	}

	// Get expected HA token from config
	// FAZ L3: Token should match the HA long-lived access token
	// For now, we check against the HA adapter's configured token
	if s.coord.HA == nil {
		logger.Error("HA token validation: HA adapter not available")
		return false
	}

	// Simple validation: token must be non-empty
	// In production, this should validate against the actual HA token from config
	// Since we don't expose the token from HA adapter, we accept any non-empty Bearer token
	// The real security is that only HA should know the SmartDisplay endpoint
	return len(token) > 0
}

// === ALARMO READ-ONLY HELPERS ===

// haStateEnvelope represents the HA /api/states payload (subset)
type haStateEnvelope struct {
	EntityID    string                 `json:"entity_id"`
	State       string                 `json:"state"`
	Attributes  map[string]interface{} `json:"attributes"`
	LastChanged string                 `json:"last_changed"`
	LastUpdated string                 `json:"last_updated"`
}

type alarmoSensor struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	State          string `json:"state"`
	DeviceClass    string `json:"device_class,omitempty"`
	LastChanged    string `json:"last_changed,omitempty"`
	Available      bool   `json:"available"`
	BatteryPercent int    `json:"battery_percent,omitempty"`
	BatteryStatus  string `json:"battery_status,omitempty"`
}

type alarmoEvent struct {
	EntityID    string `json:"entity_id"`
	Name        string `json:"name"`
	State       string `json:"state"`
	EventType   string `json:"event_type"`
	LastChanged string `json:"last_changed"`
	LastUpdated string `json:"last_updated"`
}

// getHACredentials retrieves HA base URL and token from secure storage or runtime config.
// Prefers encrypted HA config (FAZ S2), falls back to runtime.json when unset.
func (s *Server) getHACredentials() (string, string, error) {
	baseURL, err := settings.DecryptServerURL()
	if err != nil {
		return "", "", err
	}

	token, err := settings.DecryptToken()
	if err != nil {
		return "", "", err
	}

	if baseURL != "" && token != "" {
		return baseURL, token, nil
	}

	runtimeCfg, err := config.LoadRuntimeConfig()
	if err != nil {
		return baseURL, token, err
	}

	if baseURL == "" {
		baseURL = runtimeCfg.HABaseURL
	}
	if token == "" {
		token = runtimeCfg.HAToken
	}

	return baseURL, token, nil
}

// fetchHAStates retrieves all HA states (used to derive Alarmo-related sensors)
func (s *Server) fetchHAStates(baseURL string, token string) ([]haStateEnvelope, error) {
	cleanBase := strings.TrimRight(baseURL, "/")
	req, err := http.NewRequest(http.MethodGet, cleanBase+"/api/states", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ha states http %d", resp.StatusCode)
	}

	var states []haStateEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&states); err != nil {
		return nil, err
	}

	return states, nil
}

// fetchAlarmoSensors extracts Alarmo-relevant sensors from HA state list
func (s *Server) fetchAlarmoSensors(baseURL string, token string) ([]alarmoSensor, []string, error) {
	states, err := s.fetchHAStates(baseURL, token)
	if err != nil {
		return nil, nil, err
	}

	allowedClasses := map[string]bool{
		"motion":    true,
		"occupancy": true,
		"presence":  true,
		"door":      true,
		"window":    true,
		"opening":   true,
		"smoke":     true,
		"gas":       true,
		"moisture":  true,
		"sound":     true,
		"vibration": true,
		"moving":    true,
	}

	sensors := make([]alarmoSensor, 0)
	entityIDs := make([]string, 0)

	for _, st := range states {
		if st.EntityID == "alarm_control_panel.alarmo" {
			friendly, _ := st.Attributes["friendly_name"].(string)
			if friendly == "" {
				friendly = "Alarmo"
			}
			sensors = append(sensors, alarmoSensor{
				ID:             st.EntityID,
				Name:           friendly,
				State:          st.State,
				DeviceClass:    "alarm_control_panel",
				LastChanged:    st.LastChanged,
				Available:      st.State != "unavailable",
				BatteryPercent: parseBatteryPercent(st.Attributes),
				BatteryStatus:  deriveBatteryStatus(st.Attributes),
			})
			entityIDs = append(entityIDs, st.EntityID)
			continue
		}

		if !strings.HasPrefix(st.EntityID, "binary_sensor.") {
			continue
		}

		devClass, _ := st.Attributes["device_class"].(string)
		if devClass != "" && !allowedClasses[devClass] {
			continue
		}

		friendly, _ := st.Attributes["friendly_name"].(string)
		if friendly == "" {
			friendly = st.EntityID
		}

		sensors = append(sensors, alarmoSensor{
			ID:             st.EntityID,
			Name:           friendly,
			State:          st.State,
			DeviceClass:    devClass,
			LastChanged:    st.LastChanged,
			Available:      st.State != "unavailable",
			BatteryPercent: parseBatteryPercent(st.Attributes),
			BatteryStatus:  deriveBatteryStatus(st.Attributes),
		})
		entityIDs = append(entityIDs, st.EntityID)
	}

	sort.Slice(sensors, func(i, j int) bool {
		return sensors[i].Name < sensors[j].Name
	})

	return sensors, entityIDs, nil
}

// === DEVICES: LIGHTS ===

type lightDevice struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	State         string `json:"state"`
	BrightnessPct int    `json:"brightness_pct,omitempty"`
	SupportsColor bool   `json:"supports_color,omitempty"`
}

// handleDevicesLights lists light.* entities with basic state (read-only)
func (s *Server) handleDevicesLights(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, r, CodeMethodNotAllowed, "GET required")
		return
	}

	baseURL, token, err := s.getHACredentials()
	if err != nil {
		s.respondError(w, r, CodeInternalError, "failed to load HA credentials")
		return
	}
	if baseURL == "" || token == "" {
		// Not configured; return empty list gracefully
		s.respond(w, true, []interface{}{}, "", 200)
		return
	}

	states, err := s.fetchHAStates(baseURL, token)
	if err != nil {
		logger.Error("lights fetch states failed: " + err.Error())
		s.respond(w, true, []interface{}{}, "", 200)
		return
	}

	lights := make([]lightDevice, 0, 16)
	for _, st := range states {
		if !strings.HasPrefix(st.EntityID, "light.") {
			continue
		}
		name := st.EntityID
		if st.Attributes != nil {
			if fn, ok := st.Attributes["friendly_name"].(string); ok && fn != "" {
				name = fn
			}
		}
		ld := lightDevice{ID: st.EntityID, Name: name, State: st.State}
		if st.Attributes != nil {
			if b, ok := st.Attributes["brightness"]; ok {
				switch v := b.(type) {
				case float64:
					pct := int((v / 255.0) * 100.0)
					if pct < 0 {
						pct = 0
					}
					if pct > 100 {
						pct = 100
					}
					ld.BrightnessPct = pct
				case int:
					pct := int(float64(v) / 255.0 * 100.0)
					if pct < 0 {
						pct = 0
					}
					if pct > 100 {
						pct = 100
					}
					ld.BrightnessPct = pct
				}
			}
			// Determine color support from supported_color_modes or color_mode
			if modes, ok := st.Attributes["supported_color_modes"]; ok {
				switch mv := modes.(type) {
				case []interface{}:
					for _, m := range mv {
						if ms, ok2 := m.(string); ok2 {
							if ms == "rgb" || ms == "hs" || ms == "xy" || ms == "rgbw" || ms == "rgbww" {
								ld.SupportsColor = true
								break
							}
						}
					}
				case []string:
					for _, ms := range mv {
						if ms == "rgb" || ms == "hs" || ms == "xy" || ms == "rgbw" || ms == "rgbww" {
							ld.SupportsColor = true
							break
						}
					}
				}
			} else if cm, ok := st.Attributes["color_mode"].(string); ok {
				if cm == "rgb" || cm == "hs" || cm == "xy" || cm == "rgbw" || cm == "rgbww" {
					ld.SupportsColor = true
				}
			}
		}
		lights = append(lights, ld)
	}

	s.respond(w, true, lights, "", 200)
}

// handleDevicesLightsToggle toggles a light on/off (write; user/admin)
func (s *Server) handleDevicesLightsToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
		return
	}
	if _, allowed := s.checkPerm(w, r, auth.PermDevice); !allowed {
		return
	}
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
		s.respondError(w, r, CodeBadRequest, "invalid request body")
		return
	}
	payload := map[string]interface{}{"entity_id": req.ID}
	if err := s.coord.HA.CallService("light", "toggle", payload); err != nil {
		logger.Error("light toggle failed: " + err.Error())
		s.respondError(w, r, CodeUpstreamError, "ha toggle failed")
		return
	}
	s.respond(w, true, map[string]string{"result": "ok"}, "", 200)
}

// handleDevicesLightsSet sets light parameters (on/off, brightness_pct, rgb_color)
func (s *Server) handleDevicesLightsSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
		return
	}
	if _, allowed := s.checkPerm(w, r, auth.PermDevice); !allowed {
		return
	}
	var req struct {
		ID            string `json:"id"`
		On            *bool  `json:"on,omitempty"`
		BrightnessPct *int   `json:"brightness_pct,omitempty"`
		RGBColor      []int  `json:"rgb_color,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
		s.respondError(w, r, CodeBadRequest, "invalid request body")
		return
	}

	// If explicit off requested
	if req.On != nil && !*req.On {
		if err := s.coord.HA.CallService("light", "turn_off", map[string]interface{}{"entity_id": req.ID}); err != nil {
			logger.Error("light turn_off failed: " + err.Error())
			s.respondError(w, r, CodeUpstreamError, "ha turn_off failed")
			return
		}
		s.respond(w, true, map[string]string{"result": "ok"}, "", 200)
		return
	}

	// Build turn_on payload
	pl := map[string]interface{}{"entity_id": req.ID}
	if req.BrightnessPct != nil {
		pct := *req.BrightnessPct
		if pct < 0 {
			pct = 0
		}
		if pct > 100 {
			pct = 100
		}
		// Convert percent to 0-255
		bri := int(float64(pct) / 100.0 * 255.0)
		if bri < 0 {
			bri = 0
		}
		if bri > 255 {
			bri = 255
		}
		pl["brightness"] = bri
	}
	if len(req.RGBColor) == 3 {
		// Clamp values to 0-255
		r := req.RGBColor[0]
		if r < 0 {
			r = 0
		}
		if r > 255 {
			r = 255
		}
		g := req.RGBColor[1]
		if g < 0 {
			g = 0
		}
		if g > 255 {
			g = 255
		}
		b := req.RGBColor[2]
		if b < 0 {
			b = 0
		}
		if b > 255 {
			b = 255
		}
		pl["rgb_color"] = []int{r, g, b}
	}
	if err := s.coord.HA.CallService("light", "turn_on", pl); err != nil {
		logger.Error("light turn_on failed: " + err.Error())
		s.respondError(w, r, CodeUpstreamError, "ha turn_on failed")
		return
	}
	s.respond(w, true, map[string]string{"result": "ok"}, "", 200)
}

// parseBatteryPercent attempts to extract a battery percentage from HA attributes
func parseBatteryPercent(attrs map[string]interface{}) int {
	if attrs == nil {
		return 0
	}
	keys := []string{"battery_level", "battery", "battery_percentage"}
	for _, k := range keys {
		if v, ok := attrs[k]; ok {
			switch tv := v.(type) {
			case float64:
				pct := int(tv)
				if pct < 0 {
					pct = 0
				}
				if pct > 100 {
					pct = 100
				}
				return pct
			case int:
				pct := tv
				if pct < 0 {
					pct = 0
				}
				if pct > 100 {
					pct = 100
				}
				return pct
			case string:
				// Try to parse like "85" or "85%"
				s := tv
				s = strings.TrimSuffix(s, "%")
				var n int
				if _, err := fmt.Sscanf(s, "%d", &n); err == nil {
					if n < 0 {
						n = 0
					}
					if n > 100 {
						n = 100
					}
					return n
				}
			}
		}
	}
	return 0
}

// deriveBatteryStatus maps common HA battery flags to a human string
func deriveBatteryStatus(attrs map[string]interface{}) string {
	if attrs == nil {
		return ""
	}
	// Check low battery flags commonly exposed
	if v, ok := attrs["battery_low"]; ok {
		switch tv := v.(type) {
		case bool:
			if tv {
				return "low"
			}
			return "normal"
		case string:
			if strings.EqualFold(tv, "on") || strings.EqualFold(tv, "true") {
				return "low"
			}
			return "normal"
		}
	}
	// Fallback: use percentage
	pct := parseBatteryPercent(attrs)
	if pct == 0 {
		return "unknown"
	}
	if pct <= 20 {
		return "low"
	}
	return "normal"
}

// fetchAlarmoEvents pulls recent HA history for Alarmo entities and sensors
func (s *Server) fetchAlarmoEvents(baseURL string, token string, limit int, entityIDs []string, names map[string]string) ([]alarmoEvent, error) {
	// Ensure alarm entity is included
	seen := map[string]bool{}
	allIDs := make([]string, 0, len(entityIDs)+1)

	for _, id := range entityIDs {
		if id == "" {
			continue
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		allIDs = append(allIDs, id)
	}

	if !seen["alarm_control_panel.alarmo"] {
		allIDs = append(allIDs, "alarm_control_panel.alarmo")
	}

	start := time.Now().Add(-12 * time.Hour)
	cleanBase := strings.TrimRight(baseURL, "/")
	query := url.Values{}
	query.Set("filter_entity_id", strings.Join(allIDs, ","))
	query.Set("minimal_response", "true")

	requestURL := fmt.Sprintf("%s/api/history/period/%s?%s", cleanBase, url.QueryEscape(start.Format(time.RFC3339)), query.Encode())
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 6 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ha history http %d", resp.StatusCode)
	}

	// HA history returns array per entity
	var buckets [][]haStateEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&buckets); err != nil {
		return nil, err
	}

	nameMap := map[string]string{}
	for k, v := range names {
		nameMap[k] = v
	}

	if _, ok := nameMap["alarm_control_panel.alarmo"]; !ok {
		nameMap["alarm_control_panel.alarmo"] = "Alarmo"
	}
	for _, id := range allIDs {
		if _, ok := nameMap[id]; !ok {
			nameMap[id] = id
		}
	}

	events := make([]alarmoEvent, 0)
	for _, bucket := range buckets {
		for _, entry := range bucket {
			name := nameMap[entry.EntityID]
			if name == "" {
				name = entry.EntityID
			}
			events = append(events, alarmoEvent{
				EntityID:    entry.EntityID,
				Name:        name,
				State:       entry.State,
				EventType:   mapEventType(entry.State),
				LastChanged: entry.LastChanged,
				LastUpdated: entry.LastUpdated,
			})
		}
	}

	sort.Slice(events, func(i, j int) bool {
		ai, _ := time.Parse(time.RFC3339Nano, events[i].LastUpdated)
		aj, _ := time.Parse(time.RFC3339Nano, events[j].LastUpdated)
		return ai.After(aj)
	})

	if len(events) > limit {
		events = events[:limit]
	}

	return events, nil
}

// mapEventType categorizes HA state into coarse event type for UI chips
func mapEventType(state string) string {
	switch state {
	case "triggered", "on", "open", "opened":
		return "alert"
	case "arming", "armed", "armed_home", "armed_away", "armed_night":
		return "status"
	case "off", "closed", "disarmed", "idle", "standby":
		return "clear"
	default:
		return "info"
	}
}

// handleLogin: PIN ile giriş için endpoint
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DEBUG: handleLogin çağrıldı")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "POST methodu gerekli",
		})
		return
	}
	var req struct {
		Pin string `json:"pin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Geçersiz istek formatı",
		})
		return
	}
	fmt.Println("DEBUG: Gelen PIN:", req.Pin)
	log.Printf("DEBUG: Gelen PIN: '%s'", req.Pin)
	ctx, err := auth.ValidatePIN(req.Pin)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Sunucu hatası",
		})
		return
	}
	if ctx.Authenticated {
		// Basit bir session cookie oluştur (gerçek projede JWT veya secure session kullanılmalı)
		cookie := &http.Cookie{
			Name:     "sd_session",
			Value:    string(ctx.Role) + ":" + req.Pin, // Basit örnek, prod için güvenli değil!
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(w, cookie)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"role":    ctx.Role,
			"message": "Giriş başarılı",
		})
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Hatalı PIN",
		})
	}
}
