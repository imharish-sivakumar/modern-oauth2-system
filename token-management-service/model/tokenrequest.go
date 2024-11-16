package model

const (
	// RedisEmailCountKey is a unique key identifier for storing the forgot credential request count in cache .
	RedisEmailCountKey = "redisEmailCountKey"
)

// ForgotCredentialTokenRequest is JSON contract for token generation.
type ForgotCredentialTokenRequest struct {
	Email string `json:"email" binding:"required,email"`
}
