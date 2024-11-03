package authentication

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/imharish-sivakumar/modern-oauth2-system/cisauth-proto/pb"

	"github.com/imharish-sivakumar/modern-oauth2-system/service-utils/constants"
	"github.com/imharish-sivakumar/modern-oauth2-system/service-utils/models"
)

const nullString = ""

type TokenMiddleware struct {
	client pb.TokenServiceClient
}

// Middleware abstraction for token introspection middleware.
type Middleware interface {
	DoAuthenticate(c *gin.Context)
}

// DoAuthenticate parses request authentication data and introspects the token and session.
func (t *TokenMiddleware) DoAuthenticate(c *gin.Context) {
	ctx := c.Request.Context()
	token := nullString
	sessionID := nullString
	authRequestHeader := c.GetHeader(constants.Authorization)
	cookies := c.GetHeader(constants.Cookie)
	switch {
	case len(authRequestHeader) != 0:
		slog.InfoContext(ctx, "using authentication mode from headers")
		authHeaders := strings.Split(authRequestHeader, " ")
		if len(authHeaders) == 2 && authHeaders[0] == constants.Bearer {
			token = authHeaders[1]
		}
		sessionID = c.GetHeader(constants.Session)
		if len(sessionID) == 0 {
			slog.InfoContext(ctx, "session id not present in the http headers")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "missing/invalid authentication headers",
			})
			return
		}
	case len(cookies) != 0:
		slog.InfoContext(ctx, "using authentication mode from cookies")
		cookies := strings.Split(cookies, "; ")
		for _, cookie := range cookies {
			compile := regexp.MustCompile(`\w+=(.+)`)
			submatch := compile.FindStringSubmatch(cookie)
			if strings.HasPrefix(cookie, "access_token") {
				token = submatch[1]
			}
			if strings.HasPrefix(cookie, "session") {
				sessionID = submatch[1]
			}
		}
	}

	if len(token) == 0 || len(sessionID) == 0 {
		slog.ErrorContext(ctx, "session id or access token not found")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "missing/invalid authentication headers",
		})
		return
	}

	// This is mandatory for all gRPC requests.
	// According to microsoft doc.
	// Every request should be triggered from the dependency request operationID as operationParentID.
	// We don't want an abstract layer neither gRPC client doesn't support middleware before it makes a request.
	// Below operationID used to keep the dependency until the response returned from the dependency service.
	// Refer docs at -
	// https://learn.microsoft.com/en-us/azure/azure-monitor/app/distributed-trace-data#data-model-for-telemetry-correlation
	introspect, err := t.client.Introspect(ctx, &pb.IntrospectRequest{
		AccessToken: token,
		SessionID:   sessionID,
	})

	if err != nil {
		slog.ErrorContext(ctx, "unable to introspect access token with provided session id", slog.Any("error", err))
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "missing/invalid authentication headers",
		})
		return
	}

	slog.InfoContext(ctx, "Successfully completed token introspection")

	if introspect.IsAccessTokenRefreshed {
		// these logs will help us find how frequent users are using the application and continuous engagement.
		slog.ErrorContext(ctx, "access token refreshed, setting new token in cookie")
		c.Writer.Header().Add("set-cookie", fmt.Sprintf("access_token=%s; Path=/; Max-Age=%d", introspect.NewAccessToken, introspect.NewAccessTokenExpiry))
	}

	if introspect.IDToken == nil {
		slog.ErrorContext(ctx, "id token is nil from introspection response")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "missing user profile",
		})
		return
	}

	slog.InfoContext(ctx, "setting user profile in context")

	parsedTokenID, err := uuid.Parse(introspect.IDToken.UserProfile.ID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "missing userID",
		})
		return
	}
	c.Set(constants.UserContext, models.UserProfile{
		ID:    &parsedTokenID,
		Email: introspect.IDToken.UserProfile.Email,
		Name:  introspect.IDToken.UserProfile.Name,
	})
	c.Next()
}

// NewTokenMiddleware returns token middleware instance with token client connection, ensure
// TOKEN_SERVICE_URL env variable is set before calling this function.
func NewTokenMiddleware(tokenServiceURL string, opts ...pb.TokenServiceClient) (*TokenMiddleware, error) {
	t := &TokenMiddleware{}

	if len(opts) == 0 {
		if len(tokenServiceURL) == 0 {
			return nil, errors.New("either token service url or token service client is required")
		}

		tokenServiceConnection, err := grpc.Dial(tokenServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Println("connection attempt to token service has failed : ", err)
			return nil, err
		}

		t.client = pb.NewTokenServiceClient(tokenServiceConnection)
	} else {
		t.client = opts[0]
	}

	return t, nil
}
