package models

import (
	"github.com/google/uuid"
)

// UserProfile is a user profile model for jwt claims.
type UserProfile struct {
	ID    *uuid.UUID `json:"id"`
	Email string     `json:"email"`
	Name  string     `json:"name"`
}
