package apperror

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"reflect"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var (
	isRequired                  = "is required"
	formatIsIncorrect           = "format is incorrect"
	mustBeAtleastFiveCharLong   = "must be atleast 5 characters long"
	mustBeAtmostHundredCharLong = "must be atmost 100 characters long"
	shouldBeOfTypeAlpha         = "should be of type alpha"
	shouldBeOfTypePassword      = "should contain minimum 8, maximum 128 characters, 1 uppercase, 1 lowercase, 1 digit & 1 special character"
)

var (
	customErrors = map[string]error{
		"uuid.required":                 errors.New(isRequired),
		"uuid.uuid":                     errors.New(formatIsIncorrect),
		"email.required":                errors.New(isRequired),
		"email.email":                   errors.New("should be in email format"),
		"email.min":                     errors.New(mustBeAtleastFiveCharLong),
		"email.max":                     errors.New("must be atmost 50 characters long"),
		"email.domainMXRecord":          errors.New("email service provider not accepted"),
		"name.required":                 errors.New(isRequired),
		"name.min":                      errors.New(mustBeAtleastFiveCharLong),
		"name.max":                      errors.New(mustBeAtmostHundredCharLong),
		"name.alphaSpace":               errors.New(shouldBeOfTypeAlpha),
		"newPassword.required":          errors.New(isRequired),
		"password.required":             errors.New(isRequired),
		"password.password":             errors.New(shouldBeOfTypePassword),
		"oldPassword.required":          errors.New(isRequired),
		"oldPassword.password":          errors.New(shouldBeOfTypePassword),
		"oldPassword.nefield":           errors.New("old password cannot be same as new password"),
		"confirmPassword.required":      errors.New(isRequired),
		"confirmPassword.eqfield":       errors.New("confirm password and password must be equal"),
		"changePassword.nefield":        errors.New("unsupported request"),
		"loginChallenge.required":       errors.New(isRequired),
		"loginChallenge.loginChallenge": errors.New("should be 32 bit alphanumeric value"),
		"ConsentChallenge.required":     errors.New(isRequired),
		"redirectURI.required":          errors.New(isRequired),
		"code.required":                 errors.New(isRequired),
		"clientID.required":             errors.New(isRequired),
		"codeVerifier.required":         errors.New(isRequired),
		"consent_challenge.required":    errors.New(isRequired),

		// related to service config
		"appinsightsInstrumentationKey.required": errors.New(isRequired),
		"privateKey.required":                    errors.New(isRequired),
		"publicKey.required":                     errors.New(isRequired),
		"emailHostURL.required":                  errors.New(isRequired),
		"smtpServerAddress.required":             errors.New(isRequired),
		"dbHost.required":                        errors.New(isRequired),
		"dbPort.required":                        errors.New(isRequired),
		"dbName.required":                        errors.New(isRequired),
		"dbUser.required":                        errors.New(isRequired),
		"port.required":                          errors.New(isRequired),
		"dbPassword.required":                    errors.New(isRequired),
		"passwordPrivateKey.required":            errors.New(isRequired),
		"secretKeys.required":                    errors.New(isRequired),
		"keyList.required":                       errors.New(isRequired),
		"environment.oneof":                      errors.New("should be one of allowed values"),
		"address.required":                       errors.New(isRequired),
		"address.upi":                            errors.New("should be of upi format"),
		"address.provider":                       errors.New("is not valid provider"),
		"vpaid.uuid":                             errors.New("format is incorrect"),
		"isDefault.required":                     errors.New(isRequired),
	}
)

var Validator *validator.Validate

func init() {
	var ok bool
	if Validator, ok = binding.Validator.Engine().(*validator.Validate); ok {
		_ = Validator.RegisterValidation("domainMXRecord", domainMXRecord, false)
		_ = Validator.RegisterValidation("loginChallenge", loginChallenge, false)
		Validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
			tags := []string{"json", "uri", "form"}
			for _, key := range tags {
				tag := fld.Tag.Get(key)
				name := strings.SplitN(tag, ",", 2)[0]
				if name == "-" {
					return ""
				} else if len(name) != 0 {
					return name
				}
			}
			return ""
		})
	}
}

var domainMXRecord validator.Func = func(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if ok && len(value) != 0 {
		// check whether the domain has a valid MX record.
		emailSplit := strings.Split(value, "@")
		mx, err := net.LookupMX(emailSplit[1])
		if (err != nil) || (len(mx) == 0) {
			return false
		}
		return true
	}
	return false
}

var loginChallenge validator.Func = func(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if ok {
		compile := regexp.MustCompile(`^[a-z0-9_\-A-Z=]*$`)
		return compile.MatchString(value)
	}
	return false
}

// CustomValidationError converts validation and json marshal error into custom error type.
func CustomValidationError(err error) []map[string]string {
	errs := make([]map[string]string, 0)
	switch errType := err.(type) {
	case validator.ValidationErrors:
		for _, e := range errType {
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
		errs = append(errs, map[string]string{errType.Field: fmt.Sprintf("%v can not be a %v", errType.Field, errType.Value)})
		return errs
	}
	errs = append(errs, map[string]string{"unknown": fmt.Sprintf("unsupported custom error for: %v", err)})
	return errs
}
