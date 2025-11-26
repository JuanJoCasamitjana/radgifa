package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	_ "embed"
	"radgifa/internal/server"
)

const (
	appName    = "Radgifa API"
	appVersion = "0.1.0"
	ansiReset  = "\u001B[0m"
	ansiCyan   = "\u001B[36m"
	ansiPurple = "\u001B[35m"
)

//go:embed banner.txt
var banner string
var (
	startTime           = time.Now().Format(time.RFC3339)
	startupText         = fmt.Sprintf("%s%s v%s%s\nStarted at: %s\n", ansiCyan, appName, appVersion, ansiReset, startTime)
	coloredBanner       = fmt.Sprintf("%s%s%s", ansiPurple, banner, ansiReset)
	finalStartupMessage = fmt.Sprintf("%s\n%s", coloredBanner, startupText)
)

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")
	stop() // Allow Ctrl+C to force shutdown

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

func main() {

	server := server.NewServer()

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(server, done)

	// Print startup message
	fmt.Println(finalStartupMessage)

	// Start the server
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}
