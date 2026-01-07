package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type Role string

const (
	Admin Role = "admin"
	User  Role = "user"
	Guest Role = "guest"
)

type Permission string

const (
	PermAlarm  Permission = "alarm"
	PermDevice Permission = "device"
	PermGuest  Permission = "guest"
)

var rolePermissions = map[Role][]Permission{
	Admin: {PermAlarm, PermDevice, PermGuest},
	User:  {PermAlarm, PermDevice},
	Guest: {PermGuest},
}

func HasPermission(role Role, perm Permission) bool {
	perms := rolePermissions[role]
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}

// AuthContext represents authenticated request context
// FAZ L1: PIN-based authentication
type AuthContext struct {
	Role          Role
	Authenticated bool
	PIN           string // Never log or return in responses
}

// PINStore holds hashed PINs for each role
type PINStore struct {
	AdminPINHash string // SHA-256 hash
	UserPINHash  string // SHA-256 hash
}

// Default PIN store (hardcoded for FAZ L1)
var defaultPINStore = &PINStore{
	AdminPINHash: hashPIN("1234"), // Default admin PIN: 1234
	UserPINHash:  hashPIN("5678"), // Default user PIN: 5678
}

// hashPIN creates SHA-256 hash of PIN
func hashPIN(pin string) string {
	hash := sha256.Sum256([]byte(pin))
	return hex.EncodeToString(hash[:])
}

// ValidatePIN checks if PIN matches any role and returns auth context
// FAZ L1: Simple PIN-based validation
func ValidatePIN(pin string) (*AuthContext, error) {
	if pin == "" {
		// No PIN = guest/anonymous
		return &AuthContext{
			Role:          Guest,
			Authenticated: false,
			PIN:           "",
		}, nil
	}

	hashedPIN := hashPIN(pin)

	// Check admin PIN
	if hashedPIN == defaultPINStore.AdminPINHash {
		return &AuthContext{
			Role:          Admin,
			Authenticated: true,
			PIN:           pin, // Stored for context, NEVER logged
		}, nil
	}

	// Check user PIN
	if hashedPIN == defaultPINStore.UserPINHash {
		return &AuthContext{
			Role:          User,
			Authenticated: true,
			PIN:           pin,
		}, nil
	}

	// Invalid PIN
	return nil, fmt.Errorf("invalid PIN")
}

// IsAdmin checks if context has admin role
func (ac *AuthContext) IsAdmin() bool {
	return ac.Authenticated && ac.Role == Admin
}

// IsUser checks if context has user role
func (ac *AuthContext) IsUser() bool {
	return ac.Authenticated && ac.Role == User
}

// IsGuest checks if context is guest (unauthenticated)
func (ac *AuthContext) IsGuest() bool {
	return !ac.Authenticated || ac.Role == Guest
}
