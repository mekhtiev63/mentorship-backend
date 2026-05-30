package domain

import "time"

// Role matches app_role enum.
type Role string

const (
	RoleStudent Role = "student"
	RoleBuddy   Role = "buddy"
	RoleAdmin   Role = "admin"
)

// ParseRole parses role string.
func ParseRole(raw string) (Role, error) {
	switch Role(raw) {
	case RoleStudent, RoleBuddy, RoleAdmin:
		return Role(raw), nil
	default:
		return "", ErrInvalidRole
	}
}

// RoleAssignment links a user to a role.
type RoleAssignment struct {
	UserID    UserID
	Role      Role
	GrantedAt time.Time
}

// RoleSet is a list of roles.
type RoleSet []Role

// Contains reports role membership.
func (s RoleSet) Contains(role Role) bool {
	for _, r := range s {
		if r == role {
			return true
		}
	}
	return false
}

// Len returns number of roles.
func (s RoleSet) Len() int {
	return len(s)
}

// Strings returns role names.
func (s RoleSet) Strings() []string {
	out := make([]string, len(s))
	for i, r := range s {
		out[i] = string(r)
	}
	return out
}
