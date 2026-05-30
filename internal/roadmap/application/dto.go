package application

// BlockDTO is a roadmap block for API layers.
type BlockDTO struct {
	ID              string   `json:"id"`
	SortOrder       int      `json:"sortOrder"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	ExpectedSkills  []string `json:"expectedSkills"`
	Status          string   `json:"status"`
	IsActive        bool     `json:"isActive"`
	PublishedAt     *string  `json:"publishedAt,omitempty"`
	CreatedAt       string   `json:"createdAt"`
	UpdatedAt       string   `json:"updatedAt"`
}

// MaterialDTO is a material for API layers.
type MaterialDTO struct {
	ID           string `json:"id"`
	BlockID      string `json:"blockId"`
	SortOrder    int    `json:"sortOrder"`
	Title        string `json:"title"`
	MaterialType string `json:"materialType"`
	URL          string `json:"url"`
	Required     bool   `json:"required"`
	IsActive     bool   `json:"isActive"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

// BlockWithMaterialsDTO groups block and materials.
type BlockWithMaterialsDTO struct {
	Block     BlockDTO      `json:"block"`
	Materials []MaterialDTO `json:"materials"`
}

// RoadmapDTO is full student roadmap.
type RoadmapDTO struct {
	Blocks []BlockWithMaterialsDTO `json:"blocks"`
}
