package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/imharish-sivakumar/modern-oauth2-system/aws-utils/secretsmanager"
	"github.com/imharish-sivakumar/modern-oauth2-system/oauth2-clients-setup/config"
	utilsConstants "github.com/imharish-sivakumar/modern-oauth2-system/service-utils/constants"
	"github.com/imharish-sivakumar/modern-oauth2-system/service-utils/globalconfig"
	utilsLog "github.com/imharish-sivakumar/modern-oauth2-system/service-utils/log"
)

var (
	jobConfig    *config.JobConfig
	globalConfig *globalconfig.GlobalConfig
	secureClient *secretsmanager.SecretsManager
	ctx          context.Context
)

func init() {
	ctx = context.Background()
	var err error
	jobConfig, err = config.Load()
	if err != nil {
		os.Exit(1)
		return
	}

	utilsLog.InitializeLogger(jobConfig.Environment, jobConfig.Name)

	globalConfig, err = globalconfig.Load()
	if err != nil {
		os.Exit(1)
		return
	}

	secureClient, err = secretsmanager.NewSecretsManager()
	if err != nil {
		slog.ErrorContext(ctx, "unable to create secure client", slog.Any(utilsConstants.Error, err))
		return
	}
}

// ReplaceRedirectUris replaces redirect URIs if the condition is met
func replaceRedirectUris(client map[string]interface{}) map[string]interface{} {
	if clientConfig, ok := jobConfig.Clients[client["client_id"].(string)]; ok {
		client["redirect_uris"] = clientConfig.OAUTH2REDIRECTURL
	}
	return client
}

// ReplaceAllowedCorsOrigin replaces allowed CORS origin if the condition is met
func replaceAllowedCorsOrigin(client map[string]interface{}) map[string]interface{} {
	if clientConfig, ok := jobConfig.Clients[client["client_id"].(string)]; ok {
		client["allowed_cors_origin"] = clientConfig.ALLOWEDCORSORIGINS
	}
	return client
}

// MakeRequests makes the API requests
func makeRequests() error {
	// Read JSON file containing clients
	file, err := os.ReadFile("/config/clients.json")
	if err != nil {
		slog.ErrorContext(ctx, "Failed to read clients file", slog.Any(utilsConstants.Error, err))
		return err
	}

	var clientsFromFile []map[string]interface{}
	if err := json.Unmarshal(file, &clientsFromFile); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal clients file", slog.Any(utilsConstants.Error, err))
		return err
	}

	// Make GET request to fetch current clients
	resp, err := http.Get(fmt.Sprintf("%s/admin/clients", jobConfig.APIHost))
	if err != nil {
		slog.ErrorContext(ctx, "Failed to fetch current clients", slog.Any(utilsConstants.Error, err))
		return err
	}
	defer resp.Body.Close()

	var currentClients []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&currentClients); err != nil {
		slog.ErrorContext(ctx, "Failed to decode current clients", slog.Any(utilsConstants.Error, err))
		return err
	}

	existingClients := make(map[string]map[string]interface{})
	for _, client := range currentClients {
		existingClients[client["client_id"].(string)] = client
	}

	for _, client := range clientsFromFile {
		updatedClient := replaceRedirectUris(client)
		updatedClient = replaceAllowedCorsOrigin(updatedClient)

		if clientSecretEnvKey, ok := client["client_secret"].(string); ok {
			clientSecret, err := secureClient.GetSecret(ctx, clientSecretEnvKey)
			if err != nil {
				slog.ErrorContext(ctx, "unable to get client secret for client", slog.Any("clientID", client["client_id"]), slog.Any(utilsConstants.Error, err))
				return err
			}
			updatedClient["client_secret"] = clientSecret
		}

		clientJSON, err := json.Marshal(updatedClient)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to marshal updated client", slog.Any(utilsConstants.Error, err))
			return err
		}

		if existingClients[client["client_id"].(string)] != nil {
			// Update client if exists
			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/admin/clients/%s", jobConfig.APIHost, client["client_id"]), strings.NewReader(string(clientJSON)))
			if err != nil {
				slog.ErrorContext(ctx, "Failed to create PUT request", slog.Any(utilsConstants.Error, err))
				return err
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to update client", slog.Any(utilsConstants.Error, err))
				return err
			}
			defer resp.Body.Close()
			slog.InfoContext(ctx, "Updated client with id", slog.Any("clientID", client["client_id"]))
		} else {
			// Create client if not exists
			resp, err := http.Post(fmt.Sprintf("%s/admin/clients", jobConfig.APIHost), "application/json", strings.NewReader(string(clientJSON)))
			if err != nil {
				slog.ErrorContext(ctx, "Failed to create client", slog.Any(utilsConstants.Error, err))
				return err
			}
			defer resp.Body.Close()
			createClientResponseBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				slog.ErrorContext(ctx, "unable to read response body ", slog.Any(utilsConstants.Error, err))
				return err
			}
			slog.InfoContext(ctx, "create client response ", slog.String("response", string(createClientResponseBytes)))
			if resp.StatusCode == 201 {
				slog.InfoContext(ctx, "Created new client with id", slog.Any("clientID", client["client_id"]))
			} else {
				slog.ErrorContext(ctx, "non-200 status code returned from create clients", slog.Int("statusCode", resp.StatusCode))
			}
		}

		// Mark the client as processed
		delete(existingClients, client["client_id"].(string))
	}

	// Delete clients not present in file but in current API data
	for clientID := range existingClients {
		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/admin/clients/%s", jobConfig.APIHost, clientID), nil)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to create DELETE request", slog.Any(utilsConstants.Error, err))
			return err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to delete client", slog.Any(utilsConstants.Error, err))
			return err
		}
		defer resp.Body.Close()
		slog.InfoContext(ctx, "Deleted client with id", slog.Any("clientID", clientID))
	}

	slog.InfoContext(ctx, "Processing completed successfully.")
	return nil
}

func main() {
	if err := makeRequests(); err != nil {
		os.Exit(1)
		return
	}
	slog.InfoContext(ctx, "Successfully completed setting up clients")
}
