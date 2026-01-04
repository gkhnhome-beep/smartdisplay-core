package api

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"smartdisplay-core/internal/auth"
	"smartdisplay-core/internal/logger"
	"strings"
)

// handleAdminBackup streams a zip of config/runtime files (admin-only)
func (s *Server) handleAdminBackup(w http.ResponseWriter, r *http.Request) {
	role := getRole(r)
	if role != auth.Admin {
		s.respondError(w, r, CodeForbidden, "admin required")
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
		s.respondError(w, r, CodeForbidden, "admin required")
		return
	}
	if r.Method != "POST" {
		s.respondError(w, r, CodeMethodNotAllowed, "POST required")
		return
	}
	// Read zip from body
	tmpZip, err := ioutil.TempFile("", "restore-*.zip")
	if err != nil {
		s.respondError(w, r, CodeInternalError, "temporary file error")
		return
	}
	defer os.Remove(tmpZip.Name())
	io.Copy(tmpZip, r.Body)
	tmpZip.Close()
	zr, err := zip.OpenReader(tmpZip.Name())
	if err != nil {
		s.respondError(w, r, CodeBadRequest, "invalid zip format")
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
		s.respondError(w, r, CodeBadRequest, "missing required files")
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
	logger.Info("config restore completed (restart recommended)")
	s.respond(w, true, map[string]string{"result": "ok"}, "", 200)
}
