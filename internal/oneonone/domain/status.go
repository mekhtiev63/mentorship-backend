package domain

// RequestStatus matches one_on_one_status enum.
type RequestStatus string

const (
	StatusPending   RequestStatus = "pending"
	StatusAccepted  RequestStatus = "accepted"
	StatusScheduled RequestStatus = "scheduled"
	StatusCompleted RequestStatus = "completed"
	StatusCancelled RequestStatus = "cancelled"
)
