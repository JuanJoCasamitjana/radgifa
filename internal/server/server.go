package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"radgifa/internal/database"
)

type Server struct {
	port int

	service    database.Service
	kvmanager  KVManager
	httpServer *http.Server
}

func NewServer() *Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	newServer := &Server{
		port: port,

		service:   database.New(),
		kvmanager: NewKVManager(),
	}

	// Declare Server config
	newServer.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", newServer.port),
		Handler:      newServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return newServer
}

// ListenAndServe starts the HTTP server
func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

// ShutdownHTTP gracefully shuts down the HTTP server
func (s *Server) ShutdownHTTP(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// Shutdown gracefully shuts down the server and closes all resources
func (s *Server) Shutdown() error {
	if err := s.kvmanager.Close(); err != nil {
		return fmt.Errorf("failed to close KVManager: %w", err)
	}
	return nil
}
