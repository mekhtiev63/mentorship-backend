package domain

import "errors"

var (
	ErrNotFound           = errors.New("block progress not found")
	ErrMaterialNotFound   = errors.New("material not found")
	ErrBlockNotVisible    = errors.New("block not available")
	ErrInvalidTransition  = errors.New("invalid status transition")
	ErrRequiredViews      = errors.New("required materials not viewed")
	ErrSequentialBlock    = errors.New("previous block not approved")
	ErrForbidden          = errors.New("forbidden")
	ErrConflict           = errors.New("progress conflict")
	ErrRejectReason       = errors.New("reject reason is required")
	ErrInvalidRejectReason = errors.New("invalid reject reason")
)
