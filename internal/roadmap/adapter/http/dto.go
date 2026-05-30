package http

// createBlockRequest is JSON for block creation.
type createBlockRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// updateBlockRequest is JSON for block update.
type updateBlockRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      *string `json:"status,omitempty"`
}

// setActiveRequest toggles active flag.
type setActiveRequest struct {
	Active bool `json:"active"`
}

// reorderBlocksRequest is JSON for block reorder.
type reorderBlocksRequest struct {
	Items []reorderBlockItem `json:"items"`
}

type reorderBlockItem struct {
	ID        string `json:"id"`
	SortOrder int    `json:"sortOrder"`
}

// createMaterialRequest is JSON for material creation.
type createMaterialRequest struct {
	Title        string `json:"title"`
	MaterialType string `json:"materialType"`
	URL          string `json:"url"`
	Required     bool   `json:"required"`
}

// updateMaterialRequest is JSON for material update.
type updateMaterialRequest struct {
	Title        string `json:"title"`
	MaterialType string `json:"materialType"`
	URL          string `json:"url"`
	Required     bool   `json:"required"`
}

// reorderMaterialsRequest is JSON for material reorder.
type reorderMaterialsRequest struct {
	Items []reorderMaterialItem `json:"items"`
}

type reorderMaterialItem struct {
	ID        string `json:"id"`
	SortOrder int    `json:"sortOrder"`
}
