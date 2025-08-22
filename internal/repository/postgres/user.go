package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"go-fiber/internal/models"
	"go-fiber/internal/repository/interfaces"
	"go-fiber/internal/repository/postgres/queries"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

// userRepository implements the UserRepository interface for PostgreSQL
type userRepository struct {
	db      *pgxpool.Pool
	queries *queries.Queries
	logger  zerolog.Logger
}

// NewUserRepository creates a new PostgreSQL user repository
func NewUserRepository(db *pgxpool.Pool, logger zerolog.Logger) interfaces.UserRepository {
	return &userRepository{
		db:      db,
		queries: queries.New(db),
		logger:  logger,
	}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	var email, image pgtype.Text

	if user.Email != "" {
		email = pgtype.Text{String: user.Email, Valid: true}
	}
	if user.Image != "" {
		image = pgtype.Text{String: user.Image, Valid: true}
	}

	dbUser, err := r.queries.CreateUser(ctx, queries.CreateUserParams{
		Username:     user.Username,
		PasswordHash: user.Password,
		Email:        email,
		Image:        image,
	})
	if err != nil {
		r.logger.Error().Err(err).Str("username", user.Username).Msg("Failed to create user")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	result := &models.User{
		ID:        fmt.Sprintf("%v", dbUser.ID), // Convert interface{} to string
		Username:  dbUser.Username,
		Password:  dbUser.PasswordHash,
		CreatedAt: dbUser.CreatedAt.Time,
		UpdatedAt: dbUser.UpdatedAt.Time,
	}

	if dbUser.Email.Valid {
		result.Email = dbUser.Email.String
	}
	if dbUser.Image.Valid {
		result.Image = dbUser.Image.String
	}

	r.logger.Info().Str("user_id", result.ID).Str("username", result.Username).Msg("User created successfully")
	return result, nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	dbUser, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.Error().Err(err).Str("user_id", id).Msg("Failed to get user by ID")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	result := &models.User{
		ID:        fmt.Sprintf("%v", dbUser.ID), // Convert interface{} to string
		Username:  dbUser.Username,
		Password:  dbUser.PasswordHash,
		CreatedAt: dbUser.CreatedAt.Time,
		UpdatedAt: dbUser.UpdatedAt.Time,
	}

	if dbUser.Email.Valid {
		result.Email = dbUser.Email.String
	}
	if dbUser.Image.Valid {
		result.Image = dbUser.Image.String
	}

	return result, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	dbUser, err := r.queries.GetUserByEmail(ctx, pgtype.Text{String: email, Valid: true})
	if err != nil {
		r.logger.Error().Err(err).Str("email", email).Msg("Failed to get user by email")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	result := &models.User{
		ID:        fmt.Sprintf("%v", dbUser.ID), // Convert interface{} to string
		Username:  dbUser.Username,
		Password:  dbUser.PasswordHash,
		CreatedAt: dbUser.CreatedAt.Time,
		UpdatedAt: dbUser.UpdatedAt.Time,
	}

	if dbUser.Email.Valid {
		result.Email = dbUser.Email.String
	}
	if dbUser.Image.Valid {
		result.Image = dbUser.Image.String
	}

	return result, nil
}

// GetByUsername retrieves a user by username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	dbUser, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		r.logger.Error().Err(err).Str("username", username).Msg("Failed to get user by username")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	result := &models.User{
		ID:        fmt.Sprintf("%v", dbUser.ID), // Convert interface{} to string
		Username:  dbUser.Username,
		Password:  dbUser.PasswordHash,
		CreatedAt: dbUser.CreatedAt.Time,
		UpdatedAt: dbUser.UpdatedAt.Time,
	}

	if dbUser.Email.Valid {
		result.Email = dbUser.Email.String
	}
	if dbUser.Image.Valid {
		result.Image = dbUser.Image.String
	}

	return result, nil
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, user *models.User) (*models.User, error) {
	var email, image pgtype.Text

	if user.Email != "" {
		email = pgtype.Text{String: user.Email, Valid: true}
	}
	if user.Image != "" {
		image = pgtype.Text{String: user.Image, Valid: true}
	}

	dbUser, err := r.queries.UpdateUser(ctx, queries.UpdateUserParams{
		ID:       user.ID,
		Username: user.Username,
		Email:    email,
		Image:    image,
	})
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", user.ID).Msg("Failed to update user")
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	result := &models.User{
		ID:        fmt.Sprintf("%v", dbUser.ID), // Convert interface{} to string
		Username:  dbUser.Username,
		Password:  dbUser.PasswordHash,
		CreatedAt: dbUser.CreatedAt.Time,
		UpdatedAt: dbUser.UpdatedAt.Time,
	}

	if dbUser.Email.Valid {
		result.Email = dbUser.Email.String
	}
	if dbUser.Image.Valid {
		result.Image = dbUser.Image.String
	}

	r.logger.Info().Str("user_id", result.ID).Msg("User updated successfully")
	return result, nil
}

// Delete soft deletes a user
func (r *userRepository) Delete(ctx context.Context, id string) error {
	err := r.queries.SoftDeleteUser(ctx, id)
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", id).Msg("Failed to delete user")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	r.logger.Info().Str("user_id", id).Msg("User deleted successfully")
	return nil
}

// UpdateImage updates a user's image
func (r *userRepository) UpdateImage(ctx context.Context, id, imageURL string) error {
	var image pgtype.Text
	if imageURL != "" {
		image = pgtype.Text{String: imageURL, Valid: true}
	}

	_, err := r.queries.UpdateUserImage(ctx, queries.UpdateUserImageParams{
		ID:    id,
		Image: image,
	})
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", id).Msg("Failed to update user image")
		return fmt.Errorf("failed to update user image: %w", err)
	}

	r.logger.Info().Str("user_id", id).Msg("User image updated successfully")
	return nil
}

// UpdatePassword updates a user's password
func (r *userRepository) UpdatePassword(ctx context.Context, id, hashedPassword string) error {
	_, err := r.queries.UpdateUserPassword(ctx, queries.UpdateUserPasswordParams{
		ID:           id,
		PasswordHash: hashedPassword,
	})
	if err != nil {
		r.logger.Error().Err(err).Str("user_id", id).Msg("Failed to update user password")
		return fmt.Errorf("failed to update user password: %w", err)
	}

	r.logger.Info().Str("user_id", id).Msg("User password updated successfully")
	return nil
}

// List retrieves users with pagination
func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*models.User, int64, error) {
	// Get total count
	total, err := r.queries.CountUsers(ctx)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to count users")
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get users
	dbUsers, err := r.queries.ListUsers(ctx, queries.ListUsersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to list users")
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	users := make([]*models.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		user := &models.User{
			ID:        fmt.Sprintf("%v", dbUser.ID), // Convert interface{} to string
			Username:  dbUser.Username,
			Password:  dbUser.PasswordHash,
			CreatedAt: dbUser.CreatedAt.Time,
			UpdatedAt: dbUser.UpdatedAt.Time,
		}

		if dbUser.Email.Valid {
			user.Email = dbUser.Email.String
		}
		if dbUser.Image.Valid {
			user.Image = dbUser.Image.String
		}

		users[i] = user
	}

	return users, total, nil
}

// ExistsByEmail checks if a user exists by email
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if email == "" {
		return false, nil
	}

	exists, err := r.queries.ExistsByEmail(ctx, pgtype.Text{String: email, Valid: true})
	if err != nil {
		r.logger.Error().Err(err).Str("email", email).Msg("Failed to check if user exists by email")
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}

	return exists, nil
}

// ExistsByUsername checks if a user exists by username
func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	exists, err := r.queries.ExistsByUsername(ctx, username)
	if err != nil {
		r.logger.Error().Err(err).Str("username", username).Msg("Failed to check if user exists by username")
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}

	return exists, nil
}
