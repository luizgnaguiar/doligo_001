// Package dto (Data Transfer Object) contains the data structures used for transferring
// data between the API layer and the client. These objects are specifically designed
// for the external interface of the application and are separate from the internal
// domain models. This separation ensures that changes in the API contract do not
// directly impact the core business logic.
package dto

// LoginRequest represents the data structure for a user login request.
// It includes the necessary credentials for authentication.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the data structure for a successful login response.
// It provides the client with a JWT for subsequent authenticated requests.
type LoginResponse struct {
	Token string `json:"token"`
}
