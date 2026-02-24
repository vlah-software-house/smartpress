// Package store provides database access methods for all SmartPress
// entities. Each store struct wraps a *sql.DB and exposes typed query methods.
package store

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"smartpress/internal/models"
)

// UserStore handles all user-related database operations.
type UserStore struct {
	db *sql.DB
}

// NewUserStore creates a new UserStore with the given database connection.
func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

// FindByEmail retrieves a user by their email address. Returns nil if not found.
func (s *UserStore) FindByEmail(email string) (*models.User, error) {
	u := &models.User{}
	err := s.db.QueryRow(`
		SELECT id, email, password_hash, display_name, role, totp_secret, totp_enabled, created_at, updated_at
		FROM users WHERE email = $1
	`, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role,
		&u.TOTPSecret, &u.TOTPEnabled, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return u, nil
}

// FindByID retrieves a user by their UUID. Returns nil if not found.
func (s *UserStore) FindByID(id uuid.UUID) (*models.User, error) {
	u := &models.User{}
	err := s.db.QueryRow(`
		SELECT id, email, password_hash, display_name, role, totp_secret, totp_enabled, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role,
		&u.TOTPSecret, &u.TOTPEnabled, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return u, nil
}

// List returns all users ordered by creation date.
func (s *UserStore) List() ([]models.User, error) {
	rows, err := s.db.Query(`
		SELECT id, email, password_hash, display_name, role, totp_secret, totp_enabled, created_at, updated_at
		FROM users ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(
			&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role,
			&u.TOTPSecret, &u.TOTPEnabled, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// Create inserts a new user with a bcrypt-hashed password.
func (s *UserStore) Create(email, password, displayName string, role models.Role) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u := &models.User{}
	err = s.db.QueryRow(`
		INSERT INTO users (email, password_hash, display_name, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, password_hash, display_name, role, totp_secret, totp_enabled, created_at, updated_at
	`, email, string(hash), displayName, role).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Role,
		&u.TOTPSecret, &u.TOTPEnabled, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// SetTOTPSecret saves the TOTP secret for a user (during 2FA setup).
func (s *UserStore) SetTOTPSecret(userID uuid.UUID, secret string) error {
	_, err := s.db.Exec(`
		UPDATE users SET totp_secret = $1, updated_at = NOW() WHERE id = $2
	`, secret, userID)
	if err != nil {
		return fmt.Errorf("set totp secret: %w", err)
	}
	return nil
}

// EnableTOTP marks 2FA as active for a user (after successful code verification).
func (s *UserStore) EnableTOTP(userID uuid.UUID) error {
	_, err := s.db.Exec(`
		UPDATE users SET totp_enabled = TRUE, updated_at = NOW() WHERE id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("enable totp: %w", err)
	}
	return nil
}

// ResetTOTP clears the TOTP secret and disables 2FA for a user.
// The user will be forced to set up 2FA again on their next login.
func (s *UserStore) ResetTOTP(userID uuid.UUID) error {
	_, err := s.db.Exec(`
		UPDATE users SET totp_secret = NULL, totp_enabled = FALSE, updated_at = NOW() WHERE id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("reset totp: %w", err)
	}
	return nil
}

// Delete removes a user by ID.
func (s *UserStore) Delete(userID uuid.UUID) error {
	_, err := s.db.Exec(`DELETE FROM users WHERE id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

// CheckPassword verifies a plaintext password against the user's stored hash.
func (s *UserStore) CheckPassword(user *models.User, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) == nil
}
