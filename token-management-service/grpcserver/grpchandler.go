package grpcserver

import (
	"context"

	"github.com/google/uuid"

	"github.com/imharish-sivakumar/modern-oauth2-system/cisauth-proto/pb"
	"github.com/imharish-sivakumar/modern-oauth2-system/service-utils/models"

	"token-management-service/domain"
	"token-management-service/model"
)

// GRPCHandler implements pb.TokenServiceServer gRPC interface.
type GRPCHandler struct {
	pb.UnimplementedTokenServiceServer
	oauth2Service domain.Auth
}

// AcceptLogin accept login for user login challenge.
func (h *GRPCHandler) AcceptLogin(ctx context.Context, loginRequest *pb.AcceptLoginRequest) (*pb.AcceptLoginResponse, error) {
	userID := uuid.MustParse(loginRequest.UserProfile.ID)
	acceptLoginResponse, err := h.oauth2Service.Accept(ctx, loginRequest.LoginChallenge, models.UserProfile{
		ID:    &userID,
		Name:  loginRequest.UserProfile.Name,
		Email: loginRequest.UserProfile.Email,
	})
	if err != nil {
		return nil, err
	}

	return &pb.AcceptLoginResponse{RedirectTo: acceptLoginResponse.RedirectTo}, nil
}

// AcceptConsent initiate and accept login consent for user login challenge.
func (h *GRPCHandler) AcceptConsent(ctx context.Context, loginRequest *pb.AcceptConsentRequest) (*pb.AcceptConsentResponse, error) {
	acceptConsentResponse, err := h.oauth2Service.AcceptConsent(ctx, loginRequest.ConsentChallenge)
	if err != nil {
		return nil, err
	}

	return &pb.AcceptConsentResponse{RedirectTo: acceptConsentResponse.RedirectTo}, nil
}

// GenerateVerificationToken is grpc handler to generate token for verify/change credentials.
func (h *GRPCHandler) GenerateVerificationToken(ctx context.Context, request *pb.GenerateVerificationTokenRequest) (*pb.ClientTokenResponse, error) {
	clientToken, err := h.oauth2Service.AccessForClientToken(ctx, request.Email, request.ClientID)
	if err != nil {
		return nil, err
	}
	return &pb.ClientTokenResponse{
		AccessToken: clientToken.AccessToken,
		ExpiresIn:   clientToken.ExpiresIn,
		ExpiresAt:   clientToken.ExpiresAt,
		TokenType:   string(clientToken.TokenType),
		Scope:       clientToken.Scope,
		Email:       clientToken.Email,
	}, nil
}

// ExchangeToken exchanges a code for access token and refresh token.
func (h *GRPCHandler) ExchangeToken(ctx context.Context, tokenRequest *pb.TokenExchangeRequest) (*pb.TokenExchangeResponse, error) {
	tokenExchangeResponse, err := h.oauth2Service.ExchangeToken(ctx, model.TokenExchangeRequest{
		Code:         tokenRequest.Code,
		RedirectURI:  tokenRequest.RedirectURI,
		ClientID:     tokenRequest.ClientID,
		CodeVerifier: tokenRequest.CodeVerifier,
	})
	if err != nil {
		return nil, err
	}

	return &pb.TokenExchangeResponse{
		AccessToken:  tokenExchangeResponse.AccessToken,
		RefreshToken: tokenExchangeResponse.RefreshToken,
		IDToken:      tokenExchangeResponse.IDToken,
		ExpiresIn:    tokenExchangeResponse.ExpiresIn,
		ExpiresAt:    tokenExchangeResponse.ExpiresAt,
		SessionID:    tokenExchangeResponse.SessionID,
	}, nil
}

// IntrospectVerificationToken validated the access token is valid and active in redis as well as token itself.
func (h *GRPCHandler) IntrospectVerificationToken(ctx context.Context, tokenRequest *pb.IntrospectVerificationRequest) (*pb.IntrospectVerificationResponse, error) {
	introspectResponse, err := h.oauth2Service.IntrospectResponse(ctx, tokenRequest.AccessToken, model.AccessToken)
	if err != nil {
		return nil, err
	}
	return &pb.IntrospectVerificationResponse{
		Active:   introspectResponse.Active,
		Email:    introspectResponse.Email,
		ClientID: introspectResponse.ClientID,
	}, nil
}

// Introspect validates given access/refresh token is valid and active and refresh access token if access token is expired.
func (h *GRPCHandler) Introspect(ctx context.Context, tokenRequest *pb.IntrospectRequest) (*pb.IntrospectResponse, error) {
	introspectResponse, err := h.oauth2Service.IntrospectToken(ctx, tokenRequest.AccessToken, tokenRequest.SessionID, model.AccessToken)
	if err != nil {
		return nil, err
	}

	return &pb.IntrospectResponse{
		Active:                 introspectResponse.Active,
		Audience:               introspectResponse.Aud,
		ClientID:               introspectResponse.ClientId,
		Expiry:                 introspectResponse.Exp,
		IssuedAt:               introspectResponse.Iat,
		Issuer:                 introspectResponse.Iss,
		NotBefore:              introspectResponse.Nbf,
		ObfuscatedSubject:      introspectResponse.ObfuscatedSubject,
		Scope:                  introspectResponse.Scope,
		Subject:                introspectResponse.Sub,
		TokenType:              string(introspectResponse.TokenType),
		TokenUse:               introspectResponse.TokenUse,
		Username:               introspectResponse.Username,
		IsAccessTokenRefreshed: introspectResponse.IsAccessTokenRefreshed,
		NewAccessToken:         introspectResponse.NewAccessToken,
		NewAccessTokenExpiry:   introspectResponse.NewAccessTokenExpiry,
		IDToken: &pb.IDToken{
			UserProfile: &pb.UserProfile{
				ID:    introspectResponse.UserInfo.ID.String(),
				Name:  introspectResponse.UserInfo.Name,
				Email: introspectResponse.UserInfo.Email,
			},
			Subject:         introspectResponse.UserInfo.Subject,
			AccessTokenHash: introspectResponse.UserInfo.AccessTokenHash,
		},
	}, nil
}

// GenerateRefreshToken generates a new token with updated oauth2 hook claims.
func (h *GRPCHandler) GenerateRefreshToken(ctx context.Context, refreshTokenRequest *pb.GenerateRefreshTokenRequest) (*pb.TokenExchangeResponse, error) {
	tokenExchangeResponse, err := h.oauth2Service.FetchRefreshToken(ctx, refreshTokenRequest.GetAccessToken(), refreshTokenRequest.GetSessionID())
	if err != nil {
		return nil, err
	}

	return &pb.TokenExchangeResponse{
		AccessToken:  tokenExchangeResponse.AccessToken,
		RefreshToken: tokenExchangeResponse.RefreshToken,
		IDToken:      tokenExchangeResponse.IDToken,
		ExpiresIn:    tokenExchangeResponse.ExpiresIn,
		ExpiresAt:    tokenExchangeResponse.ExpiresAt,
		SessionID:    tokenExchangeResponse.SessionID,
	}, nil
}

// RevokeAccessToken revokes the access token.
func (h *GRPCHandler) RevokeAccessToken(ctx context.Context, revokeAccessTokenRequest *pb.RevokeAccessTokenRequest) (*pb.EmptyGrpcMessage, error) {
	if err := h.oauth2Service.RevokeAccessToken(ctx, revokeAccessTokenRequest.AccessToken, revokeAccessTokenRequest.SessionID, revokeAccessTokenRequest.ClientID); err != nil {
		return nil, err
	}
	return &pb.EmptyGrpcMessage{}, nil
}

// NewGRPCHandler creates an object of GRPCHandler.
func NewGRPCHandler(oAuth2 domain.Auth) *GRPCHandler {
	return &GRPCHandler{
		oauth2Service: oAuth2,
	}
}
