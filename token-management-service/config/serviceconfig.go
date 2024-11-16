package config

import (
	"encoding/json"
	"log"
	"os"

	"token-management-service/apperror"

	"github.com/go-playground/validator/v10"
	"github.com/imharish-sivakumar/modern-oauth2-system/service-utils/constants"
)

// AppSecretKeys represents the secret keys to be determined and used in the app client secret.
type AppSecretKeys map[string]string

// AppConfig model to hold application specific configs.
type AppConfig struct {
	Name        string                `json:"name" validate:"required"`
	Environment constants.Environment `json:"environment" validate:"oneof=LOCAL DEV PROD"`
	CISAuth     App                   `json:"cisAuth" validate:"required"`
	SecretKey   string                `json:"secretKey"`
}

// App represents cisauth token app config.
type App struct {
	GRPCPort                 int                      `json:"grpcPort" validate:"required"`
	Clients                  map[string]Client        `json:"clients" validate:"required"`
	OAuthServerPublicBaseURL string                   `json:"oAuthServerPublicBaseURL" validate:"required,url"`
	OAuthServerAdminBaseURL  string                   `json:"oAuthServerAdminBaseURL" validate:"required,url"`
	SecretKeys               AppSecretKeys            `json:"secretKeys"`
	CredentialsResetSettings CredentialsResetSettings `json:"credentialsResetSettings"`
}
type Secrets struct {
	RedisDBPassword string `json:"REDIS_DB_PASSWORD"`
	RedisDBHost     string `json:"REDIS_DB_HOST"`
	RedisDBPort     string `json:"REDIS_DB_PORT"`
}

// Client represents oauth2 clients.
type Client struct {
	Secret      string `json:"secret"`
	RedirectURI string `json:"redirectURI"`
}

// CredentialsResetSettings represents reset config for forgot Credentials.
type CredentialsResetSettings struct {
	RequestCount int `json:"requestCount" validate:"len=5"`
	RequestTTL   int `json:"requestTTL"`
}

// Load parses json file to application config.
func Load() (*AppConfig, error) {
	file, err := os.ReadFile("config/serviceconfig.json")
	if err != nil {
		return nil, err
	}
	appConfig := AppConfig{}
	if err := json.Unmarshal(file, &appConfig); err != nil {
		return nil, err
	}

	validate := validator.New()
	if err := validate.Struct(appConfig); err != nil {
		validationError := apperror.CustomValidationError(err)
		log.Println(validationError)
		return nil, err
	}

	return &appConfig, nil
}
