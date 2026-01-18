package users

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

// UserResponse represents a user in API responses.
type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
