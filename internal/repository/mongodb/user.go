package mongodb

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"go-fiber/internal/models"
	"go-fiber/internal/repository/interfaces"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoUser represents a user document in MongoDB
type MongoUser struct {
	ID           string     `bson:"_id" json:"id"`
	Username     string     `bson:"username" json:"username"`
	PasswordHash string     `bson:"passwordHash" json:"-"`
	Email        string     `bson:"email,omitempty" json:"email,omitempty"`
	Image        string     `bson:"image,omitempty" json:"image,omitempty"`
	CreatedAt    time.Time  `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time  `bson:"updatedAt" json:"updatedAt"`
	DeletedAt    *time.Time `bson:"deletedAt,omitempty" json:"deletedAt,omitempty"`
}

// userRepository implements the UserRepository interface for MongoDB
type userRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// NewUserRepository creates a new MongoDB user repository
func NewUserRepository(db *mongo.Database, logger zerolog.Logger) interfaces.UserRepository {
	return &userRepository{
		collection: db.Collection("users"),
		logger:     logger,
	}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	// Generate ULID for new user
	entropy := ulid.Monotonic(rand.Reader, 0)
	id := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)

	now := time.Now()
	mongoUser := &MongoUser{
		ID:           id.String(),
		Username:     user.Username,
		PasswordHash: user.Password,
		Email:        user.Email,
		Image:        user.Image,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err := r.collection.InsertOne(ctx, mongoUser)
	if err != nil {
		r.logger.Error().Err(err).Str("username", user.Username).Msg("Failed to create user")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	result := r.mongoUserToModel(mongoUser)
	r.logger.Info().Str("user_id", result.ID).Str("username", result.Username).Msg("User created successfully")
	return result, nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	filter := bson.M{
		"_id":       id,
		"deletedAt": bson.M{"$exists": false},
	}

	var mongoUser MongoUser
	err := r.collection.FindOne(ctx, filter).Decode(&mongoUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error().Err(err).Str("user_id", id).Msg("Failed to get user by ID")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return r.mongoUserToModel(&mongoUser), nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	filter := bson.M{
		"email":     email,
		"deletedAt": bson.M{"$exists": false},
	}

	var mongoUser MongoUser
	err := r.collection.FindOne(ctx, filter).Decode(&mongoUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error().Err(err).Str("email", email).Msg("Failed to get user by email")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return r.mongoUserToModel(&mongoUser), nil
}

// GetByUsername retrieves a user by username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	filter := bson.M{
		"username":  username,
		"deletedAt": bson.M{"$exists": false},
	}

	var mongoUser MongoUser
	err := r.collection.FindOne(ctx, filter).Decode(&mongoUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error().Err(err).Str("username", username).Msg("Failed to get user by username")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return r.mongoUserToModel(&mongoUser), nil
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, user *models.User) (*models.User, error) {
	filter := bson.M{
		"_id":       user.ID,
		"deletedAt": bson.M{"$exists": false},
	}

	update := bson.M{
		"$set": bson.M{
			"username":  user.Username,
			"email":     user.Email,
			"image":     user.Image,
			"updatedAt": time.Now(),
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var mongoUser MongoUser
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&mongoUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error().Err(err).Str("user_id", user.ID).Msg("Failed to update user")
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	result := r.mongoUserToModel(&mongoUser)
	r.logger.Info().Str("user_id", result.ID).Msg("User updated successfully")
	return result, nil
}

// Delete soft deletes a user
func (r *userRepository) Delete(ctx context.Context, id string) error {
	filter := bson.M{
		"_id":       id,
		"deletedAt": bson.M{"$exists": false},
	}

	update := bson.M{
		"$set": bson.M{
			"deletedAt": time.Now(),
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", id).Msg("Failed to delete user")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}

	r.logger.Info().Str("user_id", id).Msg("User deleted successfully")
	return nil
}

// UpdateImage updates a user's image
func (r *userRepository) UpdateImage(ctx context.Context, id, imageURL string) error {
	filter := bson.M{
		"_id":       id,
		"deletedAt": bson.M{"$exists": false},
	}

	update := bson.M{
		"$set": bson.M{
			"image":     imageURL,
			"updatedAt": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", id).Msg("Failed to update user image")
		return fmt.Errorf("failed to update user image: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}

	r.logger.Info().Str("user_id", id).Msg("User image updated successfully")
	return nil
}

// UpdatePassword updates a user's password
func (r *userRepository) UpdatePassword(ctx context.Context, id, hashedPassword string) error {
	filter := bson.M{
		"_id":       id,
		"deletedAt": bson.M{"$exists": false},
	}

	update := bson.M{
		"$set": bson.M{
			"passwordHash": hashedPassword,
			"updatedAt":    time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", id).Msg("Failed to update user password")
		return fmt.Errorf("failed to update user password: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}

	r.logger.Info().Str("user_id", id).Msg("User password updated successfully")
	return nil
}

// List retrieves users with pagination
func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*models.User, int64, error) {
	filter := bson.M{"deletedAt": bson.M{"$exists": false}}

	// Get total count
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to count users")
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get users with pagination
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.M{"createdAt": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to list users")
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer cursor.Close(ctx)

	var mongoUsers []MongoUser
	if err := cursor.All(ctx, &mongoUsers); err != nil {
		r.logger.Error().Err(err).Msg("Failed to decode users")
		return nil, 0, fmt.Errorf("failed to decode users: %w", err)
	}

	users := make([]*models.User, len(mongoUsers))
	for i, mongoUser := range mongoUsers {
		users[i] = r.mongoUserToModel(&mongoUser)
	}

	return users, total, nil
}

// ExistsByEmail checks if a user exists by email
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if email == "" {
		return false, nil
	}

	filter := bson.M{
		"email":     email,
		"deletedAt": bson.M{"$exists": false},
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Str("email", email).Msg("Failed to check if user exists by email")
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}

	return count > 0, nil
}

// ExistsByUsername checks if a user exists by username
func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	filter := bson.M{
		"username":  username,
		"deletedAt": bson.M{"$exists": false},
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Error().Err(err).Str("username", username).Msg("Failed to check if user exists by username")
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}

	return count > 0, nil
}

// mongoUserToModel converts a MongoDB user document to a model user
func (r *userRepository) mongoUserToModel(mongoUser *MongoUser) *models.User {
	return &models.User{
		ID:        mongoUser.ID,
		Username:  mongoUser.Username,
		Password:  mongoUser.PasswordHash,
		Email:     mongoUser.Email,
		Image:     mongoUser.Image,
		CreatedAt: mongoUser.CreatedAt,
		UpdatedAt: mongoUser.UpdatedAt,
	}
}
