package models

import "time"

type User struct {
	ID              string     `json:"ID"`
	Email           string     `json:"email" binding:"required"`
	Password        string     `json:"password" binding:"required"`
	ConfirmPassword string     `json:"confirmPassword" binding:"required"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       *time.Time `json:"updatedAt"`
	DeletedAt       *time.Time `json:"deletedAt"`
}

type EventType string

const (
	VerificationEvent EventType = "VerificationEvent"
)

type Event struct {
	Email        string
	Type         EventType
	EventPayload []byte
}
