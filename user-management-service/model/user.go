package model

import "time"

type User struct {
	ID              string     `json:"ID"`
	Email           string     `json:"email" binding:"required"`
	Name            string     `json:"name"`
	Password        string     `json:"password" binding:"required"`
	ConfirmPassword string     `json:"confirmPassword" binding:"required"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       *time.Time `json:"updatedAt"`
	DeletedAt       *time.Time `json:"deletedAt"`
}

type VerifyEmail struct {
	Code string `form:"code" binding:"uuid,required"`
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

// Login is a user login request model with email and password.
type Login struct {
	Email          string `json:"email" binding:"required,email,min=5,max=50"`
	Password       string `json:"password" binding:"required"`
	LoginChallenge string `json:"loginChallenge" binding:"required,loginChallenge"`
}

// ConsentRequest is a user login consent request model.
type ConsentRequest struct {
	ConsentChallenge string `form:"consent_challenge" binding:"required"`
}

// AcceptLogin is a user login accept request model.
type AcceptLogin struct {
	RedirectTo string `json:"redirect_to"`
}

// TokenExchangeRequest is a token exchange request model.
type TokenExchangeRequest struct {
	Code         string `json:"code" binding:"required"`
	RedirectURI  string `json:"redirectURI" binding:"required"`
	ClientID     string `json:"clientID" binding:"required"`
	CodeVerifier string `json:"codeVerifier" binding:"required"`
}

// TokenExchangeResponse is a token exchange response model.
type TokenExchangeResponse struct {
	AccessToken  string `json:"accessToken,omitempty"`
	IDToken      string `json:"-"`
	RefreshToken string `json:"-,omitempty"`
	ExpiresIn    int64  `json:"expiresIn,omitempty"`
	ExpiresAt    string `json:"expiresAt,omitempty"`
	SessionID    string `json:"sessionID"`
}
