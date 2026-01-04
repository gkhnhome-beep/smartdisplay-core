package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"smartdisplay-core/internal/alarm"
	"smartdisplay-core/internal/alarm/countdown"
	"smartdisplay-core/internal/api"
	"smartdisplay-core/internal/config"
	"smartdisplay-core/internal/firstboot"
	"smartdisplay-core/internal/guest"
	"smartdisplay-core/internal/haadapter"
	"smartdisplay-core/internal/hal"
	"smartdisplay-core/internal/hanotify"
	"smartdisplay-core/internal/health"
	"smartdisplay-core/internal/i18n"
	"smartdisplay-core/internal/logger"
	"smartdisplay-core/internal/platform"
	"smartdisplay-core/internal/system"
	"smartdisplay-core/internal/version"
	"strconv"
	"syscall"
	"time"
)

func main() {
	// Startup order: deterministic, panic-safe

	// 0. Log version (first visible output)
	logger.Init()
	logger.Info("[RC] SmartDisplay v" + version.Version)
	if version.Commit != "" {
		logger.Info("[RC] Commit: " + version.Commit)
	}

	// 1. Logger initialization (first, all logs depend on this)
	setGOMAXPROCS()

	// 2. Config loading
	runtimeCfg, err := loadRuntimeConfig()
	if err != nil {
		logger.Error("runtime config load failed (using defaults): " + err.Error())
		runtimeCfg = &config.RuntimeConfig{Language: "en"}
	}

	// 3. i18n initialization (language preferences)
	initializeI18n(runtimeCfg)

	// 4. Accessibility initialization (from runtime config)
	cfg := config.Load()

	// 5. Voice initialization (from runtime config)
	// (Voice subsystem initialization happens through coordinator)

	// 6. FirstBoot initialization (from runtime config)
	// (FirstBoot is initialized as part of coordinator)

	// 7. Coordinator and subsystems (HA adapter, alarm, guest, etc.)
	coord := initializeCoordinator(cfg, runtimeCfg)

	// 8. Health monitoring startup
	coord.StartHealthMonitor()
	health.SetCoordinator(coord)

	// 8.5 A2: Alarm polling startup (read-only Alarmo sync)
	// Create cancellable context for graceful shutdown
	pollCtx, pollCancel := context.WithCancel(context.Background())
	coord.StartAlarmPolling(pollCtx)

	// 9. HTTP server startup
	apiServer := api.NewServer(coord)
	if err := apiServer.Start(8090); err != nil {
		logger.Error("failed to start API server: " + err.Error())
		os.Exit(1)
	}
	logger.Info("ui api ready")

	// 10. Graceful shutdown handling (blocks on signal)
	handleGracefulShutdown(apiServer, pollCancel)
}

// setGOMAXPROCS sets GOMAXPROCS if env var is set
func setGOMAXPROCS() {
	if v := os.Getenv("GOMAXPROCS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			runtime.GOMAXPROCS(n)
			logger.Info("GOMAXPROCS set to " + v)
		}
	}
}

// loadRuntimeConfig loads or creates default runtime config
func loadRuntimeConfig() (*config.RuntimeConfig, error) {
	runtimeCfg, err := config.LoadRuntimeConfig()
	if err != nil {
		return nil, err
	}
	return runtimeCfg, nil
}

// initializeI18n initializes internationalization with preferred language
func initializeI18n(runtimeCfg *config.RuntimeConfig) {
	lang := runtimeCfg.Language
	if lang == "" {
		lang = "en"
	}
	if err := i18n.Init(lang); err != nil {
		logger.Error("i18n init failed: " + err.Error())
	} else {
		logger.Info("i18n initialized: language=" + lang)
	}
}

// initializeCoordinator sets up all subsystems and the coordinator
func initializeCoordinator(cfg config.Config, runtimeCfg *config.RuntimeConfig) *system.Coordinator {
	// Health check endpoint (needed before HA adapter)
	http.HandleFunc("/health", health.HealthHandler)

	// Initialize HA adapter
	adapter := haadapter.New()
	if err := adapter.Init(); err != nil {
		logger.Error("ha adapter init failed: " + err.Error())
		os.Exit(1)
	}
	if err := adapter.Start(); err != nil {
		logger.Error("ha adapter start failed: " + err.Error())
		os.Exit(1)
	}
	logger.Info("ha adapter ready")

	// Initialize state machines
	alarmSM := alarm.NewStateMachine()
	guestSM := guest.NewStateMachine()
	cd := countdown.New(30)
	notifier := &hanotify.StubNotifier{}

	// Initialize HAL registry and platform
	halReg := hal.NewRegistry()
	plat := platform.DetectPlatform()

	// Create coordinator (integrates all subsystems)
	// A2: Pass HA connection details for Alarmo adapter
	haBaseURL := os.Getenv("HA_BASE_URL")
	haToken := runtimeCfg.HAToken // Use decrypted token from runtime config
	coord := system.NewCoordinator(alarmSM, guestSM, cd, adapter, notifier, halReg, plat, haBaseURL, haToken)
	logger.Info("system coordinator ready")

	// Apply accessibility preferences
	applyAccessibilityPreferences(coord, runtimeCfg)

	// Apply voice preferences
	applyVoicePreferences(coord, runtimeCfg)

	// Initialize first-boot flow
	initializeFirstBoot(coord, runtimeCfg)

	return coord
}

// applyAccessibilityPreferences applies saved accessibility settings
func applyAccessibilityPreferences(coord *system.Coordinator, runtimeCfg *config.RuntimeConfig) {
	// TODO: Apply reduced_motion to AI engine when SetReducedMotion() method is implemented
	if runtimeCfg.ReducedMotion {
		logger.Info("accessibility: reduced_motion enabled at startup")
	}
}

// applyVoicePreferences applies saved voice feedback settings
func applyVoicePreferences(coord *system.Coordinator, runtimeCfg *config.RuntimeConfig) {
	// TODO: Integrate voice manager to Coordinator when Voice field is implemented
	if runtimeCfg.VoiceEnabled {
		logger.Info("voice: feedback enabled at startup")
	}
}

// initializeFirstBoot initializes the first-boot flow if needed
func initializeFirstBoot(coord *system.Coordinator, runtimeCfg *config.RuntimeConfig) {
	if coord.FirstBoot != nil {
		coord.FirstBoot = firstboot.New(runtimeCfg.WizardCompleted == false)
		if coord.FirstBoot.Active() {
			logger.Info("firstboot: wizard activated (wizard_completed=false)")
		}
	}
}

// handleGracefulShutdown registers signal handlers and blocks until shutdown is complete
func handleGracefulShutdown(apiServer *api.Server, pollCancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Block until signal received
	sig := <-sigChan
	logger.Info("shutdown signal received: " + sig.String())

	// A2: Stop alarm polling
	if pollCancel != nil {
		pollCancel()
	}

	// Graceful shutdown with 10-second timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := apiServer.ShutdownCtx(shutdownCtx); err != nil {
		logger.Error("http server shutdown error: " + err.Error())
	}

	// Flush logs before exit
	logger.Info("shutdown complete")
	os.Exit(0)
}
