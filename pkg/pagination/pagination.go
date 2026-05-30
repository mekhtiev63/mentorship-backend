package pagination

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Params holds normalized list query parameters.
type Params struct {
	Page     int
	PageSize int
	Offset   int
}

// Meta is returned with paginated collections.
type Meta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// Normalize applies defaults and caps for page and page size.
func Normalize(page, pageSize int) Params {
	if page < 1 {
		page = DefaultPage
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	offset := (page - 1) * pageSize
	return Params{
		Page:     page,
		PageSize: pageSize,
		Offset:   offset,
	}
}

// NewMeta builds pagination metadata for a response.
func NewMeta(page, pageSize int, total int64) Meta {
	totalPages := 0
	if pageSize > 0 && total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return Meta{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}
