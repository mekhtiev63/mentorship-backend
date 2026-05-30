package domain

// TrackStatus matches final_check_status enum.
type TrackStatus string

const (
	StatusNotAvailable TrackStatus = "not_available"
	StatusAvailable    TrackStatus = "available"
	StatusScheduled    TrackStatus = "scheduled"
	StatusCompleted    TrackStatus = "completed"
	StatusFailed       TrackStatus = "failed"
)

func (s TrackStatus) isTerminal() bool {
	return s == StatusCompleted || s == StatusFailed
}
