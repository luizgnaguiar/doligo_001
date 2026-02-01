package dto

// ErrorResponse represents a generic error response structure for the API.
type ErrorResponse struct {
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}
