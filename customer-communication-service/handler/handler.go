package handler

import (
	"encoding/json"
	"log"
	"log/slog"

	"customer-communication-service/models"

	"github.com/adjust/rmq/v5"
	"github.com/imharish-sivakumar/modern-oauth2-system/service-utils/constants"
	"gopkg.in/gomail.v2"
)

type EmailNotificationConsumer struct {
	gomailer *gomail.Dialer
	from     string
}

func NewEmailNotificationConsumer(gomailer *gomail.Dialer, from string) *EmailNotificationConsumer {
	return &EmailNotificationConsumer{gomailer: gomailer, from: from}
}

func (consumer *EmailNotificationConsumer) Consume(delivery rmq.Delivery) {
	var task models.Event
	if err := json.Unmarshal([]byte(delivery.Payload()), &task); err != nil {
		if err := delivery.Reject(); err != nil {
			slog.Error("unable to process the data", slog.Any(constants.Error, err))
		}
		return
	}

	// perform task
	log.Printf("performing task %s", task)
	if err := delivery.Ack(); err != nil {
		// handle ack error
		slog.Error("unable to acknowledge the data ", slog.Any(constants.Error, err))
	}

	template, err := task.GetTemplate()
	if err != nil {
		log.Println("unable to get template", err)
		return
	}

	gomailMessage := gomail.NewMessage()
	gomailMessage.SetHeader("From", consumer.from)
	gomailMessage.SetHeaders(map[string][]string{"To": {task.Email}})

	gomailMessage.SetHeader("Subject", task.GetSubject())
	gomailMessage.SetBody("text/html", template)

	if err := consumer.gomailer.DialAndSend(gomailMessage); err != nil {
		log.Println("unable to dial and send", err)
		return
	}
	return
}
