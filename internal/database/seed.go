package database

import (
	"database/sql"
	"fmt"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
)

// Seed populates the database with initial development data.
// It creates a default admin user if none exists. The admin will be
// prompted to set up 2FA on first login (totp_enabled = false).
func Seed(db *sql.DB) error {
	// Check if any users exist already.
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return fmt.Errorf("seed check users: %w", err)
	}

	if count > 0 {
		slog.Info("database already seeded, skipping")
		return nil
	}

	// Hash the default admin password.
	hash, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("seed bcrypt: %w", err)
	}

	// Insert default admin user. 2FA is not enabled — they must set it up
	// on first login.
	_, err = db.Exec(`
		INSERT INTO users (email, password_hash, display_name, role, totp_enabled)
		VALUES ($1, $2, $3, $4, $5)
	`, "admin@smartpress.local", string(hash), "Admin", "admin", false)
	if err != nil {
		return fmt.Errorf("seed insert admin: %w", err)
	}

	slog.Info("database seeded with default admin user",
		"email", "admin@smartpress.local",
		"password", "admin",
	)

	return nil
}
