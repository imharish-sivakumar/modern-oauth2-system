package secretsmanager

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

type SecretsManager struct {
	client *secretsmanager.Client
}

// NewSecretsManager initializes a new SecretsManager instance
func NewSecretsManager() (*SecretsManager, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}

	client := secretsmanager.NewFromConfig(cfg)

	return &SecretsManager{client: client}, nil
}

// GetSecret retrieves a secret value by its name
func (sm *SecretsManager) GetSecret(ctx context.Context, secretName string) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := sm.client.GetSecretValue(ctx, input)
	if err != nil {
		var notFound *types.ResourceNotFoundException
		if ok := errors.As(err, &notFound); ok {
			return "", fmt.Errorf("secret not found: %s", secretName)
		}
		return "", fmt.Errorf("failed to retrieve secret: %v", err)
	}

	if result.SecretString != nil {
		return *result.SecretString, nil
	}

	// If the secret is stored as binary, decode it
	decodedSecret, err := base64.StdEncoding.DecodeString(string(result.SecretBinary))
	if err != nil {
		return "", fmt.Errorf("failed to decode secret binary: %v", err)
	}

	return string(decodedSecret), nil
}
