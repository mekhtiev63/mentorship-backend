package domain

import "errors"

var (
	ErrNotFound              = errors.New("roadmap block not found")
	ErrMaterialNotFound      = errors.New("material not found")
	ErrValidation            = errors.New("validation failed")
	ErrCannotPublish         = errors.New("cannot publish block")
	ErrHasProgress           = errors.New("block has student progress")
	ErrInvalidReorder        = errors.New("invalid reorder")
	ErrTitleRequired         = errors.New("title is required")
	ErrURLRequired           = errors.New("url is required")
	ErrInvalidMaterialType   = errors.New("invalid material type")
	ErrInvalidStatus         = errors.New("invalid block status")
	ErrInvalidSortOrder      = errors.New("sort order must be positive")
)
