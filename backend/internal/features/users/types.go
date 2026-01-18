package users

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

// UserResponse represents a user in API responses.
type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName *string   `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SetDisplayNameRequest represents a request to set the user's display name.
type SetDisplayNameRequest struct {
	DisplayName string `json:"display_name"`
}

// CheckDisplayNameResponse represents the response for checking display name availability.
type CheckDisplayNameResponse struct {
	Available bool   `json:"available"`
	Name      string `json:"name"`
}
