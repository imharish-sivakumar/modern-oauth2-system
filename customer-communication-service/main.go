package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	appConfig "customer-communication-service/config"
	"customer-communication-service/handler"

	"github.com/adjust/rmq/v5"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"gopkg.in/gomail.v2"
)

func main() {
	done := make(chan struct{})
	errChan := make(chan error)
	ctx := context.Background()

	serviceConfig, err := appConfig.Load()
	if err != nil {
		log.Println(err)
		return
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Println("unable to load config", err)
		return
	}

	secretClient := secretsmanager.NewFromConfig(cfg)

	value, err := secretClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(serviceConfig.SecretKey),
	})
	if err != nil {
		log.Println("unable to get aws secrets ", err)
		return
	}
	secretString := []byte(*value.SecretString)
	var data appConfig.Secrets
	err = json.Unmarshal(secretString, &data)
	if err != nil {
		log.Println("unable to unmarshal db password from aws secrets ", err)
		return
	}

	connection, err := rmq.OpenConnection("user-management-service", "tcp", fmt.Sprintf("%s:%s", data.RedisDBHost, data.RedisDBPort), 1, errChan)
	if err != nil {
		log.Println("unable to open connection for rmq")
		return
	}

	emailQueue, err := connection.OpenQueue("email")

	if err != nil {
		log.Println("unable to open queue", err)
		return
	}
	err = emailQueue.StartConsuming(10, time.Second)
	if err != nil {
		log.Println("unable to start consuming ", err)
		return
	}

	smtpPort, err := strconv.Atoi(data.SMTPPort)
	if err != nil {
		log.Println("unable to convert smtp port", err)
		return
	}

	gomailDialer := gomail.NewDialer(data.SMTPHost, smtpPort, data.SMTPUsername, data.SMTPPassword)

	consumer := handler.NewEmailNotificationConsumer(gomailDialer, serviceConfig.FromEmail)
	_, err = emailQueue.AddConsumerFunc("tag", consumer.Consume)
	if err != nil {
		log.Println("unable to add consumer", err)
		return
	}

	select {
	case <-done:
		log.Println("exiting")
	}
}
