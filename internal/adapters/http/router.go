// Package http provides HTTP routing and server setup.
// This is part of the Adapters layer in Hexagonal Architecture.
package http

import (
	"net/http"

	"chess-backend/internal/adapters/http/auth"
	"chess-backend/internal/adapters/http/game"
	"chess-backend/internal/ports/services"

	"github.com/gorilla/mux"
)

// Server represents the HTTP server
type Server struct {
	router         *mux.Router
	authHandler    *auth.Handler
	gameHandler    *game.GameHandlers
	authMiddleware *AuthMiddleware
}

// NewServer creates a new HTTP server
func NewServer(authService services.AuthService, gameService services.GameService) *Server {
	router := mux.NewRouter()

	// Create handlers
	authHandler := auth.NewHandler(authService)
	authMiddleware := NewAuthMiddleware(authService)

	// Create game handler if gameService is provided
	var gameHandler *game.GameHandlers
	if gameService != nil {
		gameHandler = game.NewGameHandlers(gameService)
	}

	server := &Server{
		router:         router,
		authHandler:    authHandler,
		gameHandler:    gameHandler,
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

	// Protected game routes
	if s.gameHandler != nil {
		gameRoutes := api.PathPrefix("/game").Subrouter()
		gameRoutes.Use(s.authMiddleware.RequireAuth)
		s.registerGameRoutes(gameRoutes)
	}

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

// registerGameRoutes registers all game-related routes
func (s *Server) registerGameRoutes(router *mux.Router) {
	// Game management routes
	router.HandleFunc("/create", s.gameHandler.CreateGameHandler).Methods("POST")
	router.HandleFunc("/join/{gameId}", s.gameHandler.JoinGameHandler).Methods("POST")
	router.HandleFunc("/{gameId}", s.gameHandler.GetGameHandler).Methods("GET")
	router.HandleFunc("/{gameId}/move", s.gameHandler.MoveHandler).Methods("POST")
	router.HandleFunc("/{gameId}/resign", s.gameHandler.ResignGameHandler).Methods("POST")
	router.HandleFunc("/{gameId}/history", s.gameHandler.GetGameHistoryHandler).Methods("GET")

	// Game listing routes
	router.HandleFunc("/my-games", s.gameHandler.ListPlayerGamesHandler).Methods("GET")
	router.HandleFunc("/waiting", s.gameHandler.ListWaitingGamesHandler).Methods("GET")
	router.HandleFunc("/active", s.gameHandler.ListActiveGamesHandler).Methods("GET")

	// Player stats route
	router.HandleFunc("/stats", s.gameHandler.GetPlayerStatsHandler).Methods("GET")
}

// Start starts the HTTP server on the specified address
func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.router)
}