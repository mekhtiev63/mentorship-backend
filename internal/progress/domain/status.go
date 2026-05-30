package domain

// ProgressStatus is block progress state.
type ProgressStatus string

const (
	StatusNotStarted        ProgressStatus = "not_started"
	StatusInProgress        ProgressStatus = "in_progress"
	StatusAwaitingApproval  ProgressStatus = "awaiting_approval"
	StatusApproved          ProgressStatus = "approved"
	StatusRejected          ProgressStatus = "rejected"
)

// BlockProgressKey is aggregate identity.
type BlockProgressKey struct {
	StudentID StudentID
	BlockID   BlockID
}
