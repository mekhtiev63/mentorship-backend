package http

type scheduleBody struct {
	ScheduledAt string `json:"scheduled_at"`
}

type completeBody struct {
	Feedback string `json:"feedback"`
}

type failBody struct {
	Reason string `json:"reason"`
}
