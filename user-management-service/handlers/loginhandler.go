package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"golang.org/x/crypto/bcrypt"

	"user-management-service/apperror"
	"user-management-service/model"

	"github.com/imharish-sivakumar/modern-oauth2-system/cisauth-proto/pb"
	"github.com/imharish-sivakumar/modern-oauth2-system/service-utils/constants"
)

// LoginWithPassword handles user login with email and password.
func (h *Handler) LoginWithPassword(c *gin.Context) {
	login := model.Login{}
	if err := c.ShouldBindJSON(&login); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": apperror.CustomValidationError(err),
		})
		return
	}
	ctx := c.Request.Context()

	decodedText, err := base64.StdEncoding.DecodeString(login.Password)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	output, err := h.kmsClient.Decrypt(ctx, &kms.DecryptInput{
		CiphertextBlob:      decodedText,
		EncryptionAlgorithm: types.EncryptionAlgorithmSpecRsaesOaepSha256,
		EncryptionContext:   map[string]string{},
		KeyId:               aws.String(h.serviceConfig.LoginPasswordKeyID),
	})

	if err != nil {
		slog.ErrorContext(ctx, "unable to decrypt password", slog.Any(constants.Error, err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := h.userService.GetUserByEmail(ctx, login.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), output.Plaintext); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": err.Error(),
		})
		return
	}

	acceptLogin, err := h.tmsClient.AcceptLogin(context.Background(), &pb.AcceptLoginRequest{
		LoginChallenge: login.LoginChallenge,
		UserProfile: &pb.UserProfile{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "unable to login",
		})
		return
	}

	c.JSON(http.StatusOK, model.AcceptLogin{RedirectTo: acceptLogin.RedirectTo})
}

// ConsentChallenge initiates and accept user login consent.
func (h *Handler) ConsentChallenge(c *gin.Context) {
	consentRequest := model.ConsentRequest{}
	if err := c.ShouldBindQuery(&consentRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": apperror.CustomValidationError(err),
		})
		return
	}

	consent, err := h.tmsClient.AcceptConsent(context.Background(), &pb.AcceptConsentRequest{ConsentChallenge: consentRequest.ConsentChallenge})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.AcceptLogin{RedirectTo: consent.RedirectTo})
}

// Exchange exchanges a code for access token.
func (h *Handler) Exchange(c *gin.Context) {
	request := model.TokenExchangeRequest{}
	if err := c.ShouldBindBodyWith(&request, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": apperror.CustomValidationError(err),
		})
		return
	}

	exchangeToken, err := h.tmsClient.ExchangeToken(context.TODO(), &pb.TokenExchangeRequest{
		Code:         request.Code,
		RedirectURI:  request.RedirectURI,
		ClientID:     request.ClientID,
		CodeVerifier: request.CodeVerifier,
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": err.Error(),
		})
		return
	}

	tokenExchangeResponse := &model.TokenExchangeResponse{
		AccessToken: exchangeToken.AccessToken,
		ExpiresIn:   exchangeToken.ExpiresIn,
		ExpiresAt:   exchangeToken.ExpiresAt,
		SessionID:   exchangeToken.SessionID,
	}

	clearCsrfCookies(c)
	// Dev purpose
	h.setAuthCookies(c, tokenExchangeResponse)

	c.JSON(http.StatusOK, tokenExchangeResponse)
}

func (h *Handler) setAuthCookies(c *gin.Context, response *model.TokenExchangeResponse) {
	c.Writer.Header().Add("set-cookie", fmt.Sprintf("access_token=%s; Path=/; Max-Age=%d", response.AccessToken, response.ExpiresIn))
	// Session should be alive till refresh token expires.
	expiresAt := time.Now().Add(time.Hour * time.Duration(h.serviceConfig.RefreshTokenExpiry))
	c.Writer.Header().Add("set-cookie", fmt.Sprintf("session=%s; Path=/; Max-Age=%d", response.SessionID, expiresAt.Unix()))
}

func clearCsrfCookies(c *gin.Context) {
	c.Writer.Header().Add("set-cookie", "oauth2_authentication_csrf=; Path=/; Max-Age=-1")
	c.Writer.Header().Add("set-cookie", "oauth2_authentication_session=; Path=/; Max-Age=-1")
	c.Writer.Header().Add("set-cookie", "oauth2_consent_csrf=; Path=/; Max-Age=-1")
}
