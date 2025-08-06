// Package game contains HTTP handlers for game-related operations.
// This is part of the Adapters layer in Hexagonal Architecture.
package game

import (
	"encoding/json"
	"net/http"
	"strconv"

	"chess-backend/internal/ports/services"
	"chess-backend/internal/utils"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GameHandlers contains all HTTP handlers for game operations
type GameHandlers struct {
	gameService services.GameService
}

// NewGameHandlers creates a new instance of GameHandlers
func NewGameHandlers(gameService services.GameService) *GameHandlers {
	return &GameHandlers{
		gameService: gameService,
	}
}

// CreateGameHandler handles POST /api/game/create
func (h *GameHandlers) CreateGameHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(primitive.ObjectID)
	if !ok {
		utils.Response.WriteUnauthorized(w, "User not authenticated")
		return
	}

	// Create game request
	req := services.CreateGameRequest{
		PlayerID: userID,
	}

	// Call service
	createResponse, err := h.gameService.CreateGame(r.Context(), req)
	if err != nil {
		utils.Response.WriteBadRequest(w, err.Error())
		return
	}

	utils.Response.WriteCreated(w, "Game created successfully", createResponse)
}

// JoinGameHandler handles POST /api/game/join/{gameId}
func (h *GameHandlers) JoinGameHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("user_id").(primitive.ObjectID)
	if !ok {
		utils.Response.WriteUnauthorized(w, "User not authenticated")
		return
	}

	// Get game ID from URL
	vars := mux.Vars(r)
	gameIDStr, exists := vars["gameId"]
	if !exists {
		utils.Response.WriteBadRequest(w, "Game ID is required")
		return
	}

	gameID, err := primitive.ObjectIDFromHex(gameIDStr)
	if err != nil {
		utils.Response.WriteBadRequest(w, "Invalid game ID format")
		return
	}

	// Create join request
	req := services.JoinGameRequest{
		GameID:   gameID,
		PlayerID: userID,
	}

	// Call service
	joinResponse, err := h.gameService.JoinGame(r.Context(), req)
	if err != nil {
		utils.Response.WriteBadRequest(w, err.Error())
		return
	}

	utils.Response.WriteSuccess(w, joinResponse.Message, joinResponse)
}

// GetGameHandler handles GET /api/game/{gameId}
func (h *GameHandlers) GetGameHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("user_id").(primitive.ObjectID)
	if !ok {
		utils.Response.WriteUnauthorized(w, "User not authenticated")
		return
	}

	// Get game ID from URL
	vars := mux.Vars(r)
	gameIDStr, exists := vars["gameId"]
	if !exists {
		utils.Response.WriteBadRequest(w, "Game ID is required")
		return
	}

	gameID, err := primitive.ObjectIDFromHex(gameIDStr)
	if err != nil {
		utils.Response.WriteBadRequest(w, "Invalid game ID format")
		return
	}

	// Call service
	game, err := h.gameService.GetGame(r.Context(), gameID, userID)
	if err != nil {
		utils.Response.WriteBadRequest(w, err.Error())
		return
	}

	utils.Response.WriteSuccess(w, "Game retrieved successfully", game)
}

// MoveHandler handles POST /api/game/{gameId}/move
func (h *GameHandlers) MoveHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("user_id").(primitive.ObjectID)
	if !ok {
		utils.Response.WriteUnauthorized(w, "User not authenticated")
		return
	}

	// Get game ID from URL
	vars := mux.Vars(r)
	gameIDStr, exists := vars["gameId"]
	if !exists {
		utils.Response.WriteBadRequest(w, "Game ID is required")
		return
	}

	gameID, err := primitive.ObjectIDFromHex(gameIDStr)
	if err != nil {
		utils.Response.WriteBadRequest(w, "Invalid game ID format")
		return
	}

	// Parse request body
	var moveData struct {
		From     string `json:"from"`
		To       string `json:"to"`
		Piece    string `json:"piece"`
		Notation string `json:"notation"`
	}

	if decodeErr := json.NewDecoder(r.Body).Decode(&moveData); decodeErr != nil {
		utils.Response.WriteBadRequest(w, "Invalid request body")
		return
	}

	// Create move request
	req := services.MakeMoveRequest{
		GameID:   gameID,
		PlayerID: userID,
		From:     moveData.From,
		To:       moveData.To,
		Piece:    moveData.Piece,
		Notation: moveData.Notation,
	}

	// Call service
	moveResponse, err := h.gameService.MakeMove(r.Context(), req)
	if err != nil {
		utils.Response.WriteBadRequest(w, err.Error())
		return
	}

	utils.Response.WriteSuccess(w, moveResponse.Message, moveResponse)
}

// ResignGameHandler handles POST /api/game/{gameId}/resign
func (h *GameHandlers) ResignGameHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("user_id").(primitive.ObjectID)
	if !ok {
		utils.Response.WriteUnauthorized(w, "User not authenticated")
		return
	}

	// Get game ID from URL
	vars := mux.Vars(r)
	gameIDStr, exists := vars["gameId"]
	if !exists {
		utils.Response.WriteBadRequest(w, "Game ID is required")
		return
	}

	gameID, err := primitive.ObjectIDFromHex(gameIDStr)
	if err != nil {
		utils.Response.WriteBadRequest(w, "Invalid game ID format")
		return
	}

	// Create resign request
	req := services.ResignGameRequest{
		GameID:   gameID,
		PlayerID: userID,
	}

	// Call service
	resignResponse, err := h.gameService.ResignGame(r.Context(), req)
	if err != nil {
		utils.Response.WriteBadRequest(w, err.Error())
		return
	}

	utils.Response.WriteSuccess(w, resignResponse.Message, resignResponse)
}

// ListPlayerGamesHandler handles GET /api/game/my-games
func (h *GameHandlers) ListPlayerGamesHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("user_id").(primitive.ObjectID)
	if !ok {
		utils.Response.WriteUnauthorized(w, "User not authenticated")
		return
	}

	// Parse pagination parameters
	page := 1
	limit := 10

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Call service
	gamesResponse, err := h.gameService.ListPlayerGames(r.Context(), userID, page, limit)
	if err != nil {
		utils.Response.WriteBadRequest(w, err.Error())
		return
	}

	utils.Response.WriteSuccess(w, "Player games retrieved successfully", gamesResponse)
}

// ListWaitingGamesHandler handles GET /api/game/waiting
func (h *GameHandlers) ListWaitingGamesHandler(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	page := 1
	limit := 10

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Call service
	waitingResponse, err := h.gameService.ListWaitingGames(r.Context(), page, limit)
	if err != nil {
		utils.Response.WriteBadRequest(w, err.Error())
		return
	}

	utils.Response.WriteSuccess(w, "Waiting games retrieved successfully", waitingResponse)
}

// ListActiveGamesHandler handles GET /api/game/active
func (h *GameHandlers) ListActiveGamesHandler(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	page := 1
	limit := 10

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Call service
	activeResponse, err := h.gameService.ListActiveGames(r.Context(), page, limit)
	if err != nil {
		utils.Response.WriteBadRequest(w, err.Error())
		return
	}

	utils.Response.WriteSuccess(w, "Active games retrieved successfully", activeResponse)
}

// GetGameHistoryHandler handles GET /api/game/{gameId}/history
func (h *GameHandlers) GetGameHistoryHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("user_id").(primitive.ObjectID)
	if !ok {
		utils.Response.WriteUnauthorized(w, "User not authenticated")
		return
	}

	// Get game ID from URL
	vars := mux.Vars(r)
	gameIDStr, exists := vars["gameId"]
	if !exists {
		utils.Response.WriteBadRequest(w, "Game ID is required")
		return
	}

	gameID, err := primitive.ObjectIDFromHex(gameIDStr)
	if err != nil {
		utils.Response.WriteBadRequest(w, "Invalid game ID format")
		return
	}

	// Call service
	moves, err := h.gameService.GetGameHistory(r.Context(), gameID, userID)
	if err != nil {
		utils.Response.WriteBadRequest(w, err.Error())
		return
	}

	utils.Response.WriteSuccess(w, "Game history retrieved successfully", moves)
}

// GetPlayerStatsHandler handles GET /api/game/stats
func (h *GameHandlers) GetPlayerStatsHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("user_id").(primitive.ObjectID)
	if !ok {
		utils.Response.WriteUnauthorized(w, "User not authenticated")
		return
	}

	// Call service
	stats, err := h.gameService.GetPlayerStats(r.Context(), userID)
	if err != nil {
		utils.Response.WriteBadRequest(w, err.Error())
		return
	}

	utils.Response.WriteSuccess(w, "Player stats retrieved successfully", stats)
}