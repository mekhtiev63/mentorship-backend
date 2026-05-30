package http

type eventWriteBody struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	StartsAt    string   `json:"starts_at"`
	EndsAt      string   `json:"ends_at"`
	RelatedType string   `json:"related_type"`
	RelatedID   *string  `json:"related_id"`
	AttendeeIDs []string `json:"attendee_ids"`
}

type eventUpdateBody struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	StartsAt    string   `json:"starts_at"`
	EndsAt      string   `json:"ends_at"`
	AttendeeIDs []string `json:"attendee_ids"`
}
