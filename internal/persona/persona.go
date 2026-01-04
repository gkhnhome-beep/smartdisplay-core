// Package persona provides persona-based tone and wording for different user roles.
package persona

import "strings"

// PersonaType enumerates supported personas.
type PersonaType int

const (
	PersonaUnknown PersonaType = iota
	PersonaOwner
	PersonaAdult
	PersonaChild
	PersonaGuest
)

// PersonaFromRole maps an auth.Role string to a PersonaType.
func PersonaFromRole(role string) PersonaType {
	switch strings.ToLower(role) {
	case "owner", "admin":
		return PersonaOwner
	case "user", "adult":
		return PersonaAdult
	case "child":
		return PersonaChild
	case "guest":
		return PersonaGuest
	default:
		return PersonaUnknown
	}
}
