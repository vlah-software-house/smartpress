package models

import "testing"

// TestUserIsAdmin verifies that IsAdmin returns true only for the admin role.
func TestUserIsAdmin(t *testing.T) {
	tests := []struct {
		name string
		role Role
		want bool
	}{
		{name: "admin role", role: RoleAdmin, want: true},
		{name: "editor role", role: RoleEditor, want: false},
		{name: "author role", role: RoleAuthor, want: false},
		{name: "empty role", role: Role(""), want: false},
		{name: "unknown role", role: Role("superadmin"), want: false},
		{name: "uppercase ADMIN", role: Role("ADMIN"), want: false},
		{name: "mixed case Admin", role: Role("Admin"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{Role: tt.role}
			got := u.IsAdmin()
			if got != tt.want {
				t.Errorf("User{Role: %q}.IsAdmin() = %v, want %v", tt.role, got, tt.want)
			}
		})
	}
}

// TestUserNeeds2FASetup verifies 2FA setup detection based on
// TOTPEnabled and TOTPSecret fields.
func TestUserNeeds2FASetup(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"

	tests := []struct {
		name        string
		totpSecret  *string
		totpEnabled bool
		want        bool
	}{
		{
			name:        "no secret and not enabled",
			totpSecret:  nil,
			totpEnabled: false,
			want:        true,
		},
		{
			name:        "secret set but not enabled",
			totpSecret:  &secret,
			totpEnabled: false,
			want:        true,
		},
		{
			name:        "secret set and enabled",
			totpSecret:  &secret,
			totpEnabled: true,
			want:        false,
		},
		{
			name:        "nil secret but enabled (edge case)",
			totpSecret:  nil,
			totpEnabled: true,
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				TOTPSecret:  tt.totpSecret,
				TOTPEnabled: tt.totpEnabled,
			}
			got := u.Needs2FASetup()
			if got != tt.want {
				t.Errorf("Needs2FASetup() = %v, want %v (secret=%v, enabled=%v)",
					got, tt.want, tt.totpSecret != nil, tt.totpEnabled)
			}
		})
	}
}

// TestRoleConstants verifies that role string constants have the expected values.
func TestRoleConstants(t *testing.T) {
	tests := []struct {
		name string
		role Role
		want string
	}{
		{name: "admin", role: RoleAdmin, want: "admin"},
		{name: "editor", role: RoleEditor, want: "editor"},
		{name: "author", role: RoleAuthor, want: "author"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.role) != tt.want {
				t.Errorf("Role constant %s = %q, want %q", tt.name, string(tt.role), tt.want)
			}
		})
	}
}
