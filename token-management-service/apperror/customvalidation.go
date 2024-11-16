package apperror

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

const isRequired = "is required"

var (
	customErrors = map[string]error{
		// serviceConfig related errors.
		"grpcPort.required":                      errors.New(isRequired),
		"clients.required":                       errors.New(isRequired),
		"environment.oneof":                      errors.New("environment should be one of local, dev or prod"),
		"oAuthServerPublicBaseURL.required":      errors.New(isRequired),
		"oAuthServerAdminBaseURL.required":       errors.New(isRequired),
		"oAuthServerPublicBaseURL.url":           errors.New("oauth2 url is invalid"),
		"oAuthServerAdminBaseURL.url":            errors.New("oauth2 url is invalid"),
		"redisHost.required":                     errors.New(isRequired),
		"redisPort.required":                     errors.New(isRequired),
		"appinsightsInstrumentationKey.required": errors.New(isRequired),
	}
)

// CustomValidationError converts validation and json marshal error into custom error type.
func CustomValidationError(err error) []map[string]string {
	errs := make([]map[string]string, 0)
	switch e := err.(type) {
	case validator.ValidationErrors:
		for _, e := range err.(validator.ValidationErrors) {
			errorMap := make(map[string]string)

			key := e.Field() + "." + e.Tag()

			if v, ok := customErrors[key]; ok {
				errorMap[e.Field()] = v.Error()
			} else {
				errorMap[e.Field()] = fmt.Sprintf("custom message is not available: %v", err)
			}
			errs = append(errs, errorMap)
		}
		return errs
	case *json.UnmarshalTypeError:
		errs = append(errs, map[string]string{e.Field: fmt.Sprintf("%v can not be a %v", e.Field, e.Value)})
		return errs
	}
	errs = append(errs, map[string]string{"unknown": fmt.Sprintf("unsupported custom error for: %v", err)})
	return errs
}
