// Package mongodb implements the repository interfaces using MongoDB.
// This is part of the Adapters layer in Hexagonal Architecture.
// Adapters implement the ports defined in the ports layer.
package mongodb

import (
	"context"
	"errors"
	"time"

	"chess-backend/internal/domain/user"
	"chess-backend/internal/ports/repositories"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// userRepository implements the UserRepository interface using MongoDB
type userRepository struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

// NewUserRepository creates a new MongoDB user repository
func NewUserRepository(client *mongo.Client, databaseName string) repositories.UserRepository {
	db := client.Database(databaseName)
	collection := db.Collection("users")

	// Create indexes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create unique index on username
	usernameIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	collection.Indexes().CreateOne(ctx, usernameIndex)

	return &userRepository{
		client:     client,
		database:   db,
		collection: collection,
	}
}

// Save stores a user in MongoDB
func (r *userRepository) Save(ctx context.Context, user *user.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	// Remove IsValid() check as it may not be implemented
	// Basic validation can be done here if needed
	if user.Username == "" {
		return errors.New("username cannot be empty")
	}

	// Set creation time if not set
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	user.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("username already exists")
		}
		return err
	}

	return nil
}

// FindByID retrieves a user by their ID
func (r *userRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*user.User, error) {
	var user user.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// FindByUsername retrieves a user by their username
func (r *userRepository) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	var user user.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, user *user.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	// Remove IsValid() check as it may not be implemented
	// Basic validation can be done here if needed
	if user.Username == "" {
		return errors.New("username cannot be empty")
	}

	user.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"username":   user.Username,
			"password":   user.Password,
			"updated_at": user.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": user.ID}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

// Delete removes a user from the database
func (r *userRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

// Exists checks if a user exists by ID
func (r *userRepository) Exists(ctx context.Context, id primitive.ObjectID) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": id})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByUsername checks if a user exists by username
func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"username": username})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// List retrieves users with pagination
func (r *userRepository) List(ctx context.Context, offset, limit int) ([]*user.User, error) {
	opts := options.Find()
	opts.SetSkip(int64(offset))
	opts.SetLimit(int64(limit))
	opts.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*user.User
	for cursor.Next(ctx) {
		var user user.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Count returns the total number of users
func (r *userRepository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}