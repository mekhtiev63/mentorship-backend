package domain

import "github.com/google/uuid"

// StudentID identifies a student user.
type StudentID string

// BlockID identifies a roadmap block.
type BlockID string

// MaterialID identifies a material.
type MaterialID string

// UserID identifies any user (approver).
type UserID string

// ParseStudentID parses student UUID.
func ParseStudentID(raw string) (StudentID, error) {
	if _, err := uuid.Parse(raw); err != nil {
		return "", ErrForbidden
	}
	return StudentID(raw), nil
}

// ParseBlockID parses block UUID.
func ParseBlockID(raw string) (BlockID, error) {
	if _, err := uuid.Parse(raw); err != nil {
		return "", ErrNotFound
	}
	return BlockID(raw), nil
}

// ParseMaterialID parses material UUID.
func ParseMaterialID(raw string) (MaterialID, error) {
	if _, err := uuid.Parse(raw); err != nil {
		return "", ErrMaterialNotFound
	}
	return MaterialID(raw), nil
}

// ParseUserID parses user UUID.
func ParseUserID(raw string) (UserID, error) {
	if _, err := uuid.Parse(raw); err != nil {
		return "", ErrForbidden
	}
	return UserID(raw), nil
}
