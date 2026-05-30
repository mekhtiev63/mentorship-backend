package response

import (
	"encoding/json"
	"net/http"
)

// Envelope is the standard success response shape for the REST API.
type Envelope struct {
	Data any `json:"data"`
	Meta any `json:"meta,omitempty"`
}

// ErrorBody is the standard error response shape.
type ErrorBody struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail describes a client-visible error.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// JSON writes a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// OK writes a 200 response with a data envelope.
func OK(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, Envelope{Data: data})
}

// Error writes an error response.
func Error(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, ErrorBody{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}
