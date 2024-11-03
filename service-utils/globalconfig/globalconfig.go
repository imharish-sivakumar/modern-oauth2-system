package globalconfig

import (
	"embed"
	"encoding/json"
	"log"

	"github.com/go-playground/validator/v10"
)

//go:embed config/*
var content embed.FS

type KeyValuePair struct {
	Tag   string `json:"tag" validate:"required"`
	Value string `json:"value" validate:"required"`
}

// GlobalConfig model for serializing global config for microservices from kube config map.
type GlobalConfig struct {
	UserManagementServiceHost  KeyValuePair `json:"userManagementServiceHost" validate:"required"`
	TokenManagementServiceHost KeyValuePair `json:"tokenManagementServiceHost" validate:"required"`
}

// Load loads the json file into app config struct and process validation.
func Load() (*GlobalConfig, error) {
	file, err := content.ReadFile("config/globalconfig.json")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	globalConfig := GlobalConfig{}

	if err := json.Unmarshal(file, &globalConfig); err != nil {
		return nil, err
	}

	validate := validator.New()

	if err := validate.Struct(globalConfig); err != nil {
		log.Println(err)
		return nil, err
	}

	return &globalConfig, err
}
