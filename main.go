package main

import (
	"log"
	"os"

	"github.com/systask/systask/internal/app"
	"github.com/systask/systask/internal/config"
)

func main() {
	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Warning: Could not load config: %v, using defaults", err)
		cfg = config.Default()
	}

	// Create and run the application
	application := app.New(cfg)
	if err := application.Run(); err != nil {
		log.Printf("Application error: %v", err)
		os.Exit(1)
	}
}
