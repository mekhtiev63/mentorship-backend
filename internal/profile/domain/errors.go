package domain

import "errors"

var (
	ErrNotFound           = errors.New("profile not found")
	ErrInvalidVisibility  = errors.New("invalid visibility")
	ErrInvalidTelegram    = errors.New("invalid telegram username")
	ErrTelegramTaken      = errors.New("telegram username taken")
	ErrForbidden          = errors.New("forbidden")
)
