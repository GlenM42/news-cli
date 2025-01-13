package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
)

// loadEnv loads environment variables from a .env file.
func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func main() {
	// Load environment variables
	loadEnv()

	// Load authorized keys from JSON file
	err := loadAuthorizedKeys("authorized_keys.json")
	if err != nil {
		log.Fatalf("Could not load authorized keys: %v", err)
	}

	// Retrieve host and port from environment variables
	host := getEnvOrDefault("SSH_HOST", "localhost")
	port := getEnvOrDefault("SSH_PORT", "23234")

	// Create a new app instance
	app := newApp()

	// Start the SSH server in a separate goroutine
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", host, "port", port)

	go func() {
		if err := app.ListenAndServe(); err != nil {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		log.Error("Could not stop server", "error", err)
	}
}
