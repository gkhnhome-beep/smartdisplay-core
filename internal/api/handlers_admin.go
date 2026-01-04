package api

import (
	"net/http"
	"os"
	"smartdisplay-core/internal/audit"
	"smartdisplay-core/internal/auth"
	"syscall"
	"time"
)

// handleAdminSmoke runs self-check and returns summary (admin-only)
func (s *Server) handleAdminSmoke(w http.ResponseWriter, r *http.Request) {
	role := getRole(r)
	if role != auth.Admin {
		s.respondError(w, r, CodeForbidden, "admin required")
		return
	}
	if r.Method != "POST" {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
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

// handleAdminRestart triggers graceful shutdown and exits with code 42 (admin-only)
func (s *Server) handleAdminRestart(w http.ResponseWriter, r *http.Request) {
	role := getRole(r)
	if role != auth.Admin {
		s.respondError(w, r, CodeForbidden, "admin required")
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
