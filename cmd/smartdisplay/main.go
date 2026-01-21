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
	"smartdisplay-core/internal/settings"
	"smartdisplay-core/internal/system"
	"smartdisplay-core/internal/version"
	"strconv"
	"syscall"
	"time"
)

func main() {
	// Sade başlatma ve shutdown
	logger.Init()
	logger.Info("SmartDisplay v" + version.Version)
	setGOMAXPROCS()
	runtimeCfg, err := loadRuntimeConfig()
	if err != nil {
		logger.Error("runtime config load failed: " + err.Error())
		runtimeCfg = &config.RuntimeConfig{Language: "en"}
	}
	initializeI18n(runtimeCfg)
	cfg := config.Load()
	coord := initializeCoordinator(cfg, runtimeCfg)
	coord.StartHealthMonitor()
	health.SetCoordinator(coord)
	pollCtx, pollCancel := context.WithCancel(context.Background())
	coord.StartAlarmPolling(pollCtx)
	settings.InitGlobalHealthMonitor()
	apiServer := api.NewServer(coord, runtimeCfg)
	if err := apiServer.Start(8090); err != nil {
		logger.Error("failed to start API server: " + err.Error())
		os.Exit(1)
	}
	logger.Info("ui api ready")
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
	// FAZ S2: Load HA config from secure storage, decrypt token
	haBaseURL := os.Getenv("HA_BASE_URL")
	haToken := ""

	// Load HA config from file if already configured by user
	haConfig, err := settings.LoadHAConfig()
	if err == nil && haConfig != nil && haConfig.ServerURL != "" {
		haBaseURL = haConfig.ServerURL
		// Decrypt token from secure storage
		decryptedToken, err := settings.DecryptToken()
		if err != nil {
			logger.Error("failed to decrypt ha token at startup: " + err.Error())
		} else if decryptedToken != "" {
			haToken = decryptedToken
			logger.Info("ha config loaded from secure storage at startup")
		}
	}

	// Fallback: Try environment variables if not configured in secure storage
	if haBaseURL == "" {
		haBaseURL = os.Getenv("HA_BASE_URL")
	}
	if haToken == "" {
		haToken = os.Getenv("HA_TOKEN")
	}

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
	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("http server shutdown error: " + err.Error())
	}

	// Flush logs before exit
	logger.Info("shutdown complete")
	os.Exit(0)
}
