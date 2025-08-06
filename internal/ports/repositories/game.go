// Package repositories defines the interfaces for data persistence.
// This is part of the Ports layer in Hexagonal Architecture.
// Ports define contracts that adapters must implement.
package repositories

import (
	"context"

	"chess-backend/internal/domain/game"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GameRepository defines the interface for game data persistence
type GameRepository interface {
	// Save creates or updates a game in the repository
	Save(ctx context.Context, game *game.Game) error

	// FindByID retrieves a game by its ID
	FindByID(ctx context.Context, id primitive.ObjectID) (*game.Game, error)

	// Update updates an existing game in the repository
	Update(ctx context.Context, game *game.Game) error

	// Delete removes a game from the repository
	Delete(ctx context.Context, id primitive.ObjectID) error

	// FindByPlayerID retrieves all games for a specific player
	FindByPlayerID(ctx context.Context, playerID primitive.ObjectID) ([]*game.Game, error)

	// FindActiveGames retrieves all active games
	FindActiveGames(ctx context.Context) ([]*game.Game, error)

	// FindWaitingGames retrieves all games waiting for players
	FindWaitingGames(ctx context.Context) ([]*game.Game, error)

	// FindByStatus retrieves games by their status
	FindByStatus(ctx context.Context, status game.GameStatus) ([]*game.Game, error)

	// FindByPlayers retrieves a game between two specific players
	FindByPlayers(ctx context.Context, player1, player2 primitive.ObjectID) (*game.Game, error)

	// List retrieves games with pagination
	List(ctx context.Context, offset, limit int) ([]*game.Game, error)

	// Count returns the total number of games
	Count(ctx context.Context) (int64, error)

	// CountByStatus returns the number of games with a specific status
	CountByStatus(ctx context.Context, status game.GameStatus) (int64, error)

	// CountByPlayer returns the number of games for a specific player
	CountByPlayer(ctx context.Context, playerID primitive.ObjectID) (int64, error)
}