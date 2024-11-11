package models

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"os"
)

var (
	templateMapping = map[EventType]string{
		VerificationEvent: "templates/verify_email.html",
	}
)

type EventType string

const (
	VerificationEvent EventType = "VerificationEvent"
)

type VerificationPayload struct {
	VerificationID string `json:"verificationId"`
}

type Event struct {
	Email        string
	Type         EventType
	EventPayload []byte
}

func (e *Event) GetSubject() string {
	switch e.Type {
	case VerificationEvent:
		return "Verify your email address"
	}
	return ""
}

func (e *Event) GetEventData() interface{} {
	switch e.Type {
	case VerificationEvent:
		return VerificationPayload{}
	}
	return nil
}

func (e *Event) GetTemplate() (string, error) {
	var data interface{}
	switch e.Type {
	case VerificationEvent:
		var verificationPayload VerificationPayload
		if err := json.Unmarshal(e.EventPayload, &verificationPayload); err != nil {
			return "", err
		}
		data = verificationPayload
	}

	contentBuffer := new(bytes.Buffer)
	var err error

	file, err := os.ReadFile(templateMapping[e.Type])
	if err != nil {
		log.Println("unable to read templates", err)
		return "", err
	}

	// adding func to avoid escaping conditional HTML comments
	finalTemplate, err := template.New("email").Funcs(template.FuncMap{
		"safe": func(s string) template.HTML { return template.HTML(s) },
	}).Parse(string(file))

	if err != nil {
		log.Println(err)
		return "", err
	}

	err = finalTemplate.Execute(contentBuffer, data)
	if err != nil {
		log.Println("Error while executing template ", err)
		return "", err
	}
	return contentBuffer.String(), nil
}
