package api

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"smartdisplay-core/internal/audit"
	"smartdisplay-core/internal/auth"
	"smartdisplay-core/internal/contexthelp"
	"smartdisplay-core/internal/health"
	"smartdisplay-core/internal/logger"
	"smartdisplay-core/internal/morning"
	"smartdisplay-core/internal/system"
	"smartdisplay-core/internal/scorecard"
	"smartdisplay-core/internal/telemetry"
	"smartdisplay-core/internal/update"
	"syscall"
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
		s.respond(w, false, nil, "permission denied", 403)
	}
	return role, allowed
}

type Server struct {
	coord       *system.Coordinator
	httpServer  *http.Server
	mu          sync.Mutex
	telemetry   *telemetry.Collector
	updateMgr   *update.Manager
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

	return &Server{coord: coord, telemetry: tel, updateMgr: updateMgr}
}

func (s *Server) Start(port int) error {
				// Admin smoke test endpoint (admin-only)
				mux.HandleFunc("/api/admin/smoke", s.handleAdminSmoke)
			// handleAdminSmoke runs self-check and returns summary (admin-only)
			func (s *Server) handleAdminSmoke(w http.ResponseWriter, r *http.Request) {
			   role := getRole(r)
			   if role != auth.Admin {
				   s.respond(w, false, nil, "admin only", 403)
				   return
			   }
			   if r.Method != "POST" {
				   s.respond(w, false, nil, "invalid method", 400)
				   return
			   }
			   res := s.coord.SelfCheck()
			   summary := map[string]interface{}{
				   "ha_connected": res.HAConnected,
				   "alarm_valid":  res.AlarmValid,
				   "ai_running":   res.AIRunning,
				   "hardware":     res.Hardware,
				   "details":      res.Details,
			   }
			   audit.Record("smoke_test", "admin ran smoke test")
			   s.respond(w, true, summary, "", 200)
			}
			// Admin restart endpoint (admin-only)
			mux.HandleFunc("/api/admin/restart", s.handleAdminRestart)
		// handleAdminRestart triggers graceful shutdown and exits with code 42 (admin-only)
		func (s *Server) handleAdminRestart(w http.ResponseWriter, r *http.Request) {
		   role := getRole(r)
		   if role != auth.Admin {
			   s.respond(w, false, nil, "admin only", 403)
			   return
		   }
		   go func() {
			   // Give response before shutting down
			   time.Sleep(200 * time.Millisecond)
			   // Graceful shutdown: send SIGTERM to self
			   p, _ := os.FindProcess(os.Getpid())
			   p.Signal(syscall.SIGTERM)
			   // Exit with code 42 after a short delay
			   time.Sleep(500 * time.Millisecond)
			   os.Exit(42)
		   }()
		   s.respond(w, true, map[string]string{"result": "restarting"}, "", 200)
		}
		// Admin backup/restore endpoints (admin-only)
		mux.HandleFunc("/api/admin/backup", s.handleAdminBackup)
		mux.HandleFunc("/api/admin/restore", s.handleAdminRestore)
	// handleAdminBackup streams a zip of config/runtime files (admin-only)
	func (s *Server) handleAdminBackup(w http.ResponseWriter, r *http.Request) {
	   role := getRole(r)
	   if role != auth.Admin {
		   s.respond(w, false, nil, "admin only", 403)
		   return
	   }
	   w.Header().Set("Content-Type", "application/zip")
	   w.Header().Set("Content-Disposition", "attachment; filename=backup.zip")
	   zw := zip.NewWriter(w)
	   files := []string{"data/runtime.json", "configs/features.json"}
	   // Add any other small config files in data/
	   _ = filepath.Walk("data", func(path string, info os.FileInfo, err error) error {
		   if err == nil && !info.IsDir() && path != "data/runtime.json" {
			   files = append(files, path)
		   }
		   return nil
	   })
	   for _, file := range files {
		   f, err := os.Open(file)
		   if err != nil {
			   continue // skip missing
		   }
		   defer f.Close()
		   wtr, err := zw.Create(file)
		   if err != nil {
			   continue
		   }
		   io.Copy(wtr, f)
	   }
	   zw.Close()
	}

	// handleAdminRestore accepts a zip, validates, and atomically restores config/runtime files (admin-only)
	func (s *Server) handleAdminRestore(w http.ResponseWriter, r *http.Request) {
	   role := getRole(r)
	   if role != auth.Admin {
		   s.respond(w, false, nil, "admin only", 403)
		   return
	   }
	   if r.Method != "POST" {
		   s.respond(w, false, nil, "invalid method", 400)
		   return
	   }
	   // Read zip from body
	   tmpZip, err := ioutil.TempFile("", "restore-*.zip")
	   if err != nil {
		   s.respond(w, false, nil, "tempfile error", 500)
		   return
	   }
	   defer os.Remove(tmpZip.Name())
	   io.Copy(tmpZip, r.Body)
	   tmpZip.Close()
	   zr, err := zip.OpenReader(tmpZip.Name())
	   if err != nil {
		   s.respond(w, false, nil, "invalid zip", 400)
		   return
	   }
	   defer zr.Close()
	   valid := false
	   for _, f := range zr.File {
		   if f.Name == "data/runtime.json" || f.Name == "configs/features.json" {
			   valid = true
			   break
		   }
	   }
	   if !valid {
		   s.respond(w, false, nil, "missing required files", 400)
		   return
	   }
	   // Extract to temp, then swap
	   for _, f := range zr.File {
		   if !(strings.HasPrefix(f.Name, "data/") || strings.HasPrefix(f.Name, "configs/")) {
			   continue
		   }
		   rc, err := f.Open()
		   if err != nil {
			   continue
		   }
		   tmpPath := f.Name + ".tmp"
		   out, err := os.Create(tmpPath)
		   if err != nil {
			   rc.Close()
			   continue
		   }
		   io.Copy(out, rc)
		   rc.Close()
		   out.Close()
	   }
	   // Atomically swap
	   for _, f := range zr.File {
		   if !(strings.HasPrefix(f.Name, "data/") || strings.HasPrefix(f.Name, "configs/")) {
			   continue
		   }
		   tmpPath := f.Name + ".tmp"
		   if _, err := os.Stat(tmpPath); err == nil {
			   os.Rename(tmpPath, f.Name)
		   }
	   }
	   // Never log tokens
	   importLogger()
	   logger.Info("config restore completed (restart recommended)")
	   s.respond(w, true, map[string]string{"result": "ok"}, "", 200)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/ui/home", s.handleUIHome)
	mux.HandleFunc("/api/ui/alarm", s.handleUIAlarm)
	mux.HandleFunc("/api/ui/ai/hint", s.handleUIAIHint)
	mux.HandleFunc("/api/ui/ai/why", s.handleUIAIWhy)
	mux.HandleFunc("/api/ai/daily", s.handleAIDaily)
	mux.HandleFunc("/api/ai/anomalies", s.handleAIAnomalies)
	mux.HandleFunc("/api/ai/morning", s.handleAIMorning)
	// Admin telemetry endpoints (admin-only)
	mux.HandleFunc("/api/admin/telemetry/summary", s.handleTelemetrySummary)
	mux.HandleFunc("/api/admin/telemetry/optin", s.handleTelemetryOptIn)
	// Admin update endpoints (admin-only, no-op stubs in this phase)
	mux.HandleFunc("/api/admin/update/status", s.handleUpdateStatus)
	mux.HandleFunc("/api/admin/update/stage", s.handleUpdateStage)
	// Setup wizard endpoints (local only by default)
	mux.HandleFunc("/api/setup/status", s.handleSetupStatus)
	mux.HandleFunc("/api/setup/save", s.handleSetupSave)
	mux.HandleFunc("/api/setup/test_ha", s.handleSetupTestHA)
	mux.HandleFunc("/api/setup/complete", s.handleSetupComplete)
	// Health endpoint (same as /health)
	mux.HandleFunc("/api/health", health.HealthHandler)
	// isLocalhostOrAllowed returns true if the request is from localhost or SETUP_ALLOW_LAN=true
	func isLocalhostOrAllowed(r *http.Request) bool {
	   allowLAN := os.Getenv("SETUP_ALLOW_LAN") == "true"
	   host := r.RemoteAddr
	   if strings.HasPrefix(host, "[") { // IPv6
		   host = host[1:strings.Index(host, "]")]
	   } else if strings.Contains(host, ":") {
		   host = host[:strings.Index(host, ":")]
	   }
	   if host == "127.0.0.1" || host == "::1" || host == "localhost" {
		   return true
	   }
	   return allowLAN
	}

	// handleSetupStatus returns current setup status and config (no secrets)
	func (s *Server) handleSetupStatus(w http.ResponseWriter, r *http.Request) {
	   if !isLocalhostOrAllowed(r) {
		   s.respond(w, false, nil, "forbidden", 403)
		   return
	   }
	   cfg, _ := config.LoadRuntimeConfig()
	   if cfg == nil {
		   cfg = &config.RuntimeConfig{}
	   }
	   // Do not expose token
	   resp := map[string]interface{}{
		   "ha_base_url": cfg.HABaseURL,
		   "ui_enabled": cfg.UIEnabled,
		   "hardware_profile": cfg.HardwareProfile,
		   "wizard_completed": cfg.WizardCompleted,
	   }
	   s.respond(w, true, resp, "", 200)
	}

	// handleSetupSave saves new config (no logging of secrets)
	func (s *Server) handleSetupSave(w http.ResponseWriter, r *http.Request) {
	   if !isLocalhostOrAllowed(r) {
		   s.respond(w, false, nil, "forbidden", 403)
		   return
	   }
	   if r.Method != "POST" {
		   s.respond(w, false, nil, "invalid method", 400)
		   return
	   }
	   var req struct {
		   HABaseURL       string `json:"ha_base_url"`
		   HAToken         string `json:"ha_token"`
		   UIEnabled       bool   `json:"ui_enabled"`
		   HardwareProfile string `json:"hardware_profile"`
	   }
	   if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		   s.respond(w, false, nil, "invalid json", 400)
		   return
	   }
	   cfg, _ := config.LoadRuntimeConfig()
	   if cfg == nil {
		   cfg = &config.RuntimeConfig{}
	   }
	   cfg.HABaseURL = req.HABaseURL
	   cfg.HAToken = req.HAToken
	   cfg.UIEnabled = req.UIEnabled
	   cfg.HardwareProfile = req.HardwareProfile
	   // Do not log token
	   if err := config.SaveRuntimeConfig(cfg); err != nil {
		   s.respond(w, false, nil, "save error", 500)
		   return
	   }
	   s.respond(w, true, map[string]string{"result": "ok"}, "", 200)
	}

	// handleSetupTestHA tests HA connection (no token exposure)
	func (s *Server) handleSetupTestHA(w http.ResponseWriter, r *http.Request) {
	   if !isLocalhostOrAllowed(r) {
		   s.respond(w, false, nil, "forbidden", 403)
		   return
	   }
	   if r.Method != "POST" {
		   s.respond(w, false, nil, "invalid method", 400)
		   return
	   }
	   var req struct {
		   HABaseURL string `json:"ha_base_url"`
		   HAToken   string `json:"ha_token"`
	   }
	   if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		   s.respond(w, false, nil, "invalid json", 400)
		   return
	   }
	   // Minimal REST GET /api/ with token
	   client := &http.Client{Timeout: 5 * time.Second}
	   apiURL := strings.TrimRight(req.HABaseURL, "/") + "/api/"
	   httpReq, err := http.NewRequest("GET", apiURL, nil)
	   if err != nil {
		   s.respond(w, false, nil, "bad url", 400)
		   return
	   }
	   httpReq.Header.Set("Authorization", "Bearer "+req.HAToken)
	   resp, err := client.Do(httpReq)
	   if err != nil {
		   s.respond(w, false, nil, "connect fail", 200)
		   return
	   }
	   defer resp.Body.Close()
	   if resp.StatusCode == 200 {
		   s.respond(w, true, map[string]string{"result": "success"}, "", 200)
	   } else {
		   s.respond(w, false, nil, "auth fail", 200)
	   }
	}

	// handleSetupComplete marks wizard as completed
	func (s *Server) handleSetupComplete(w http.ResponseWriter, r *http.Request) {
	   if !isLocalhostOrAllowed(r) {
		   s.respond(w, false, nil, "forbidden", 403)
		   return
	   }
	   if r.Method != "POST" {
		   s.respond(w, false, nil, "invalid method", 400)
		   return
	   }
	   cfg, _ := config.LoadRuntimeConfig()
	   if cfg == nil {
		   cfg = &config.RuntimeConfig{}
	   }
	   cfg.WizardCompleted = true
	   if err := config.SaveRuntimeConfig(cfg); err != nil {
		   s.respond(w, false, nil, "save error", 500)
		   return
	   }
	   s.respond(w, true, map[string]string{"result": "ok"}, "", 200)
	}
	// handleUIAIHint returns one short AI sentence for UI card
	func (s *Server) handleUIAIHint(w http.ResponseWriter, r *http.Request) {
	   start := time.Now()
	ai := s.coord.GetCurrentInsight()
	hint := ai.Detail
	resp := map[string]interface{}{"hint": hint, "tone": ai.Tone}
	   elapsed := time.Since(start)
	   if elapsed > 100*time.Millisecond {
		   log.Printf("SLOW /api/ui/ai/hint: %v", elapsed)
	   }
	   s.respond(w, true, resp, "", 200)
	}

	// handleUIAIWhy returns explanation text only
	func (s *Server) handleUIAIWhy(w http.ResponseWriter, r *http.Request) {
	   start := time.Now()
	   why := s.coord.ExplainInsight()
	   resp := map[string]string{"why": why}
	   elapsed := time.Since(start)
	   if elapsed > 100*time.Millisecond {
		   log.Printf("SLOW /api/ui/ai/why: %v", elapsed)
	   }
	   s.respond(w, true, resp, "", 200)
	}
	// handleUIHome serves optimized payload for main screen
	func (s *Server) handleUIHome(w http.ResponseWriter, r *http.Request) {
	   start := time.Now()
	   alarmSummary := s.coord.Alarm.CurrentState()
	ai := s.coord.GetCurrentInsight()
	   var countdownRemaining int
	   var countdownActive bool
	   if s.coord.Countdown != nil {
		   // Use exported fields or add method if needed
		   countdownActive = s.coord.CountdownActive
		   countdownRemaining = s.coord.CountdownRemaining
	   }
	   haOnline := false
	   if s.coord.HA != nil {
		   haOnline = s.coord.HA.IsConnected()
	   }
	   resp := map[string]interface{}{
		  "alarm": alarmSummary,
		  "ai": map[string]interface{}{
			  "summary": ai.Detail,
			  "severity": ai.Severity,
			  "tone": ai.Tone,
		  },
		  "countdown": map[string]interface{}{
			  "active": countdownActive,
			  "remaining": countdownRemaining,
		  },
		  "ha_online": haOnline,
	   }
	   elapsed := time.Since(start)
	   if elapsed > 100*time.Millisecond {
		   log.Printf("SLOW /api/ui/home: %v", elapsed)
	   }
	   s.respond(w, true, resp, "", 200)
	}

	// handleUIAlarm serves optimized alarm screen data
	func (s *Server) handleUIAlarm(w http.ResponseWriter, r *http.Request) {
	   start := time.Now()
	   alarmState := s.coord.Alarm.CurrentState()
	   var countdownRemaining int
	   if s.coord.Countdown != nil {
		   countdownRemaining = s.coord.CountdownRemaining
	   }
	   guestStatus := ""
	   if s.coord.Guest != nil {
		   guestStatus = s.coord.Guest.CurrentState()
	   }
	   lastAlarmEvent := ""
	   if s.coord.Alarm != nil {
		   lastAlarmEvent = s.coord.Alarm.LastEvent
	   }
	   resp := map[string]interface{}{
		   "alarm_state": alarmState,
		   "countdown_remaining": countdownRemaining,
		   "guest_status": guestStatus,
		   "last_alarm_event": lastAlarmEvent,
	   }
	   elapsed := time.Since(start)
	   if elapsed > 100*time.Millisecond {
		   log.Printf("SLOW /api/ui/alarm: %v", elapsed)
	   }
	   s.respond(w, true, resp, "", 200)
	}
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/hardware/status", s.handleHardwareStatus)
	// handleHardwareStatus exposes hardware health/status
	func (s *Server) handleHardwareStatus(w http.ResponseWriter, r *http.Request) {
		status := s.coord.HardwareHealth()
		s.respond(w, true, status, "", 200)
	}
	mux.HandleFunc("/api/state/overview", s.handleOverview)
	mux.HandleFunc("/api/alarm/state", s.handleAlarmState)
	mux.HandleFunc("/api/alarm/arm", s.handleAlarmArm)
	mux.HandleFunc("/api/alarm/disarm", s.handleAlarmDisarm)
	mux.HandleFunc("/api/guest/state", s.handleGuestState)
	mux.HandleFunc("/api/guest/request", s.handleGuestRequest)
	mux.HandleFunc("/api/guest/approve", s.handleGuestApprove)
	mux.HandleFunc("/api/guest/deny", s.handleGuestDeny)
	mux.HandleFunc("/api/guest/exit", s.handleGuestExit)
	mux.HandleFunc("/api/ai/insight", s.handleAIInsight)
	mux.HandleFunc("/api/ai/insight/explain", s.handleAIExplain)
	mux.HandleFunc("/api/ai/history", s.handleAIHistory)
	mux.HandleFunc("/api/failsafe", s.handleFailsafe)
	mux.HandleFunc("/api/logbook", s.handleLogbook)
// handleLogbook returns the audit log as a timeline of human-readable entries (no raw logs)
func (s *Server) handleLogbook(w http.ResponseWriter, r *http.Request) {
	entries := audit.GetEntries()
	timeline := audit.ToTimeline(entries)
	s.respond(w, true, timeline, "", 200)
}

	s.httpServer = &http.Server{
		Addr:           ":" + itoa(port),
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("api server error:", err)
		}
	}()
	return nil
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
	       "active": s.coord.InFailsafeMode(),
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
       overview := map[string]interface{}{
	       "alarm": s.coord.Alarm.CurrentState(),
	       "guest": s.coord.Guest.CurrentState(),
	       "ha":    s.coord.HA.IsConnected(),
	       "ai":    s.coord.GetCurrentInsight(),
       }
       s.respond(w, true, overview, "", 200)

// Dedicated failsafe endpoint
func (s *Server) handleFailsafe(w http.ResponseWriter, r *http.Request) {
       s.coord.UpdateFailsafeState()
       state := map[string]interface{}{
	       "active": s.coord.InFailsafeMode(),
	       "explanation": s.coord.FailsafeExplanation(),
       }
       s.respond(w, true, state, "", 200)
}
}

func (s *Server) handleAlarmState(w http.ResponseWriter, r *http.Request) {
	_, allowed := s.checkPerm(w, r, auth.PermAlarm)
	if !allowed {
		return
	}
	s.respond(w, true, map[string]string{"state": s.coord.Alarm.CurrentState()}, "", 200)
}

func (s *Server) handleAlarmArm(w http.ResponseWriter, r *http.Request) {
	_, allowed := s.checkPerm(w, r, auth.PermAlarm)
	if !allowed {
		return
	}
	if r.Method != "POST" {
		s.respond(w, false, nil, "invalid method", 400)
		return
	}
	err := s.coord.Alarm.Handle("ARM_REQUEST")
	if err != nil {
		s.respond(w, false, nil, "alarm arm error", 500)
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
		s.respond(w, false, nil, "invalid method", 400)
		return
	}
	err := s.coord.Alarm.Handle("DISARM_REQUEST")
	if err != nil {
		s.respond(w, false, nil, "alarm disarm error", 500)
		return
	}
	s.respond(w, true, map[string]string{"result": "ok"}, "", 200)
}

func (s *Server) handleGuestState(w http.ResponseWriter, r *http.Request) {
	_, allowed := s.checkPerm(w, r, auth.PermGuest)
	if !allowed {
		return
	}
	s.respond(w, true, map[string]string{"state": s.coord.Guest.CurrentState()}, "", 200)
}

func (s *Server) handleGuestRequest(w http.ResponseWriter, r *http.Request) {
	_, allowed := s.checkPerm(w, r, auth.PermGuest)
	if !allowed {
		return
	}
	if r.Method != "POST" {
		s.respond(w, false, nil, "invalid method", 400)
		return
	}
	err := s.coord.Guest.Handle("REQUEST")
	if err != nil {
		s.respond(w, false, nil, "guest request error", 500)
		return
	}
	s.respond(w, true, map[string]string{"result": "ok"}, "", 200)
}

func (s *Server) handleGuestApprove(w http.ResponseWriter, r *http.Request) {
	role, allowed := s.checkPerm(w, r, auth.PermGuest)
	if !allowed || role != auth.Admin {
		if allowed {
			s.respond(w, false, nil, "permission denied", 403)
		}
		return
	}
	if r.Method != "POST" {
		s.respond(w, false, nil, "invalid method", 400)
		return
	}
	err := s.coord.Guest.Handle("APPROVE")
	if err != nil {
		s.respond(w, false, nil, "guest approve error", 500)
		return
	}
	s.respond(w, true, map[string]string{"result": "ok"}, "", 200)
}

func (s *Server) handleGuestDeny(w http.ResponseWriter, r *http.Request) {
	role, allowed := s.checkPerm(w, r, auth.PermGuest)
	if !allowed || role != auth.Admin {
		if allowed {
			s.respond(w, false, nil, "permission denied", 403)
		}
		return
	}
	if r.Method != "POST" {
		s.respond(w, false, nil, "invalid method", 400)
		return
	}
	err := s.coord.Guest.Handle("DENY")
	if err != nil {
		s.respond(w, false, nil, "guest deny error", 500)
		return
	}
	s.respond(w, true, map[string]string{"result": "ok"}, "", 200)
}

func (s *Server) handleGuestExit(w http.ResponseWriter, r *http.Request) {
	_, allowed := s.checkPerm(w, r, auth.PermGuest)
	if !allowed {
		return
	}
	if r.Method != "POST" {
		s.respond(w, false, nil, "invalid method", 400)
		return
	}
	err := s.coord.Guest.Handle("EXIT")
	if err != nil {
		s.respond(w, false, nil, "guest exit error", 500)
		return
	}
	s.respond(w, true, map[string]string{"result": "ok"}, "", 200)
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
       s.respond(w, false, nil, "invalid method", 400)
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
       s.respond(w, false, nil, "invalid method", 400)
       return
   }
   // Gather current state (stub: replace with real data sources)
   alarmStatus := s.coord.Alarm.CurrentState()
   nightEvents := []string{} // TODO: fetch from nightguard or event log
   todayContext := ""        // TODO: fetch from guest/events/issues
   // Example: if guests expected
   if s.coord.Guest != nil && s.coord.Guest.ExpectedToday() {
       todayContext = "Guests are expected today."
   }
   // Example: if issues
   if s.coord.HasIssues() {
       if todayContext != "" {
           todayContext += " "
       }
       todayContext += "Some issues need your attention."
   }
   briefing := morning.GenerateBriefing(alarmStatus, nightEvents, todayContext)
   msg := morning.FormatBriefing(briefing)
   s.respond(w, true, map[string]string{"message": msg}, "", 200)
}
// handleUIScorecard returns a simple system quality scorecard for UI display
func (s *Server) handleUIScorecard(w http.ResponseWriter, r *http.Request) {
   if r.Method != "GET" {
       s.respond(w, false, nil, "invalid method", 400)
       return
   }
   // Example: replace with real checks
   securityOK := s.coord.Alarm != nil && s.coord.Alarm.IsArmed()
   stable := !s.coord.HasIssues()
   aware := s.coord.HA != nil && s.coord.HA.IsConnected()
   score := scorecard.ComputeStatusScore(securityOK, stable, aware)
   s.respond(w, true, score, "", 200)
}
// handleTelemetrySummary returns aggregated telemetry summary (admin-only)
// GET /api/admin/telemetry/summary
// Returns: aggregated feature usage, error categories, and performance buckets (no personal data)
func (s *Server) handleTelemetrySummary(w http.ResponseWriter, r *http.Request) {
	role := getRole(r)
	if role != auth.Admin {
		s.respond(w, false, nil, "admin only", 403)
		return
	}
	if r.Method != "GET" {
		s.respond(w, false, nil, "invalid method", 400)
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
		s.respond(w, false, nil, "admin only", 403)
		return
	}
	if r.Method != "POST" {
		s.respond(w, false, nil, "invalid method", 400)
		return
	}
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respond(w, false, nil, "invalid json", 400)
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
		s.respond(w, false, nil, "admin only", 403)
		return
	}
	if r.Method != "GET" {
		s.respond(w, false, nil, "invalid method", 400)
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
		s.respond(w, false, nil, "admin only", 403)
		return
	}
	if r.Method != "POST" {
		s.respond(w, false, nil, "invalid method", 400)
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
		s.respond(w, false, nil, "invalid json", 400)
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
