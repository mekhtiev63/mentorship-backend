package domain

// Principal is the authenticated subject attached to a request context.
type Principal struct {
	UserID     UserID
	Roles      RoleSet
	ActiveRole *Role
}

// HasRole reports membership in the role set from the token.
func (p Principal) HasRole(role Role) bool {
	return p.Roles.Contains(role)
}
