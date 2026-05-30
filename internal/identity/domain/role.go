package domain

// Role is an application role assigned to a user.
type Role string

const (
	RoleStudent Role = "student"
	RoleBuddy   Role = "buddy"
	RoleAdmin   Role = "admin"
)

// RoleSet is an immutable set of roles.
type RoleSet []Role

// Contains reports whether the set includes the role.
func (s RoleSet) Contains(role Role) bool {
	for _, r := range s {
		if r == role {
			return true
		}
	}
	return false
}

// Len returns the number of roles.
func (s RoleSet) Len() int {
	return len(s)
}

// Strings returns role names as plain strings.
func (s RoleSet) Strings() []string {
	out := make([]string, len(s))
	for i, r := range s {
		out[i] = string(r)
	}
	return out
}

// ParseRole parses a role string.
func ParseRole(raw string) (Role, error) {
	switch Role(raw) {
	case RoleStudent, RoleBuddy, RoleAdmin:
		return Role(raw), nil
	default:
		return "", ErrInvalidRole
	}
}
