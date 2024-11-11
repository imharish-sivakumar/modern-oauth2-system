package config

import (
	"encoding/json"
	"os"
)

type Secrets struct {
	RedisDBPassword    string `json:"REDIS_DB_PASSWORD"`
	RedisDBHost        string `json:"REDIS_DB_HOST"`
	RedisDBPort        string `json:"REDIS_DB_PORT"`
	PostgresDBName     string `json:"POSTGRES_DB_NAME"`
	PostgresDBUser     string `json:"POSTGRES_DB_USER"`
	PostgresDBPassword string `json:"POSTGRES_DB_PASSWORD"`
	PostgresDBPort     string `json:"POSTGRES_DB_PORT"`
	PostgresDBHost     string `json:"POSTGRES_DB_HOST"`
	SMTPHost           string `json:"SMTP_HOST"`
	SMTPPort           string `json:"SMTP_PORT"`
	SMTPUsername       string `json:"SMTP_USERNAME"`
	SMTPPassword       string `json:"SMTP_PASSWORD"`
}

type ServiceConfig struct {
	Name        string
	Environment string
	SecretKey   string
	FromEmail   string
}

func Load() (*ServiceConfig, error) {
	file, err := os.ReadFile("config/config.json")
	if err != nil {
		return nil, err
	}
	var serviceConfig ServiceConfig
	if err := json.Unmarshal(file, &serviceConfig); err != nil {
		return nil, err
	}

	return &serviceConfig, nil
}
