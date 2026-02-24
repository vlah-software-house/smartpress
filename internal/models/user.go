// Package models defines the data structures that map to database tables
// and provides the core types used throughout the application.
package models

import (
	"time"

	"github.com/google/uuid"
)

// Role represents a user's permission level in the system.
type Role string

const (
	RoleAdmin  Role = "admin"
	RoleEditor Role = "editor"
	RoleAuthor Role = "author"
)

// User represents a CMS user with authentication and 2FA fields.
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never serialize the hash
	DisplayName  string    `json:"display_name"`
	Role         Role      `json:"role"`
	TOTPSecret   *string   `json:"-"` // Nullable; set during 2FA setup
	TOTPEnabled  bool      `json:"totp_enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// IsAdmin returns true if the user has the admin role.
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// Needs2FASetup returns true if the user has not completed 2FA enrollment.
// All users must set up 2FA on their first login.
func (u *User) Needs2FASetup() bool {
	return !u.TOTPEnabled
}
