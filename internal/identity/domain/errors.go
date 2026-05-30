package domain

import "errors"

var (
	ErrInvalidUserID      = errors.New("invalid user id")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidRole        = errors.New("invalid role")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountInactive    = errors.New("account inactive")
	ErrInvalidToken       = errors.New("invalid token")
	ErrActiveRoleNotAllowed = errors.New("active role not allowed")
)
