package handlers

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"user-management-service/apperror"
	"user-management-service/config"
	umsConstants "user-management-service/constants"
	"user-management-service/domain"
	"user-management-service/model"

	"github.com/adjust/rmq/v5"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/imharish-sivakumar/modern-oauth2-system/cisauth-proto/pb"
	"github.com/imharish-sivakumar/modern-oauth2-system/service-utils/constants"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	kmsClient     *kms.Client
	tmsClient     pb.TokenServiceClient
	serviceConfig *config.ServiceConfig
	redisClient   *redis.Client
	emailQueue    rmq.Queue
	userService   domain.Service
}

func NewHandler(kmsClient *kms.Client,
	tmsClient pb.TokenServiceClient,
	serviceConfig *config.ServiceConfig,
	redisClient *redis.Client,
	userService domain.Service,
	emailQueue rmq.Queue) *Handler {
	return &Handler{
		kmsClient:     kmsClient,
		tmsClient:     tmsClient,
		serviceConfig: serviceConfig,
		redisClient:   redisClient,
		emailQueue:    emailQueue,
		userService:   userService,
	}
}

func (h *Handler) Register(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	_, err := h.userService.GetUserByEmail(ctx, user.Email)
	if err == nil || !errors.Is(err, sql.ErrNoRows) {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"status": "please try again",
		})
		return
	}

	decoded := [][]byte{}
	decryptables := []string{user.Password, user.ConfirmPassword}
	for _, encryptedText := range decryptables {
		decodeString, err := base64.StdEncoding.DecodeString(encryptedText)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}
		decoded = append(decoded, decodeString)
	}

	decrypted := [][]byte{}
	for _, decodedText := range decoded {
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
		decrypted = append(decrypted, output.Plaintext)
	}

	if string(decrypted[0]) != string(decrypted[1]) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "password and confirm password is not matching",
		})
		return
	}

	password, err := bcrypt.GenerateFromPassword(decrypted[0], bcrypt.DefaultCost)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "unable to generate hash for password",
		})
		return
	}

	user.Password = string(password)
	user.ConfirmPassword = string(password)
	userID := uuid.NewString()

	retryCount := 0
	count, err := h.redisClient.Get(ctx, fmt.Sprintf("%s:%s", user.Email, umsConstants.RegistrationEmailCount)).Result()
	if err != nil {
		slog.ErrorContext(ctx, "unable to fetch email retry count from redis", slog.Any(constants.Error, err))
		if !errors.Is(err, redis.Nil) {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Please try after sometime",
			})
			return
		}
		retryCount = 1
		goto incrementAndSend
	}

	retryCount, err = strconv.Atoi(count)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Please try after sometime",
		})
		return
	}

	if retryCount > h.serviceConfig.MaxVerificationRetryCount {
		slog.ErrorContext(ctx, "user reached max retry count")
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Please try after sometime",
		})
		return
	}

	goto incrementAndSend

incrementAndSend:
	err = h.redisClient.Set(ctx, fmt.Sprintf("%s:%s", user.Email, umsConstants.RegistrationEmailCount), retryCount+1, time.Minute*h.serviceConfig.VerificationLinkExpiry).Err()
	if err != nil {
		slog.ErrorContext(ctx, "unable to increment retry count", slog.Any(constants.Error, err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Please try after sometime",
		})
		return
	}

	userMarshalBytes, _ := json.Marshal(user)
	if err = h.redisClient.Set(ctx, userID, string(userMarshalBytes), time.Minute*h.serviceConfig.VerificationLinkExpiry).Err(); err != nil {
		slog.ErrorContext(ctx, "unable to set user in cache", slog.Any(constants.Error, err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Please try after sometime",
		})
		return
	}

	event := model.Event{
		Email: user.Email,
		Type:  model.VerificationEvent,
		EventPayload: []byte(fmt.Sprintf(`{
	"verificationID": "%s"
}`, userID)),
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
		return
	}

	if err = h.emailQueue.PublishBytes(eventBytes); err != nil {
		slog.ErrorContext(ctx, "unable to send new user message to SQS", slog.Any(constants.Error, err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

func (h *Handler) User(c *gin.Context) {
	userProfile := c.MustGet(constants.UserContext)

	c.JSON(http.StatusOK, userProfile)
}

func (handler *Handler) VerifyEmail(c *gin.Context) {
	var verificationRequest model.VerifyEmail
	if err := c.ShouldBindQuery(&verificationRequest); err != nil {
		validationError := apperror.CustomValidationError(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": validationError,
		})
		return
	}

	ctx := c.Request.Context()
	userDetails, err := handler.redisClient.Get(ctx, verificationRequest.Code).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "not found",
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	var user model.User
	if err := json.Unmarshal([]byte(userDetails), &user); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = handler.userService.CreateUser(ctx, user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	defer func() {
		handler.redisClient.Del(context.Background(), fmt.Sprintf("%s:%s", user.Email, umsConstants.RegistrationEmailCount), verificationRequest.Code)
	}()

	c.JSON(http.StatusCreated, gin.H{
		"status": "created",
	})
}
