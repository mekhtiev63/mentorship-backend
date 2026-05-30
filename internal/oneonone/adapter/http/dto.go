package http

import "encoding/json"

type createRequestBody struct {
	Message        string          `json:"message"`
	PreferredSlots json.RawMessage `json:"preferred_slots"`
}

type rejectRequestBody struct {
	Reason string `json:"reason"`
}
