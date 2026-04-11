package main

import (
	"log"
	"os"

	"quick_app_switcher/internal/app"
	"quick_app_switcher/internal/config"
	"quick_app_switcher/internal/startup"
)

func main() {
	logger := log.New(os.Stderr, "quick-app-switcher: ", log.LstdFlags|log.Lmicroseconds)
	cfg, err := config.Load()
	configLoaded := err == nil
	if err != nil {
		logger.Printf("config warning: %v; using defaults", err)
		cfg = config.Default()
	}
	if configLoaded {
		if err := startup.Sync(cfg.LaunchOnStartup); err != nil {
			logger.Printf("startup warning: %v", err)
		}
	}
	if err := app.Run(logger, cfg); err != nil {
		logger.Fatalf("run: %v", err)
	}
}
