// Package http provides HTTP routing and server setup.
// This is part of the Adapters layer in Hexagonal Architecture.
package http

import (
	"net/http"

	"chess-backend/internal/adapters/http/auth"
	"chess-backend/internal/ports/services"

	"github.com/gorilla/mux"
)

// Server represents the HTTP server
type Server struct {
	router      *mux.Router
	authHandler *auth.Handler
	authMiddleware *AuthMiddleware
}

// NewServer creates a new HTTP server
func NewServer(authService services.AuthService, gameService services.GameService) *Server {
	router := mux.NewRouter()

	// Create handlers
	authHandler := auth.NewHandler(authService)
	authMiddleware := NewAuthMiddleware(authService)

	server := &Server{
		router:         router,
		authHandler:    authHandler,
		authMiddleware: authMiddleware,
	}

	// Setup routes
	server.setupRoutes()

	return server
}

// setupRoutes configures all application routes
func (s *Server) setupRoutes() {
	// Apply global middleware
	s.router.Use(LoggingMiddleware)
	s.router.Use(CORSMiddleware)

	// API routes
	api := s.router.PathPrefix("/api").Subrouter()

	// Health check endpoint
	api.HandleFunc("/health", s.healthCheckHandler).Methods("GET")

	// Authentication routes (public)
	s.authHandler.RegisterRoutes(api)

	// Protected routes would go here
	// Example:
	// protected := api.PathPrefix("/protected").Subrouter()
	// protected.Use(s.authMiddleware.RequireAuth)
	// gameHandler.RegisterRoutes(protected)

	// Serve static files (if needed)
	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/"))).Methods("GET")
}

// healthCheckHandler provides a simple health check endpoint
func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","message":"Server is running"}`))
}

// GetRouter returns the configured router
func (s *Server) GetRouter() *mux.Router {
	return s.router
}

// Start starts the HTTP server on the specified address
func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.router)
}