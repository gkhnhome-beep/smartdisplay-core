package main

import (
	"net/http"
	"os"
	"runtime"
	"smartdisplay-core/internal/alarm"
	"smartdisplay-core/internal/alarm/countdown"
	"smartdisplay-core/internal/api"
	"smartdisplay-core/internal/config"
	"smartdisplay-core/internal/guest"
	"smartdisplay-core/internal/haadapter"
	"smartdisplay-core/internal/hanotify"
	"smartdisplay-core/internal/health"
	"smartdisplay-core/internal/logger"
	"smartdisplay-core/internal/system"
	"strconv"
)

func main() {
	// Set GOMAXPROCS if env is set
	if v := os.Getenv("GOMAXPROCS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			runtime.GOMAXPROCS(n)
			logger.Info("GOMAXPROCS set to " + v)
		}
	}
	logger.Init()
	cfg := config.Load()
	var coord *system.Coordinator
	http.HandleFunc("/health", health.HealthHandler)
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
	alarmSM := alarm.NewStateMachine()
	guestSM := guest.NewStateMachine()
	cd := countdown.New(30)
	notifier := &hanotify.StubNotifier{}
	coord = system.NewCoordinator(alarmSM, guestSM, cd, adapter, notifier, cfg)
	logger.Info("system coordinator ready")
	coord.StartHealthMonitor()
	health.SetCoordinator(coord)
	apiServer := api.NewServer(coord)
	if err := apiServer.Start(8090); err != nil {
		logger.Error("failed to start API server: " + err.Error())
		os.Exit(1)
	}
	logger.Info("ui api ready")
	select {}
}
