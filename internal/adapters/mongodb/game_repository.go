// Package mongodb provides MongoDB adapter implementations.
// This is part of the Adapters layer in Hexagonal Architecture.
package mongodb

import (
	"context"
	"errors"
	"time"

	"chess-backend/internal/domain/game"
	"chess-backend/internal/ports/repositories"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// gameRepository implements the GameRepository interface using MongoDB
type gameRepository struct {
	collection *mongo.Collection
}

// NewGameRepository creates a new instance of GameRepository
func NewGameRepository(collection *mongo.Collection) repositories.GameRepository {
	return &gameRepository{
		collection: collection,
	}
}

// Save creates or updates a game in the repository
func (r *gameRepository) Save(ctx context.Context, g *game.Game) error {
	if g == nil {
		return errors.New("game cannot be nil")
	}

	// Set timestamps
	now := time.Now()
	if g.CreatedAt.IsZero() {
		g.CreatedAt = now
	}
	g.UpdatedAt = now

	// If ID is empty, this is a new game
	if g.ID.IsZero() {
		g.ID = primitive.NewObjectID()
		result, err := r.collection.InsertOne(ctx, g)
		if err != nil {
			return err
		}
		g.ID = result.InsertedID.(primitive.ObjectID)
		return nil
	}

	// Update existing game
	filter := bson.M{"_id": g.ID}
	update := bson.M{"$set": g}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// FindByID retrieves a game by its ID
func (r *gameRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*game.Game, error) {
	var g game.Game
	filter := bson.M{"_id": id}
	err := r.collection.FindOne(ctx, filter).Decode(&g)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("game not found")
		}
		return nil, err
	}
	return &g, nil
}

// Update updates an existing game in the repository
func (r *gameRepository) Update(ctx context.Context, g *game.Game) error {
	if g == nil {
		return errors.New("game cannot be nil")
	}
	if g.ID.IsZero() {
		return errors.New("game ID cannot be empty")
	}

	g.UpdatedAt = time.Now()
	filter := bson.M{"_id": g.ID}
	update := bson.M{"$set": g}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// Delete removes a game from the repository
func (r *gameRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// FindByPlayerID retrieves all games for a specific player
func (r *gameRepository) FindByPlayerID(ctx context.Context, playerID primitive.ObjectID) ([]*game.Game, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"white_player": playerID},
			{"black_player": playerID},
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var games []*game.Game
	for cursor.Next(ctx) {
		var g game.Game
		if err := cursor.Decode(&g); err != nil {
			return nil, err
		}
		games = append(games, &g)
	}

	return games, cursor.Err()
}

// FindActiveGames retrieves all active games
func (r *gameRepository) FindActiveGames(ctx context.Context) ([]*game.Game, error) {
	return r.FindByStatus(ctx, game.GameStatusActive)
}

// FindWaitingGames retrieves all games waiting for players
func (r *gameRepository) FindWaitingGames(ctx context.Context) ([]*game.Game, error) {
	return r.FindByStatus(ctx, game.GameStatusWaiting)
}

// FindByStatus retrieves games by their status
func (r *gameRepository) FindByStatus(ctx context.Context, status game.GameStatus) ([]*game.Game, error) {
	filter := bson.M{"status": status}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var games []*game.Game
	for cursor.Next(ctx) {
		var g game.Game
		if err := cursor.Decode(&g); err != nil {
			return nil, err
		}
		games = append(games, &g)
	}

	return games, cursor.Err()
}

// FindByPlayers retrieves a game between two specific players
func (r *gameRepository) FindByPlayers(ctx context.Context, player1, player2 primitive.ObjectID) (*game.Game, error) {
	filter := bson.M{
		"$or": []bson.M{
			{
				"white_player": player1,
				"black_player": player2,
			},
			{
				"white_player": player2,
				"black_player": player1,
			},
		},
	}

	var g game.Game
	err := r.collection.FindOne(ctx, filter).Decode(&g)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &g, nil
}

// List retrieves games with pagination
func (r *gameRepository) List(ctx context.Context, offset, limit int) ([]*game.Game, error) {
	opts := options.Find()
	opts.SetSkip(int64(offset))
	opts.SetLimit(int64(limit))
	opts.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var games []*game.Game
	for cursor.Next(ctx) {
		var g game.Game
		if err := cursor.Decode(&g); err != nil {
			return nil, err
		}
		games = append(games, &g)
	}

	return games, cursor.Err()
}

// Count returns the total number of games
func (r *gameRepository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}

// CountByStatus returns the number of games with a specific status
func (r *gameRepository) CountByStatus(ctx context.Context, status game.GameStatus) (int64, error) {
	filter := bson.M{"status": status}
	return r.collection.CountDocuments(ctx, filter)
}

// CountByPlayer returns the number of games for a specific player
func (r *gameRepository) CountByPlayer(ctx context.Context, playerID primitive.ObjectID) (int64, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"white_player": playerID},
			{"black_player": playerID},
		},
	}
	return r.collection.CountDocuments(ctx, filter)
}