package model

import (
	"time"

	"github.com/imharish-sivakumar/modern-oauth2-system/service-utils/models"
)

// TokenType is a enum type for oauth2 token types.
type TokenType string

const (
	// RefreshToken oauth2 refresh token workflow.
	RefreshToken = "refresh_token"
	// AccessToken oauth2 access token workflow.
	AccessToken = "access_token"
)

// AcceptLoginResponse model for accept login oauth2 response.
type AcceptLoginResponse struct {
	RedirectTo string `json:"redirect_to"`
}

// AcceptConsentResponse model for accept consent oauth2 response.
type AcceptConsentResponse struct {
	RedirectTo string `json:"redirect_to"`
}

// LoginAcceptRequest model for oauth2 login accept response.
// Acr sets the Authentication AuthorizationContext Class Reference value for this authentication session.
// We can use it to express that, for example, a user authenticated using two-factor authentication.
type LoginAcceptRequest struct {
	Subject     string             `json:"subject"`
	Remember    bool               `json:"remember"`
	RememberFor int                `json:"remember_for"`
	Acr         string             `json:"acr"`
	Userprofile models.UserProfile `json:"Context"`
}

// ConsentAcceptResponse model for oauth2 consent accept response.
type ConsentAcceptResponse struct {
	Userprofile                  models.UserProfile `json:"Context"`
	RequestedAccessTokenAudience []string           `json:"requested_access_token_audience"`
	RequestedScope               []string           `json:"requested_scope"`
}

// ConsentAcceptInitiateRequest model for oauth2 initiate consent response.
type ConsentAcceptInitiateRequest struct {
	GrantAccessTokenAudience []string `json:"grant_access_token_audience"`
	GrantScope               []string `json:"grant_scope"`
	Remember                 bool     `json:"remember"`
	RememberFor              int      `json:"remember_for"`
	Session                  Session  `json:"session"`
}

// Session model for session id token/user claim.
type Session struct {
	Userprofile models.UserProfile `json:"id_token"`
}

// TokenExchangeResponse model for response of code exchange for token with oauth2 server.
type TokenExchangeResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token,omitempty"`
	ExpiresIn    int64  `json:"expires_in"`
	ExpiresAt    string `json:"expires_at"`
	SessionID    string `json:"session_id"`
}

// ClientTokenResponse model for response of forgot credential token with oauth2 server .
type ClientTokenResponse struct {
	AccessToken string    `json:"access_token"`
	ExpiresIn   int64     `json:"expires_in"`
	ExpiresAt   string    `json:"expires_at"`
	Scope       string    `json:"scope"`
	TokenType   TokenType `json:"token_type"`
	Email       string    `json:"email"`
}

// TokenExchangeRequest model for request of code exchange for token with oauth2 server.
type TokenExchangeRequest struct {
	Code         string `json:"code" binding:"required"`
	RedirectURI  string `json:"redirect_uri" binding:"required"`
	ClientID     string `json:"client_id" binding:"required"`
	CodeVerifier string `json:"code_verifier" binding:"required"`
}

// IntrospectResponse is a model for oauth2 token introspection response.
// Aud Service-specific string identifier or list of string identifiers representingthe intended audience for this token, as defined in JWT.
// Exp Integer timestamp, measured in the number of seconds since January 1 1970 UTC, indicating when this token will expire, as defined in JWT.
// Ext is Extra is arbitrary data set by the session.
// Iat Integer timestamp, measured in the number of seconds since January 1 1970 UTC, indicating when this token was originally issued, as defined in JWT.
// Iss String representing the issuer of this token, as defined in JWT.
// Nbf Integer timestamp, measured in the number of seconds since January 1 1970 UTC, indicating when this token is not to be used before, as defined in JWT.
// ObfuscatedSubject packed subject variable(Sub).
// Sub Subject of the token, as defined in JWT. Usually a machine-readable identifier of the resource owner who authorized this token.
type IntrospectResponse struct {
	Active                 bool               `json:"active"`
	Aud                    []string           `json:"aud"`
	ClientId               string             `json:"client_id"`
	Exp                    int64              `json:"exp"`
	Ext                    models.UserProfile `json:"ext"`
	Iat                    int64              `json:"iat"`
	Iss                    string             `json:"iss"`
	Nbf                    int64              `json:"nbf"`
	ObfuscatedSubject      string             `json:"obfuscated_subject"`
	Scope                  string             `json:"scope"`
	Sub                    string             `json:"sub"`
	TokenType              TokenType          `json:"token_type"`
	TokenUse               string             `json:"token_use"`
	Username               string             `json:"username"`
	IsAccessTokenRefreshed bool               `json:"is_access_token_rotated"`
	NewAccessToken         string             `json:"new_access_token"`
	NewAccessTokenExpiry   int64              `json:"new_access_token_expiry"`
	UserInfo               IDToken            `json:"user_info"`
	Email                  string             `json:"email"`
}

// IntrospectVerificationResponse is a model for forgot password token introspection response.
type IntrospectVerificationResponse struct {
	Active   bool   `json:"active"`
	Email    string `json:"email"`
	ClientID string `json:"client_id"`
}

// IDToken is a model for oauth2 jwt id token.
type IDToken struct {
	models.UserProfile
	AccessTokenHash string `json:"at_hash"`
	Subject         string `json:"subject"`
}

// EmailRequestCount is a model for keeping request count of forgot credential in cache .
type EmailRequestCount struct {
	RequestCount int       `json:"request_count"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// RevokeAccessTokenErrorResponse is a model for revokeAccessToken response.
type RevokeAccessTokenErrorResponse struct {
	Error            string `json:"error"`
	ErrorDebug       string `json:"error_debug"`
	ErrorDescription string `json:"error_description"`
	ErrorHint        string `json:"error_hint"`
	StatusCode       int    `json:"status_code"`
}
