package domain

// InterviewOutcome matches interview_outcome enum.
type InterviewOutcome string

const (
	OutcomeOffer     InterviewOutcome = "offer"
	OutcomeReject    InterviewOutcome = "reject"
	OutcomePending   InterviewOutcome = "pending"
	OutcomeNoResult  InterviewOutcome = "no_result"
)

// IsFinalRealOutcome reports whether outcome is valid when completing a real interview.
func IsFinalRealOutcome(o InterviewOutcome) bool {
	return o == OutcomeOffer || o == OutcomeReject || o == OutcomeNoResult
}
