// Package game contains the Game application service implementation.
// This is part of the Application layer in Hexagonal Architecture.
// Application services orchestrate domain entities and repository operations.
package game

import (
	"context"
	"errors"
	"fmt"

	"chess-backend/internal/domain/game"
	"chess-backend/internal/ports/repositories"
	"chess-backend/internal/ports/services"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// gameService implements the GameService interface
type gameService struct {
	gameRepo repositories.GameRepository
}

// NewGameService creates a new instance of GameService
func NewGameService(gameRepo repositories.GameRepository) services.GameService {
	return &gameService{
		gameRepo: gameRepo,
	}
}

// CreateGame creates a new chess game
func (s *gameService) CreateGame(ctx context.Context, req services.CreateGameRequest) (*services.GameResponse, error) {
	// Validate request
	if req.PlayerID.IsZero() {
		return nil, errors.New("player ID is required")
	}

	// Create new game using domain entity
	newGame, err := game.NewGame(req.PlayerID)
	if err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	// Save game to repository
	if err := s.gameRepo.Save(ctx, newGame); err != nil {
		return nil, fmt.Errorf("failed to save game: %w", err)
	}

	return &services.GameResponse{
		Message: "Game created successfully",
		Game:    newGame,
		GameID:  newGame.ID.Hex(),
	}, nil
}

// JoinGame allows a player to join an existing game
func (s *gameService) JoinGame(ctx context.Context, req services.JoinGameRequest) (*services.GameResponse, error) {
	// Validate request
	if req.GameID.IsZero() {
		return nil, errors.New("game ID is required")
	}
	if req.PlayerID.IsZero() {
		return nil, errors.New("player ID is required")
	}

	// Find the game
	gameEntity, err := s.gameRepo.FindByID(ctx, req.GameID)
	if err != nil {
		return nil, fmt.Errorf("failed to find game: %w", err)
	}

	// Join the game using domain logic
	if err := gameEntity.JoinGame(req.PlayerID); err != nil {
		return nil, fmt.Errorf("failed to join game: %w", err)
	}

	// Update game in repository
	if err := s.gameRepo.Update(ctx, gameEntity); err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	return &services.GameResponse{
		Message: "Successfully joined game",
		Game:    gameEntity,
		GameID:  gameEntity.ID.Hex(),
	}, nil
}

// GetGame retrieves a game by ID if the player is authorized
func (s *gameService) GetGame(ctx context.Context, gameID primitive.ObjectID, playerID primitive.ObjectID) (*game.Game, error) {
	// Validate input
	if gameID.IsZero() {
		return nil, errors.New("game ID is required")
	}
	if playerID.IsZero() {
		return nil, errors.New("player ID is required")
	}

	// Find the game
	gameEntity, err := s.gameRepo.FindByID(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to find game: %w", err)
	}

	// Check if player is authorized to view this game
	if !gameEntity.IsPlayerInGame(playerID) {
		return nil, errors.New("player is not authorized to view this game")
	}

	return gameEntity, nil
}

// MakeMove processes a chess move
func (s *gameService) MakeMove(ctx context.Context, req services.MakeMoveRequest) (*services.GameResponse, error) {
	// Validate request
	if req.GameID.IsZero() {
		return nil, errors.New("game ID is required")
	}
	if req.PlayerID.IsZero() {
		return nil, errors.New("player ID is required")
	}
	if req.From == "" || req.To == "" {
		return nil, errors.New("from and to positions are required")
	}

	// Find the game
	gameEntity, err := s.gameRepo.FindByID(ctx, req.GameID)
	if err != nil {
		return nil, fmt.Errorf("failed to find game: %w", err)
	}

	// Make the move using domain logic
	if err := gameEntity.MakeMove(req.PlayerID, req.From, req.To, req.Piece, req.Notation); err != nil {
		return nil, fmt.Errorf("failed to make move: %w", err)
	}

	// Update game in repository
	if err := s.gameRepo.Update(ctx, gameEntity); err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	return &services.GameResponse{
		Message: "Move made successfully",
		Game:    gameEntity,
		GameID:  gameEntity.ID.Hex(),
	}, nil
}

// ResignGame allows a player to resign from a game
func (s *gameService) ResignGame(ctx context.Context, req services.ResignGameRequest) (*services.GameResponse, error) {
	// Validate request
	if req.GameID.IsZero() {
		return nil, errors.New("game ID is required")
	}
	if req.PlayerID.IsZero() {
		return nil, errors.New("player ID is required")
	}

	// Find the game
	gameEntity, err := s.gameRepo.FindByID(ctx, req.GameID)
	if err != nil {
		return nil, fmt.Errorf("failed to find game: %w", err)
	}

	// Resign from the game using domain logic
	if err := gameEntity.ResignGame(req.PlayerID); err != nil {
		return nil, fmt.Errorf("failed to resign game: %w", err)
	}

	// Update game in repository
	if err := s.gameRepo.Update(ctx, gameEntity); err != nil {
		return nil, fmt.Errorf("failed to update game: %w", err)
	}

	return &services.GameResponse{
		Message: "Successfully resigned from game",
		Game:    gameEntity,
		GameID:  gameEntity.ID.Hex(),
	}, nil
}

// ListPlayerGames retrieves all games for a specific player with pagination
func (s *gameService) ListPlayerGames(ctx context.Context, playerID primitive.ObjectID, page, limit int) (*services.GameListResponse, error) {
	if playerID.IsZero() {
		return nil, errors.New("player ID is required")
	}

	// Get games for the player
	games, err := s.gameRepo.FindByPlayerID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to find player games: %w", err)
	}

	// Get total count
	total, err := s.gameRepo.CountByPlayer(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to count player games: %w", err)
	}

	// Convert to response format
	gameList := make([]game.Game, len(games))
	for i, g := range games {
		gameList[i] = *g
	}

	return &services.GameListResponse{
		Games: gameList,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

// ListWaitingGames retrieves all games waiting for players with pagination
func (s *gameService) ListWaitingGames(ctx context.Context, page, limit int) (*services.GameListResponse, error) {
	games, err := s.gameRepo.FindWaitingGames(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find waiting games: %w", err)
	}

	total, err := s.gameRepo.CountByStatus(ctx, game.GameStatusWaiting)
	if err != nil {
		return nil, fmt.Errorf("failed to count waiting games: %w", err)
	}

	// Convert to response format
	gameList := make([]game.Game, len(games))
	for i, g := range games {
		gameList[i] = *g
	}

	return &services.GameListResponse{
		Games: gameList,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

// ListActiveGames retrieves all active games with pagination
func (s *gameService) ListActiveGames(ctx context.Context, page, limit int) (*services.GameListResponse, error) {
	games, err := s.gameRepo.FindActiveGames(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find active games: %w", err)
	}

	total, err := s.gameRepo.CountByStatus(ctx, game.GameStatusActive)
	if err != nil {
		return nil, fmt.Errorf("failed to count active games: %w", err)
	}

	// Convert to response format
	gameList := make([]game.Game, len(games))
	for i, g := range games {
		gameList[i] = *g
	}

	return &services.GameListResponse{
		Games: gameList,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

// GetGameHistory retrieves the move history for a game
func (s *gameService) GetGameHistory(ctx context.Context, gameID primitive.ObjectID, playerID primitive.ObjectID) ([]game.Move, error) {
	if gameID.IsZero() {
		return nil, errors.New("game ID is required")
	}
	if playerID.IsZero() {
		return nil, errors.New("player ID is required")
	}

	// Find the game
	gameEntity, err := s.gameRepo.FindByID(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to find game: %w", err)
	}

	// Check if player is authorized to view this game
	if !gameEntity.IsPlayerInGame(playerID) {
		return nil, errors.New("player is not authorized to view this game")
	}

	return gameEntity.Moves, nil
}

// IsPlayerInGame checks if a player is participating in a specific game
func (s *gameService) IsPlayerInGame(ctx context.Context, gameID primitive.ObjectID, playerID primitive.ObjectID) (bool, error) {
	if gameID.IsZero() || playerID.IsZero() {
		return false, errors.New("game ID and player ID are required")
	}

	gameEntity, err := s.gameRepo.FindByID(ctx, gameID)
	if err != nil {
		return false, fmt.Errorf("failed to find game: %w", err)
	}

	return gameEntity.IsPlayerInGame(playerID), nil
}

// GetPlayerStats retrieves statistics for a player
func (s *gameService) GetPlayerStats(ctx context.Context, playerID primitive.ObjectID) (map[string]interface{}, error) {
	if playerID.IsZero() {
		return nil, errors.New("player ID is required")
	}

	// Get all games for the player
	games, err := s.gameRepo.FindByPlayerID(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to find player games: %w", err)
	}

	// Calculate statistics
	stats := map[string]interface{}{
		"total_games": len(games),
		"wins":        0,
		"losses":      0,
		"draws":       0,
		"active_games": 0,
	}

	for _, g := range games {
		if g.Status == game.GameStatusActive {
			stats["active_games"] = stats["active_games"].(int) + 1
			continue
		}

		if g.Status == game.GameStatusFinished {
			switch g.Result {
			case game.GameResultWhiteWins:
				if g.WhitePlayer == playerID {
					stats["wins"] = stats["wins"].(int) + 1
				} else {
					stats["losses"] = stats["losses"].(int) + 1
				}
			case game.GameResultBlackWins:
				if g.BlackPlayer == playerID {
					stats["wins"] = stats["wins"].(int) + 1
				} else {
					stats["losses"] = stats["losses"].(int) + 1
				}
			case game.GameResultDraw:
				stats["draws"] = stats["draws"].(int) + 1
			}
		}
	}

	return stats, nil
}

// DeleteGame removes a game from the repository
func (s *gameService) DeleteGame(ctx context.Context, gameID primitive.ObjectID) error {
	if gameID.IsZero() {
		return errors.New("game ID is required")
	}

	return s.gameRepo.Delete(ctx, gameID)
}