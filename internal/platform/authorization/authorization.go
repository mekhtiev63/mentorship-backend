package authorization

// Role represents a system role assigned to a user.
type Role string

const (
	RoleStudent Role = "student"
	RoleBuddy   Role = "buddy"
	RoleAdmin   Role = "admin"
)
