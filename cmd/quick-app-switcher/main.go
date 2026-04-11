package main

import (
	"log"
	"os"

	"quick_app_switcher/internal/app"
	"quick_app_switcher/internal/config"
)

func main() {
	logger := log.New(os.Stderr, "quick-app-switcher: ", log.LstdFlags|log.Lmicroseconds)
	cfg, err := config.Load()
	if err != nil {
		logger.Printf("config warning: %v; using defaults", err)
		cfg = config.Default()
	}
	if err := app.Run(logger, cfg); err != nil {
		logger.Fatalf("run: %v", err)
	}
}
