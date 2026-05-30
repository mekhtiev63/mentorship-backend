package domain

// InterviewStatus matches interview_status enum.
type InterviewStatus string

const (
	StatusSubmitted InterviewStatus = "submitted"
	StatusReviewed  InterviewStatus = "reviewed"
	StatusScheduled InterviewStatus = "scheduled"
	StatusCompleted InterviewStatus = "completed"
	StatusCancelled InterviewStatus = "cancelled"
)

func statusValidForKind(kind InterviewKind, status InterviewStatus) bool {
	switch kind {
	case KindReal:
		switch status {
		case StatusSubmitted, StatusReviewed, StatusCompleted, StatusCancelled:
			return true
		}
	case KindMock:
		switch status {
		case StatusScheduled, StatusCompleted, StatusCancelled:
			return true
		}
	}
	return false
}
