package application

// BlockProgressDTO is block progress for API.
type BlockProgressDTO struct {
	BlockID      string  `json:"blockId"`
	SortOrder    int     `json:"sortOrder"`
	Title        string  `json:"title"`
	Status       string  `json:"status"`
	SubmittedAt  *string `json:"submittedAt,omitempty"`
	ApprovedAt   *string `json:"approvedAt,omitempty"`
	RejectedAt   *string `json:"rejectedAt,omitempty"`
	RejectReason *string `json:"rejectReason,omitempty"`
	Required     int     `json:"requiredMaterials"`
	Viewed       int     `json:"viewedMaterials"`
}

// BlockDetailDTO includes materials view flags.
type BlockDetailDTO struct {
	Block      BlockProgressDTO `json:"block"`
	Materials  []MaterialItemDTO `json:"materials"`
}

// MaterialItemDTO is material with viewed flag.
type MaterialItemDTO struct {
	MaterialID string  `json:"materialId"`
	Title      string  `json:"title"`
	Required   bool    `json:"required"`
	Viewed     bool    `json:"viewed"`
	FirstViewedAt *string `json:"firstViewedAt,omitempty"`
}

// MaterialViewResultDTO is response for record view.
type MaterialViewResultDTO struct {
	MaterialID string `json:"materialId"`
	BlockID    string `json:"blockId"`
	Created    bool   `json:"created"`
	Status     string `json:"status"`
}

// StudentProgressItemDTO is buddy list item with summary.
type StudentProgressItemDTO struct {
	StudentID        string `json:"studentId"`
	AwaitingCount    int    `json:"awaitingCount"`
	ApprovedCount    int    `json:"approvedCount"`
	InProgressCount  int    `json:"inProgressCount"`
}

// ApprovalQueueItemDTO is awaiting approval row for buddy.
type ApprovalQueueItemDTO struct {
	StudentID   string `json:"studentId"`
	BlockID     string `json:"blockId"`
	SubmittedAt string `json:"submittedAt"`
}
