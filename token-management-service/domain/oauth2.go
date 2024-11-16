package domain

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	utilconstants "github.com/imharish-sivakumar/modern-oauth2-system/service-utils/constants"
	"github.com/imharish-sivakumar/modern-oauth2-system/service-utils/models"

	"token-management-service/config"
	"token-management-service/model"
)

var (
	// ErrInvalidLoginChallenge for login challenge issues.
	ErrInvalidLoginChallenge = errors.New("invalid login challenge")
	// ErrInvalidConsentChallenge for invalid consent challenge code.
	ErrInvalidConsentChallenge = errors.New("invalid consent challenge code")
	// ErrTokenExchangeBadRequest when OAuth2 server returns bad request for token exchange.
	ErrTokenExchangeBadRequest = errors.New("bad token exchange request")
	// ErrUnauthorisedTokenExchange when OAuth2 server returns unauthorised for token exchange.
	ErrUnauthorisedTokenExchange = errors.New("unauthorised token exchange request")
	// ErrTokenExchange when OAuth2 server returns internal server error for token exchange.
	ErrTokenExchange = errors.New("unable to exchange code for token")
	// ErrTokenExpired when access token is invalid or expired beyond refresh token validation.
	ErrTokenExpired = errors.New("token is expired")
	// ErrSessionExpired when both access token and refresh token are expired.
	ErrSessionExpired = errors.New("session expired")
	// ErrEmailLimitReached when email request limit for forgot credential is reached.
	ErrEmailLimitReached = errors.New("email request threshold limit reached")
	// ErrSessionNotFound when session is not found in redis.
	ErrSessionNotFound = errors.New("session could not be found")
	// ErrAccessTokenExpired when access token is expired or does not exist.
	ErrAccessTokenExpired = errors.New("either access token expired or does not exist")
	// ErrInvalidIDToken when token format is invalid.
	ErrInvalidIDToken = errors.New("idToken is invalid")
)

var (
	exchangeErrorMap = map[int]error{
		http.StatusBadRequest:          ErrTokenExchangeBadRequest,
		http.StatusUnauthorized:        ErrUnauthorisedTokenExchange,
		http.StatusInternalServerError: ErrTokenExchange,
	}
	introspectErrorMap = map[int]error{
		http.StatusUnauthorized: ErrTokenExpired,
	}
)

const (
	refreshTokenExpiry = 1 // in hours
	oauthKey           = "hydra.openid.id-token"

	// reused const's
	grantType   = "grant_type"
	clientID    = "client_id"
	redirectURI = "redirect_uri"
	contentType = "Content-Type"
	accept      = "Accept"
	token       = "token"
)

// Auth provides abstraction for OAuth2 authentication flow operations.
type Auth interface {
	Accept(ctx context.Context, loginChallenge string, UserProfile models.UserProfile) (*model.AcceptLoginResponse, error)
	AcceptConsent(ctx context.Context, consentChallenge string) (*model.AcceptConsentResponse, error)
	ExchangeToken(ctx context.Context, tokenExchangeRequest model.TokenExchangeRequest) (*model.TokenExchangeResponse, error)
	IntrospectToken(ctx context.Context, accessToken, sessionID string, tokenType model.TokenType) (*model.IntrospectResponse, error)
	IntrospectResponse(ctx context.Context, accessToken string, tokenType model.TokenType) (*model.IntrospectVerificationResponse, error)
	AccessForRefreshToken(ctx context.Context, refreshToken, clientID, existingSessionID string) (*model.TokenExchangeResponse, error)
	AccessForClientToken(ctx context.Context, email, clientID string) (*model.ClientTokenResponse, error)
	FetchRefreshToken(ctx context.Context, accessToken, sessionID string) (*model.TokenExchangeResponse, error)
	RevokeAccessToken(ctx context.Context, accessToken, sessionID, clientID string) error
}

// OAuth2 model for oauth2 dependencies.
type OAuth2 struct {
	httpClient  *http.Client
	redisClient *redis.Client
	appConfig   *config.App
}

// Accept calls OAuth2 admin login accept.
func (o *OAuth2) Accept(ctx context.Context, loginChallenge string, userProfile models.UserProfile) (*model.AcceptLoginResponse, error) {
	acceptLoginRequest := model.LoginAcceptRequest{
		Subject:     userProfile.ID.String(),
		Remember:    false,
		RememberFor: 0,
		Acr:         "1",
		Userprofile: userProfile,
	}

	queryParams := make(map[string]string)
	queryParams["login_challenge"] = loginChallenge

	payloadBytes, err := json.Marshal(acceptLoginRequest)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to marshal acceptLoginRequest", slog.Any(utilconstants.Error, err))
		return nil, err
	}

	slog.InfoContext(ctx, "Successfully marshaled acceptLoginRequest body")

	acceptLoginRequestEndpoint := o.appConfig.OAuthServerAdminBaseURL + "/oauth2/auth/requests/login/accept"
	request, err := http.NewRequestWithContext(ctx, http.MethodPut, acceptLoginRequestEndpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		slog.ErrorContext(ctx, "Unable to create accept login oauth2 request", slog.Any(utilconstants.Error, err))
		return nil, err
	}
	query := request.URL.Query()
	query.Add("login_challenge", loginChallenge)
	encodeQuery := query.Encode()
	request.URL.RawQuery = encodeQuery

	slog.InfoContext(ctx, "accept login request query parameter")

	response, err := o.httpClient.Do(request)
	if err != nil {
		slog.ErrorContext(ctx, "accept login request query parameter", slog.Any(utilconstants.Error, err), slog.String("query", encodeQuery))
		return nil, err
	}

	slog.InfoContext(ctx, "received accept login response")

	defer response.Body.Close()

	content, err := io.ReadAll(response.Body)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to read response body", slog.Any(utilconstants.Error, err))
		return nil, err
	}

	if response.StatusCode == http.StatusUnauthorized ||
		response.StatusCode == http.StatusNotFound ||
		response.StatusCode == http.StatusInternalServerError {
		slog.ErrorContext(ctx, "Unexpected status code returned for acceptLoginRequest")
		return nil, ErrInvalidLoginChallenge
	}

	var (
		acceptLoginResponse model.AcceptLoginResponse
	)

	if err := json.Unmarshal(content, &acceptLoginResponse); err != nil {
		slog.ErrorContext(ctx, "Unable to unmarshal acceptLoginResponse", slog.Any(utilconstants.Error, err))
		return nil, err
	}

	slog.InfoContext(ctx, "successfully returned acceptLoginResponse")

	return &acceptLoginResponse, nil
}

// AcceptConsent calls OAuth2 consent and accept consent.
func (o *OAuth2) AcceptConsent(ctx context.Context, consentChallenge string) (*model.AcceptConsentResponse, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.Join([]string{o.appConfig.OAuthServerAdminBaseURL, "oauth2/auth/requests/consent"}, "/"), nil)
	if err != nil {
		slog.ErrorContext(ctx, "Unable to create get login consent request", slog.Any(utilconstants.Error, err), slog.String("consentChallenge", consentChallenge))
		return nil, err
	}

	query := request.URL.Query()
	query.Add("consent_challenge", consentChallenge)
	encodedQuery := query.Encode()
	request.URL.RawQuery = encodedQuery

	slog.InfoContext(ctx, "constructed consent challenge query")

	response, err := o.httpClient.Do(request)
	if err != nil {
		slog.ErrorContext(ctx, "unable to make get consent request", slog.Any(utilconstants.Error, err))
		return nil, err
	}

	defer response.Body.Close()

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		slog.ErrorContext(ctx, "unable to parse get consent response", slog.Any(utilconstants.Error, err), slog.Int("statusCode", response.StatusCode))
		return nil, err
	}

	if response.StatusCode >= http.StatusMultipleChoices {
		slog.ErrorContext(ctx, "unexpected status code returned from oauth2 server for get consent request", slog.Int("statusCode", response.StatusCode))
		return nil, ErrInvalidConsentChallenge
	}

	var (
		consentAcceptResponse model.ConsentAcceptResponse
	)
	if err := json.Unmarshal(responseBytes, &consentAcceptResponse); err != nil {
		slog.ErrorContext(ctx, "unable to marshal get consent response", slog.Any(utilconstants.Error, err), slog.Int("statusCode", response.StatusCode), slog.String("consentResponse", string(responseBytes)))
		return nil, err
	}

	slog.InfoContext(ctx, "successfully parsed get consent response", slog.Any("consentResponse", consentAcceptResponse))

	consentAcceptInitiateRequest := model.ConsentAcceptInitiateRequest{
		GrantAccessTokenAudience: consentAcceptResponse.RequestedAccessTokenAudience,
		GrantScope:               consentAcceptResponse.RequestedScope,
		Remember:                 true,
		RememberFor:              1,
		Session: model.Session{
			Userprofile: consentAcceptResponse.Userprofile,
		},
	}

	consentBytes, err := json.Marshal(consentAcceptInitiateRequest)
	if err != nil {
		slog.ErrorContext(ctx, "unable to marshal accept consent request", slog.Any(utilconstants.Error, err))
		return nil, err
	}

	slog.InfoContext(ctx, "successfully created accept consent request body", slog.Any("acceptConsent", consentAcceptInitiateRequest))

	consentAcceptRequest, err := http.NewRequestWithContext(ctx, http.MethodPut, strings.Join([]string{o.appConfig.OAuthServerAdminBaseURL, "oauth2/auth/requests/consent/accept"}, "/"), bytes.NewBuffer(consentBytes))
	if err != nil {
		slog.ErrorContext(ctx, "unable to create accept consent request", slog.Any(utilconstants.Error, err))
		return nil, err
	}

	query = consentAcceptRequest.URL.Query()
	query.Add("consent_challenge", consentChallenge)
	acceptConsentEncodedQuery := query.Encode()
	consentAcceptRequest.URL.RawQuery = acceptConsentEncodedQuery

	slog.InfoContext(ctx, "formed accept consent query param", slog.String("acceptConsentEncodedQuery", acceptConsentEncodedQuery))

	slog.InfoContext(ctx, "making oauth2 consent accept request")
	consentAccept, err := o.httpClient.Do(consentAcceptRequest)
	if err != nil {
		slog.ErrorContext(ctx, "unable to make accept consent request", slog.Any(utilconstants.Error, err))
		return nil, err
	}

	if consentAccept.StatusCode >= http.StatusMultipleChoices {
		slog.ErrorContext(ctx, "received unexpected status code for accept consent request", slog.Int("statusCode", consentAccept.StatusCode))
		return nil, ErrInvalidConsentChallenge
	}

	defer consentAccept.Body.Close()

	consentAcceptBytes, err := io.ReadAll(consentAccept.Body)
	if err != nil {
		slog.ErrorContext(ctx, "unable to read accept consent response body", slog.Any(utilconstants.Error, err))
		return nil, err
	}

	slog.InfoContext(ctx, "successfully parsed accept consent response body", slog.String("acceptConsentResponse", string(consentAcceptBytes)))

	var (
		acceptConsentResponse model.AcceptConsentResponse
	)
	if err := json.Unmarshal(consentAcceptBytes, &acceptConsentResponse); err != nil {
		slog.ErrorContext(ctx, "unable to unmarshal accept consent response", slog.Any(utilconstants.Error, err), slog.String("acceptConsentResponse", string(consentAcceptBytes)))
		return nil, err
	}

	slog.InfoContext(ctx, "successfully returned accept consent response", slog.Any("result", acceptConsentResponse))

	return &acceptConsentResponse, nil
}

// ExchangeToken calls OAuth2 consent and accept consent.
func (o *OAuth2) ExchangeToken(ctx context.Context, tokenExchangeRequest model.TokenExchangeRequest) (*model.TokenExchangeResponse, error) {

	slog.InfoContext(ctx, "entered token exchange request", slog.Any("tokenExchangeRequest", tokenExchangeRequest))

	data := url.Values{
		grantType:       []string{"authorization_code"},
		"code_verifier": []string{tokenExchangeRequest.CodeVerifier},
		clientID:        []string{tokenExchangeRequest.ClientID},
		redirectURI:     []string{tokenExchangeRequest.RedirectURI},
		"code":          []string{tokenExchangeRequest.Code},
	}

	tokenResponse, err := o.exchangeToken(ctx, data)
	if err != nil {
		slog.ErrorContext(ctx, "unable to exchange token", slog.Any(utilconstants.Error, err))
		return nil, err
	}

	slog.InfoContext(ctx, "successfully exchanged token")

	// Suppressing marshal errors since marshaling errors are unlikely for manually constructed objects.
	tokenResponseBytes, _ := json.Marshal(tokenResponse)
	// RefreshToken expiry
	refreshTokenExpireAt := time.Hour * refreshTokenExpiry
	id := sessionId()
	tokenResponse.SessionID = id

	slog.InfoContext(ctx, "setting session id in redis")
	redisCmd := o.redisClient.Set(ctx, id, string(tokenResponseBytes), refreshTokenExpireAt)
	if _, err := redisCmd.Result(); err != nil {
		slog.ErrorContext(ctx, "unable to store session in redis cache", slog.Any(utilconstants.Error, err))
		return nil, err
	}

	slog.InfoContext(ctx, "successfully exchanged token!")

	return tokenResponse, nil
}

// IntrospectResponse is response for token introspect.
func (o *OAuth2) IntrospectResponse(ctx context.Context, accessToken string, tokenType model.TokenType) (*model.IntrospectVerificationResponse, error) {
	// check if it exists in the redis cache.
	cachedToken, err := o.getAccessTokenResponse(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	if cachedToken.AccessToken != accessToken {
		return nil, ErrAccessTokenExpired
	}
	headers := map[string][]string{
		contentType: {"application/x-www-form-urlencoded"},
		accept:      {"application/json"},
	}

	data := url.Values{
		token: []string{accessToken},
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.Join([]string{o.appConfig.OAuthServerAdminBaseURL, "oauth2/introspect"}, "/"), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	request.Header = headers

	slog.InfoContext(ctx, "making oauth2 introspect request")
	response, err := o.httpClient.Do(request)
	if err != nil {
		slog.ErrorContext(ctx, "making oauth2 introspect request")
		return nil, err
	}

	if response.StatusCode >= http.StatusBadRequest {
		slog.ErrorContext(ctx, "unexpected status code returned from consent accept response", slog.Any(utilconstants.Error, err), slog.Int("statusCode", response.StatusCode))
		return nil, introspectErrorMap[response.StatusCode]
	}

	all, err := io.ReadAll(response.Body)
	if err != nil {
		slog.ErrorContext(ctx, "unable to read response body for consent accept request", slog.Any(utilconstants.Error, err))
		return nil, err
	}
	introspectResponse := &model.IntrospectVerificationResponse{}
	if err := json.Unmarshal(all, introspectResponse); err != nil {
		slog.ErrorContext(ctx, "unable to unmarshal response body to introspect verification response", slog.Int("statusCode", response.StatusCode), slog.Any(utilconstants.Error, err))
		return nil, err
	}
	introspectResponse.Email = cachedToken.Email

	slog.InfoContext(ctx, "successfully completed consent accept request", slog.Int("statusCode", response.StatusCode))
	return introspectResponse, nil
}

// IntrospectToken introspect access token and validate user.
func (o *OAuth2) IntrospectToken(ctx context.Context, token, sessionID string, tokenType model.TokenType) (*model.IntrospectResponse, error) {
	tokenResponse, err := o.getTokenResponse(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if !(tokenResponse.AccessToken == token || tokenResponse.RefreshToken == token) {
		return nil, ErrSessionNotFound
	}

	headers := map[string][]string{
		contentType: {"application/x-www-form-urlencoded"},
		accept:      {"application/json"},
	}

	data := url.Values{
		"token": []string{token},
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.Join([]string{o.appConfig.OAuthServerAdminBaseURL, "oauth2/introspect"}, "/"), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	request.Header = headers

	slog.InfoContext(ctx, "making oauth2 introspection request")
	response, err := o.httpClient.Do(request)
	if err != nil {
		slog.ErrorContext(ctx, "unable to make oauth2 introspection request", slog.Any(utilconstants.Error, err))
		return nil, err
	}

	if response.StatusCode >= http.StatusBadRequest {
		slog.ErrorContext(ctx, "unable to make oauth2 introspection request", slog.Any(utilconstants.Error, err), slog.Int("statusCode", response.StatusCode))
		return nil, introspectErrorMap[response.StatusCode]
	}

	all, err := io.ReadAll(response.Body)
	if err != nil {
		slog.ErrorContext(ctx, "unable to read response body", slog.Any(utilconstants.Error, err), slog.Int("statusCode", response.StatusCode))
		return nil, err
	}
	introspectResponse := &model.IntrospectResponse{}
	if err := json.Unmarshal(all, introspectResponse); err != nil {
		slog.ErrorContext(ctx, "unable to unmarshal introspection response body", slog.Any(utilconstants.Error, err), slog.Int("statusCode", response.StatusCode))
		return nil, err
	}

	slog.InfoContext(ctx, "successfully introspected token")

	if !introspectResponse.Active {
		if tokenType == model.RefreshToken {
			slog.ErrorContext(ctx, "refresh token is expired")
			return nil, ErrSessionExpired
		}
		introspectToken, err := o.IntrospectToken(ctx, tokenResponse.RefreshToken, sessionID, model.RefreshToken)
		if err != nil {
			return nil, err
		}

		if introspectToken.Active {
			newToken, err := o.AccessForRefreshToken(ctx, tokenResponse.RefreshToken, introspectToken.ClientId, sessionID)
			if err != nil {
				return nil, err
			}
			introspectToken.TokenType = model.AccessToken
			introspectToken.IsAccessTokenRefreshed = true
			introspectToken.NewAccessToken = newToken.AccessToken
			expiry := time.Now().UTC().Add(time.Second * time.Duration(newToken.ExpiresIn)).Unix()
			introspectToken.NewAccessTokenExpiry = expiry
			return introspectToken, nil
		}
		return nil, ErrSessionExpired
	}
	userInfo, err := o.decodeIdTokenFromJWT(ctx, tokenResponse.IDToken)
	if err != nil {
		slog.ErrorContext(ctx, "unable to parse user information from jwt", slog.Any(utilconstants.Error, err))
		return nil, err
	}
	introspectResponse.UserInfo = userInfo

	return introspectResponse, nil
}

// AccessForClientToken creates a temp token for client.
func (o *OAuth2) AccessForClientToken(ctx context.Context, email, clientID string) (*model.ClientTokenResponse, error) {
	slog.InfoContext(ctx, "fetching email count keys from redis")
	// check whether the request already exists.
	result, err := o.redisClient.HGet(ctx, model.RedisEmailCountKey, email).Result()
	// if first time request.
	if err != nil {
		slog.ErrorContext(ctx, "unable to fetch email count keys from redis", slog.Any(utilconstants.Error, err))
		if errors.Is(err, redis.Nil) {
			ttl := time.Duration(o.appConfig.CredentialsResetSettings.RequestTTL) * time.Minute
			emailRequestCount := model.EmailRequestCount{
				RequestCount: 1,
				ExpiresAt:    time.Now().Add(ttl),
			}
			emailRequestCountBytes, err := json.Marshal(emailRequestCount)
			if err != nil {
				return nil, err
			}
			slog.InfoContext(ctx, "setting email count keys in redis")
			redisCmd := o.redisClient.HSet(ctx, model.RedisEmailCountKey, email, string(emailRequestCountBytes))
			if _, err := redisCmd.Result(); err != nil {
				slog.ErrorContext(ctx, "unable to set email count keys in redis", slog.Any(utilconstants.Error, err))
				return nil, err
			}
			slog.InfoContext(ctx, "successfully set email count keys in redis")
		}
		return o.clientCredentials(ctx, clientID, email)
	}
	emailRequestCount := model.EmailRequestCount{}
	if err := json.Unmarshal([]byte(result), &emailRequestCount); err != nil {
		slog.ErrorContext(ctx, "unable to unmarshal email count keys response from redis", slog.Any(utilconstants.Error, err))
		return nil, err
	}

	slog.InfoContext(ctx, "successfully fetched email count keys from redis")

	// if the request time expires reset count to 1.
	if emailRequestCount.ExpiresAt.Before(time.Now().UTC()) {
		emailRequestCount.RequestCount = 1
		ttl := time.Duration(o.appConfig.CredentialsResetSettings.RequestTTL) * time.Minute
		emailRequestCount.ExpiresAt = time.Now().Add(ttl)
		emailRequestCountBytes, err := json.Marshal(emailRequestCount)
		if err != nil {
			return nil, err
		}
		slog.InfoContext(ctx, "successfully updated the count of email notification in redis")
		redisCmd := o.redisClient.HSet(ctx, model.RedisEmailCountKey, email, string(emailRequestCountBytes))
		if _, err := redisCmd.Result(); err != nil {
			slog.ErrorContext(ctx, "unable to update email count keys notification in redis", slog.Any(utilconstants.Error, err))
			return nil, err
		}
		slog.InfoContext(ctx, "updated the email count keys in redis")
	} else if emailRequestCount.RequestCount < o.appConfig.CredentialsResetSettings.RequestCount {
		emailRequestCount.RequestCount++
	} else {
		return nil, status.Error(codes.PermissionDenied, ErrEmailLimitReached.Error())
	}
	emailRequestCountBytes, err := json.Marshal(emailRequestCount)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "setting the email notification count in redis", slog.Any(utilconstants.Error, err))
	redisCmd := o.redisClient.HSet(ctx, model.RedisEmailCountKey, email, string(emailRequestCountBytes))
	if _, err := redisCmd.Result(); err != nil {
		slog.ErrorContext(ctx, "unable to set email notification count in redis", slog.Any(utilconstants.Error, err))
		return nil, err
	}
	slog.InfoContext(ctx, "setting the email notification count in redis")
	return o.clientCredentials(ctx, clientID, email)
}

func (o *OAuth2) clientCredentials(ctx context.Context, clientID, email string) (*model.ClientTokenResponse, error) {
	data := url.Values{
		grantType: []string{"client_credentials"},
		"scope":   []string{"api"},
	}
	clientCredentials, err := o.generateToken(ctx, data, clientID)
	if err != nil {
		return nil, err
	}
	clientCredentials.Email = email

	// Suppressing marshal errors since marshaling errors are unlikely for manually constructed objects.
	clientCredentialsBytes, _ := json.Marshal(clientCredentials)
	clientTokenExpireAt := time.Duration(o.appConfig.CredentialsResetSettings.RequestTTL) * time.Minute
	slog.InfoContext(ctx, "setting client credentials token in redis against access token")
	redisGlobalCmd := o.redisClient.Set(ctx, clientCredentials.AccessToken, string(clientCredentialsBytes), clientTokenExpireAt)
	if _, err := redisGlobalCmd.Result(); err != nil {
		slog.ErrorContext(ctx, "unable to set client credentials token in redis")
		return nil, err
	}
	slog.InfoContext(ctx, "successfully set the client credentials token in redis")
	return clientCredentials, nil
}

// AccessForRefreshToken refresh token rotation.
func (o *OAuth2) AccessForRefreshToken(ctx context.Context, refreshToken, actualClientID, existingSessionID string) (*model.TokenExchangeResponse, error) {
	redirectURI := o.appConfig.Clients[actualClientID].RedirectURI
	data := url.Values{
		grantType:       []string{model.RefreshToken},
		clientID:        []string{actualClientID},
		redirectURI:     []string{redirectURI},
		"refresh_token": []string{refreshToken},
	}

	tokenResponse, err := o.exchangeToken(ctx, data)
	if err != nil {
		return nil, err
	}

	// Suppressing marshal errors since marshaling errors are unlikely for manually constructed objects.
	tokenResponseBytes, _ := json.Marshal(tokenResponse)
	// RefreshToken expiry
	refreshTokenExpireAt := time.Hour * refreshTokenExpiry
	tokenResponse.SessionID = existingSessionID

	slog.InfoContext(ctx, "setting token response against session id")
	redisCmd := o.redisClient.Set(ctx, existingSessionID, string(tokenResponseBytes), refreshTokenExpireAt)
	if _, err := redisCmd.Result(); err != nil {
		slog.ErrorContext(ctx, "unable to set session in redis", slog.Any(utilconstants.Error, err))
		return nil, err
	}
	slog.InfoContext(ctx, "successfully set token in session")

	return tokenResponse, nil
}

func (o *OAuth2) generateToken(ctx context.Context, data url.Values, clientID string) (*model.ClientTokenResponse, error) {
	username := clientID
	password := o.appConfig.Clients[clientID].Secret

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.Join([]string{o.appConfig.OAuthServerPublicBaseURL, "oauth2/token"}, "/"), strings.NewReader(data.Encode()))
	if err != nil {
		slog.ErrorContext(ctx, "unable to create new request to generate token for client id", slog.String("clientID", clientID), slog.Any(utilconstants.Error, err))
		return nil, err
	}

	request.SetBasicAuth(username, password)
	request.Header.Set(contentType, "application/x-www-form-urlencoded")

	slog.InfoContext(ctx, "making generate token request")
	response, err := o.httpClient.Do(request)
	if err != nil {
		slog.ErrorContext(ctx, "unable to make generate oauth2 token request for client id", slog.Any(utilconstants.Error, err), slog.String("clientID", clientID))
		return nil, err
	}

	bodyText, err := io.ReadAll(response.Body)
	if err != nil {
		slog.ErrorContext(ctx, "unable to parse response body bytes for oauth2 token request", slog.Any(utilconstants.Error, err), slog.String("clientID", clientID), slog.Int("statusCode", response.StatusCode))
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode > http.StatusMultipleChoices {
		slog.ErrorContext(ctx, "unexpected non 200 status code for oauth2 token request", slog.Int("statusCode", response.StatusCode), slog.String("clientID", clientID))
		return nil, exchangeErrorMap[response.StatusCode]
	}
	var clientTokenResponse model.ClientTokenResponse
	if err := json.Unmarshal(bodyText, &clientTokenResponse); err != nil {
		slog.InfoContext(ctx, "unable to unmarshal client token response for client id", slog.Any(utilconstants.Error, err), slog.Int("statusCode", response.StatusCode), slog.String("clientID", clientID))
		return nil, err
	}
	clientTokenResponse.ExpiresAt = time.Now().UTC().Add(time.Minute * time.Duration(o.appConfig.CredentialsResetSettings.RequestTTL)).
		Format(http.TimeFormat)

	slog.InfoContext(ctx, "successfully returned client token response for client id", slog.String("clientID", clientID), slog.String("expiresAt", clientTokenResponse.ExpiresAt), slog.Int("statusCode", response.StatusCode))

	return &clientTokenResponse, nil
}

func (o *OAuth2) exchangeToken(ctx context.Context, data url.Values) (*model.TokenExchangeResponse, error) {
	slog.InfoContext(ctx, "creating exchange token request")
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.Join([]string{o.appConfig.OAuthServerPublicBaseURL, "oauth2/token"}, "/"), strings.NewReader(data.Encode()))
	if err != nil {
		slog.ErrorContext(ctx, "unable to create exchange token request", slog.Any(utilconstants.Error, err), slog.Any("data", data))
		return nil, err
	}

	request.Header.Set(contentType, "application/x-www-form-urlencoded")

	actualClientID := data.Get(clientID)
	clientSecret := o.appConfig.Clients[actualClientID].Secret
	request.SetBasicAuth(actualClientID, clientSecret)

	slog.InfoContext(ctx, "making exchange token request")
	response, err := o.httpClient.Do(request)
	if err != nil {
		slog.ErrorContext(ctx, "unable to make exchange token request", slog.Any(utilconstants.Error, err))
		return nil, err
	}

	defer response.Body.Close()

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		slog.ErrorContext(ctx, "unable to read exchange toke response", slog.Any(utilconstants.Error, err), slog.Int("statusCode", response.StatusCode))
		return nil, err
	}

	if response.StatusCode > http.StatusMultipleChoices {
		slog.ErrorContext(ctx, "unexpected status code returned for exchange token request", slog.Int("statusCode", response.StatusCode), slog.Any("responseBody", string(responseBytes)))
		return nil, exchangeErrorMap[response.StatusCode]
	}

	var tokenExchangeResponse model.TokenExchangeResponse
	if err := json.Unmarshal(responseBytes, &tokenExchangeResponse); err != nil {
		slog.ErrorContext(ctx, "unable to unmarshal token exchange response", slog.Any(utilconstants.Error, err), slog.String("tokenExchangeResponse", string(responseBytes)))
		return nil, err
	}

	tokenExchangeResponse.ExpiresAt = time.Now().UTC().Add(time.Second * time.Duration(tokenExchangeResponse.ExpiresIn)).
		Format(http.TimeFormat)

	slog.InfoContext(ctx, "successfully returned token exchange response", slog.String("expiresAt", tokenExchangeResponse.ExpiresAt))

	return &tokenExchangeResponse, nil
}

func (o *OAuth2) getTokenResponse(ctx context.Context, sessionID string) (*model.TokenExchangeResponse, error) {
	slog.InfoContext(ctx, "fetching token for session")
	cmd := o.redisClient.Get(ctx, sessionID)
	result, err := cmd.Result()
	if err != nil {
		slog.ErrorContext(ctx, "unable to get the token response from redis cache using session id", slog.Any(utilconstants.Error, err), slog.String("sessionID", sessionID))
		return nil, err
	}

	slog.InfoContext(ctx, "successfully fetched token response from redis for sessionID", slog.String("sessionID", sessionID))

	tokenExchangeResponse := model.TokenExchangeResponse{}
	if err := json.Unmarshal([]byte(result), &tokenExchangeResponse); err != nil {
		slog.ErrorContext(ctx, "unable to unmarshal token exchange response", slog.Any(utilconstants.Error, err), slog.String("sessionID", sessionID))
		return nil, err
	}

	slog.InfoContext(ctx, "successfully returned token response from redis session", slog.String("sessionID", sessionID))

	return &tokenExchangeResponse, nil
}
func (o *OAuth2) getAccessTokenResponse(ctx context.Context, accessToken string) (*model.ClientTokenResponse, error) {
	slog.InfoContext(ctx, "fetching access token response")
	cmd := o.redisClient.Get(ctx, accessToken)
	result, err := cmd.Result()
	if err != nil {
		slog.ErrorContext(ctx, "unable to get access token response", slog.Any(utilconstants.Error, err))
		return nil, err
	}
	slog.InfoContext(ctx, "successfully fetched access token response from redis")
	clientTokenResponse := model.ClientTokenResponse{}
	if err := json.Unmarshal([]byte(result), &clientTokenResponse); err != nil {
		slog.ErrorContext(ctx, "unable to unmarshal access token response bytes")
		return nil, err
	}

	slog.InfoContext(ctx, "successfully returned access token response")

	return &clientTokenResponse, nil
}

// FetchRefreshToken creates a new access token based on accessToken and sessionID.
func (o *OAuth2) FetchRefreshToken(ctx context.Context, accessToken, sessionID string) (*model.TokenExchangeResponse, error) {
	token, err := o.IntrospectToken(ctx, accessToken, sessionID, model.AccessToken)
	if err != nil {
		slog.ErrorContext(ctx, "unable to introspect access token", slog.Any(utilconstants.Error, err), slog.String("sessionID", sessionID))
		return nil, err
	}

	slog.InfoContext(ctx, "successfully introspected access token with session id", slog.String("sessionID", sessionID))

	tokenResponse, err := o.getTokenResponse(ctx, sessionID)
	if err != nil {
		slog.ErrorContext(ctx, "unable to get token response from redis cache", slog.Any(utilconstants.Error, err), slog.String("sessionID", sessionID))
		return nil, err
	}

	slog.InfoContext(ctx, "successfully fetched token response from redis for sessionID", slog.String("sessionID", sessionID))

	newToken, err := o.AccessForRefreshToken(ctx, tokenResponse.RefreshToken, token.ClientId, sessionID)
	if err != nil {
		slog.ErrorContext(ctx, "unable to exchange refresh token for access token", slog.Any(utilconstants.Error, err), slog.String("sessionID", sessionID), slog.String("clientID", token.ClientId))
		return nil, err
	}

	slog.InfoContext(ctx, "successfully returned new token for access token", slog.String("sessionID", sessionID))

	return newToken, nil
}

// RevokeAccessToken revokes the access token.
func (o *OAuth2) RevokeAccessToken(ctx context.Context, accessToken, sessionID, clientID string) error {
	slog.InfoContext(ctx, "deleting token details from redis")
	redisCmd := o.redisClient.Del(ctx, sessionID)
	if _, err := redisCmd.Result(); err != nil {
		slog.ErrorContext(ctx, "unable to delete session in redis", slog.String("sessionID", sessionID), slog.String("clientID", clientID))
		return err
	}
	slog.InfoContext(ctx, "successfully deleted token from redis")

	username := clientID
	password := o.appConfig.Clients[clientID].Secret

	data := url.Values{
		token: []string{accessToken},
	}

	slog.InfoContext(ctx, "form data for revoke token request", slog.Any("data", data), slog.String("sessionID", sessionID), slog.String("clientID", clientID))

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.Join([]string{o.appConfig.OAuthServerPublicBaseURL, "oauth2/revoke"}, "/"), strings.NewReader(data.Encode()))
	if err != nil {
		slog.ErrorContext(ctx, "unable to create new request for oauth2 revoke token", slog.Any(utilconstants.Error, err))
		return err
	}

	request.SetBasicAuth(username, password)
	request.Header.Set(contentType, "application/x-www-form-urlencoded")

	slog.InfoContext(ctx, "making oauth2 revoke access token api call")
	response, err := o.httpClient.Do(request)
	if err != nil {
		slog.ErrorContext(ctx, "unable to make oauth2 revoke token request", slog.Any(utilconstants.Error, err), slog.String("sessionID", sessionID), slog.String("clientID", clientID))
		return err
	}
	if response.StatusCode == http.StatusOK {
		slog.InfoContext(ctx, "successfully revoked access token for session id", slog.String("sessionID", sessionID), slog.String("clientID", clientID))
		return nil
	}

	slog.ErrorContext(ctx, "status code is other than 200 returned from oauth2 during revoke token", slog.Any(utilconstants.Error, err), slog.Int("statusCode", response.StatusCode))

	if _, err := io.ReadAll(response.Body); err != nil {
		slog.ErrorContext(ctx, "unable to read response body for oauth2 revoke token", slog.Any(utilconstants.Error, err), slog.String("sessionID", sessionID), slog.String("clientID", clientID))
		return err
	}
	defer response.Body.Close()
	if response.StatusCode > http.StatusMultipleChoices {
		return exchangeErrorMap[response.StatusCode]
	}
	return errors.New("something went wrong")
}

func sessionId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// NewOAuth2 creates a new object for OAuth2.
func NewOAuth2(client *http.Client, redisClient *redis.Client, app *config.App) *OAuth2 {
	return &OAuth2{httpClient: client, redisClient: redisClient, appConfig: app}
}

func (o *OAuth2) decodeIdTokenFromJWT(ctx context.Context, idToken string) (model.IDToken, error) {
	// Parse the JWT token string without providing a key for signature verification
	token, _, err := new(jwt.Parser).ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		slog.ErrorContext(ctx, "unable to parse unverified id token claims from id token", slog.Any(utilconstants.Error, err))
		return model.IDToken{}, err
	}
	var idTokenClaims model.IDToken

	// Suppressing marshal errors since marshaling errors are unlikely for manually constructed objects.
	idTokenClaimsBytes, err := json.Marshal(token.Claims)
	if err != nil {
		slog.ErrorContext(ctx, "unable to marshal token claims", slog.Any(utilconstants.Error, err))
		return model.IDToken{}, err
	}
	if err := json.Unmarshal(idTokenClaimsBytes, &idTokenClaims); err != nil {
		return model.IDToken{}, err
	}
	val, ok := token.Header["kid"].(string)
	if !ok {
		slog.ErrorContext(ctx, "failed to fet kid from id token claims headers")
		return model.IDToken{}, errors.New("failed to get kid")
	}

	// check integrity of token data (redis) with oauth2 server
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.Join([]string{o.appConfig.OAuthServerAdminBaseURL, "keys", oauthKey, val}, "/"), nil)
	if err != nil {
		slog.ErrorContext(ctx, "unable to create id token validation request", slog.Any(utilconstants.Error, err))
		return model.IDToken{}, err
	}

	response, err := o.httpClient.Do(request)
	if err != nil {
		slog.ErrorContext(ctx, "unable to make id token validation request", slog.Any(utilconstants.Error, err))
		return model.IDToken{}, err
	}

	if response.StatusCode == http.StatusOK {
		slog.InfoContext(ctx, "successfully validated id token")
		return idTokenClaims, nil
	}

	slog.ErrorContext(ctx, "unexpected non 200 status code returned from id token validation", slog.Int("statusCode", response.StatusCode), slog.Any(utilconstants.Error, ErrInvalidIDToken))

	return model.IDToken{}, ErrInvalidIDToken
}
