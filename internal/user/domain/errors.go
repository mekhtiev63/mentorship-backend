package domain

import "errors"

var (
	ErrNotFound            = errors.New("user not found")
	ErrEmailTaken          = errors.New("email already taken")
	ErrInvalidEmail        = errors.New("invalid email")
	ErrInvalidRole         = errors.New("invalid role")
	ErrInvalidStatus       = errors.New("invalid status")
	ErrAssignmentNotFound  = errors.New("assignment not found")
	ErrInvalidAssignment   = errors.New("invalid assignment")
	ErrForbidden           = errors.New("forbidden")
	ErrBuddyRoleRequired   = errors.New("buddy role required")
	ErrStudentRoleRequired = errors.New("student role required")
	ErrWeakPassword        = errors.New("password too short")
)
