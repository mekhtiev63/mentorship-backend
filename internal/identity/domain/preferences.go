package domain

// UserPreferences stores session preferences for a user.
type UserPreferences struct {
	UserID     UserID
	ActiveRole *Role
}
