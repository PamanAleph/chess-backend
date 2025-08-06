// Package game contains the Game domain entity and its business logic.
// This is part of the Domain layer in Hexagonal Architecture.
// Domain entities should be pure business logic without any external dependencies.
package game

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GameStatus represents the current status of a game
type GameStatus string

const (
	GameStatusWaiting   GameStatus = "waiting"   // Waiting for second player
	GameStatusActive    GameStatus = "active"    // Game in progress
	GameStatusFinished  GameStatus = "finished"  // Game completed
	GameStatusAbandoned GameStatus = "abandoned" // Game abandoned
)

// GameResult represents the result of a finished game
type GameResult string

const (
	GameResultWhiteWins GameResult = "white_wins"
	GameResultBlackWins GameResult = "black_wins"
	GameResultDraw      GameResult = "draw"
	GameResultAbandoned GameResult = "abandoned"
)

// Move represents a chess move
type Move struct {
	From      string    `bson:"from" json:"from"`           // e.g., "e2"
	To        string    `bson:"to" json:"to"`               // e.g., "e4"
	Piece     string    `bson:"piece" json:"piece"`         // e.g., "pawn", "king"
	Player    string    `bson:"player" json:"player"`       // "white" or "black"
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
	Notation  string    `bson:"notation" json:"notation"`   // Algebraic notation
}

// Game represents a chess game entity in the domain
type Game struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	WhitePlayer primitive.ObjectID `bson:"white_player" json:"white_player"`
	BlackPlayer primitive.ObjectID `bson:"black_player,omitempty" json:"black_player,omitempty"`
	Status      GameStatus         `bson:"status" json:"status"`
	Result      GameResult         `bson:"result,omitempty" json:"result,omitempty"`
	CurrentTurn string             `bson:"current_turn" json:"current_turn"` // "white" or "black"
	Moves       []Move             `bson:"moves" json:"moves"`
	Board       string             `bson:"board" json:"board"` // FEN notation
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
	FinishedAt  *time.Time         `bson:"finished_at,omitempty" json:"finished_at,omitempty"`
}

// NewGame creates a new chess game with white player
func NewGame(whitePlayerID primitive.ObjectID) (*Game, error) {
	if whitePlayerID.IsZero() {
		return nil, errors.New("white player ID cannot be empty")
	}

	now := time.Now()
	return &Game{
		ID:          primitive.NewObjectID(),
		WhitePlayer: whitePlayerID,
		Status:      GameStatusWaiting,
		CurrentTurn: "white",
		Moves:       []Move{},
		Board:       "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", // Starting FEN
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// JoinGame allows a second player to join the game
func (g *Game) JoinGame(blackPlayerID primitive.ObjectID) error {
	if blackPlayerID.IsZero() {
		return errors.New("black player ID cannot be empty")
	}
	if g.Status != GameStatusWaiting {
		return errors.New("game is not waiting for players")
	}
	if g.WhitePlayer == blackPlayerID {
		return errors.New("player cannot play against themselves")
	}

	g.BlackPlayer = blackPlayerID
	g.Status = GameStatusActive
	g.UpdatedAt = time.Now()
	return nil
}

// MakeMove adds a move to the game
func (g *Game) MakeMove(playerID primitive.ObjectID, from, to, piece, notation string) error {
	if g.Status != GameStatusActive {
		return errors.New("game is not active")
	}

	// Validate it's the player's turn
	var expectedPlayer string
	if g.CurrentTurn == "white" && g.WhitePlayer != playerID {
		return errors.New("it's not your turn")
	}
	if g.CurrentTurn == "black" && g.BlackPlayer != playerID {
		return errors.New("it's not your turn")
	}

	if g.CurrentTurn == "white" {
		expectedPlayer = "white"
	} else {
		expectedPlayer = "black"
	}

	// Create the move
	move := Move{
		From:      from,
		To:        to,
		Piece:     piece,
		Player:    expectedPlayer,
		Timestamp: time.Now(),
		Notation:  notation,
	}

	// Add move to game
	g.Moves = append(g.Moves, move)

	// Switch turns
	if g.CurrentTurn == "white" {
		g.CurrentTurn = "black"
	} else {
		g.CurrentTurn = "white"
	}

	g.UpdatedAt = time.Now()
	return nil
}

// ResignGame allows a player to resign
func (g *Game) ResignGame(playerID primitive.ObjectID) error {
	if g.Status != GameStatusActive {
		return errors.New("game is not active")
	}

	var result GameResult
	if g.WhitePlayer == playerID {
		result = GameResultBlackWins
	} else if g.BlackPlayer == playerID {
		result = GameResultWhiteWins
	} else {
		return errors.New("player is not part of this game")
	}

	g.Status = GameStatusFinished
	g.Result = result
	now := time.Now()
	g.FinishedAt = &now
	g.UpdatedAt = now
	return nil
}

// FinishGame marks the game as finished with a result
func (g *Game) FinishGame(result GameResult) error {
	if g.Status != GameStatusActive {
		return errors.New("game is not active")
	}

	g.Status = GameStatusFinished
	g.Result = result
	now := time.Now()
	g.FinishedAt = &now
	g.UpdatedAt = now
	return nil
}

// IsPlayerInGame checks if a player is part of this game
func (g *Game) IsPlayerInGame(playerID primitive.ObjectID) bool {
	return g.WhitePlayer == playerID || g.BlackPlayer == playerID
}

// GetPlayerColor returns the color of the player in this game
func (g *Game) GetPlayerColor(playerID primitive.ObjectID) (string, error) {
	if g.WhitePlayer == playerID {
		return "white", nil
	}
	if g.BlackPlayer == playerID {
		return "black", nil
	}
	return "", errors.New("player is not part of this game")
}

// IsValid checks if the game entity is valid
func (g *Game) IsValid() error {
	if g.WhitePlayer.IsZero() {
		return errors.New("white player cannot be empty")
	}
	if g.Status == GameStatusActive && g.BlackPlayer.IsZero() {
		return errors.New("active game must have black player")
	}
	if g.CurrentTurn != "white" && g.CurrentTurn != "black" {
		return errors.New("current turn must be white or black")
	}
	return nil
}