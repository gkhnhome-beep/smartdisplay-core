package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

var logger *log.Logger
var debugMode bool

const logPath = "logs/app.log"
const maxLogSize = 5 * 1024 * 1024 // 5MB
const maxLogFiles = 5

func Init() {
	rotateLogIfNeeded()
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	multi := io.MultiWriter(os.Stdout, &rotatingWriter{file: file})
	logger = log.New(multi, "", log.LstdFlags)
	debugMode = os.Getenv("DEBUG") == "true"
}

// rotatingWriter wraps an *os.File and rotates on write threshold
type rotatingWriter struct {
	file *os.File
}

func (w *rotatingWriter) Write(p []byte) (n int, err error) {
	fi, err := w.file.Stat()
	if err == nil && fi.Size()+int64(len(p)) > maxLogSize {
		w.file.Close()
		rotateLogIfNeeded()
		w.file, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return 0, err
		}
	}
	return w.file.Write(p)
}

// rotateLogIfNeeded rotates logs/app.log if >5MB, keeps last 5 rotated files
func rotateLogIfNeeded() {
	fi, err := os.Stat(logPath)
	if err == nil && fi.Size() > maxLogSize {
		// Remove oldest if needed
		os.Remove(fmt.Sprintf("logs/app.log.%d", maxLogFiles))
		// Shift files
		for i := maxLogFiles - 1; i >= 1; i-- {
			old := fmt.Sprintf("logs/app.log.%d", i)
			new := fmt.Sprintf("logs/app.log.%d", i+1)
			os.Rename(old, new)
		}
		os.Rename(logPath, fmt.Sprintf("logs/app.log.1"))
	}
}
func Debug(msg string) {
	if debugMode {
		logger.Printf("[DEBUG] %s", msg)
	}
}

func IsDebug() bool {
	return debugMode
}

func Info(msg string) {
	logger.Printf("[INFO] %s", msg)
}

func Error(msg string) {
	logger.Printf("[ERROR] %s", msg)
}
