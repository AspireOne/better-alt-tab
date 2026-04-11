package main

import (
	"log"
	"os"

	"better_alt_tab/internal/app"
	"better_alt_tab/internal/config"
	"better_alt_tab/internal/startup"
	"better_alt_tab/internal/theme"
)

func main() {
	logger := log.New(os.Stderr, "better-alt-tab: ", log.LstdFlags|log.Lmicroseconds)
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
	overlayTheme, err := config.LoadTheme(cfg.Theme)
	if err != nil {
		logger.Printf("theme warning: %v; using defaults", err)
		overlayTheme = theme.Default()
	}
	if err := app.Run(logger, cfg, overlayTheme); err != nil {
		logger.Fatalf("run: %v", err)
	}
}
