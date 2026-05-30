package http

type realWriteBody struct {
	Company             string  `json:"company"`
	Position            string  `json:"position"`
	ScheduledAt         string  `json:"scheduled_at"`
	StudentNotes        string  `json:"student_notes"`
	ExternalInterviewer *string `json:"external_interviewer"`
}

type completeBody struct {
	Outcome string `json:"outcome"`
}

type mockCreateBody struct {
	StudentID    string `json:"student_id"`
	ScheduledAt  string `json:"scheduled_at"`
	StudentNotes string `json:"student_notes"`
}

type feedbackBody struct {
	Feedback string `json:"feedback"`
	Outcome  string `json:"outcome"`
}
