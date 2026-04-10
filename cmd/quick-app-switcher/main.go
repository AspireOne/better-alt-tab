package main

import (
	"log"
	"os"

	"quick_app_switcher/internal/app"
)

func main() {
	logger := log.New(os.Stderr, "quick-app-switcher: ", log.LstdFlags|log.Lmicroseconds)
	if err := app.Run(logger); err != nil {
		logger.Fatalf("run: %v", err)
	}
}
