package config

import (
	"encoding/json"
	"log"
	"os"

	"github.com/imharish-sivakumar/modern-oauth2-system/service-utils/constants"
)

type JobConfig struct {
	Name        string                  `json:"name"`
	Environment constants.Environment   `json:"environment"`
	APIHost     string                  `json:"APIHost"`
	Clients     map[string]ClientConfig `json:"clients"`
}

type ClientConfig struct {
	ALLOWEDCORSORIGINS []string `json:"ALLOWED_CORS_ORIGINS"`
	OAUTH2REDIRECTURL  []string `json:"OAUTH2_REDIRECT_URL"`
}

func Load() (*JobConfig, error) {
	file, err := os.ReadFile("/config/jobconfig.json")
	if err != nil {
		log.Println("unable to read job config ", err)
		return nil, err
	}
	var jobConfig JobConfig
	if err := json.Unmarshal(file, &jobConfig); err != nil {
		log.Println("unable to unmarshal job config", err)
		return nil, err
	}

	return &jobConfig, err
}
