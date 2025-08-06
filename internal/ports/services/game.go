// Package services defines the interfaces for business logic services.
// This is part of the Ports layer in Hexagonal Architecture.
// These interfaces define the contracts for application services.
package services

import (
	"context"

	"chess-backend/internal/domain/game"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateGameRequest represents the data needed to create a new game
type CreateGameRequest struct {
	PlayerID primitive.ObjectID `json:"player_id"`
}

// JoinGameRequest represents the data needed to join a game
type JoinGameRequest struct {
	GameID   primitive.ObjectID `json:"game_id"`
	PlayerID primitive.ObjectID `json:"player_id"`
}

// MakeMoveRequest represents the data needed to make a move
type MakeMoveRequest struct {
	GameID   primitive.ObjectID `json:"game_id"`
	PlayerID primitive.ObjectID `json:"player_id"`
	From     string             `json:"from"`
	To       string             `json:"to"`
	Piece    string             `json:"piece"`
	Notation string             `json:"notation"`
}

// ResignGameRequest represents the data needed to resign from a game
type ResignGameRequest struct {
	GameID   primitive.ObjectID `json:"game_id"`
	PlayerID primitive.ObjectID `json:"player_id"`
}

// GameResponse represents the response for game operations
type GameResponse struct {
	Message string     `json:"message"`
	Game    *game.Game `json:"game,omitempty"`
	GameID  string     `json:"game_id,omitempty"`
}

// GameListResponse represents the response for listing games
type GameListResponse struct {
	Games []game.Game `json:"games"`
	Total int64       `json:"total"`
	Page  int          `json:"page"`
	Limit int          `json:"limit"`
}

// GameService defines the interface for game business logic
type GameService interface {
	// CreateGame creates a new chess game
	CreateGame(ctx context.Context, req CreateGameRequest) (*GameResponse, error)

	// JoinGame allows a player to join an existing game
	JoinGame(ctx context.Context, req JoinGameRequest) (*GameResponse, error)

	// GetGame retrieves a game by ID
	GetGame(ctx context.Context, gameID primitive.ObjectID, playerID primitive.ObjectID) (*game.Game, error)

	// MakeMove processes a move in a game
	MakeMove(ctx context.Context, req MakeMoveRequest) (*GameResponse, error)

	// ResignGame allows a player to resign from a game
	ResignGame(ctx context.Context, req ResignGameRequest) (*GameResponse, error)

	// ListPlayerGames retrieves all games for a specific player
	ListPlayerGames(ctx context.Context, playerID primitive.ObjectID, page, limit int) (*GameListResponse, error)

	// ListWaitingGames retrieves all games waiting for players
	ListWaitingGames(ctx context.Context, page, limit int) (*GameListResponse, error)

	// ListActiveGames retrieves all active games
	ListActiveGames(ctx context.Context, page, limit int) (*GameListResponse, error)

	// GetGameHistory retrieves the move history of a game
	GetGameHistory(ctx context.Context, gameID primitive.ObjectID, playerID primitive.ObjectID) ([]game.Move, error)

	// IsPlayerInGame checks if a player is part of a specific game
	IsPlayerInGame(ctx context.Context, gameID primitive.ObjectID, playerID primitive.ObjectID) (bool, error)

	// GetPlayerStats retrieves statistics for a player
	GetPlayerStats(ctx context.Context, playerID primitive.ObjectID) (map[string]interface{}, error)

	// DeleteGame removes a game (admin function)
	DeleteGame(ctx context.Context, gameID primitive.ObjectID) error
}