package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	appConfig "user-management-service/config"
	"user-management-service/handlers"

	"github.com/adjust/rmq/v5"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/gin-gonic/gin"
	"github.com/imharish-sivakumar/modern-oauth2-system/cisauth-proto/pb"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

func main() {
	errChan := make(chan error)
	ctx := context.Background()

	serviceConfig, err := appConfig.Load()
	if err != nil {
		log.Println(err)
		return
	}
	router := gin.Default()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Println("unable to load config", err)
		return
	}

	kmsClient := kms.NewFromConfig(cfg)

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

	redisAddr := fmt.Sprintf("%s:%s", data.RedisDBHost, data.RedisDBPort)
	redisClient := redis.NewClient(&redis.Options{
		Addr:            redisAddr,
		Password:        data.RedisDBPassword,
		MaxRetryBackoff: time.Duration(10) * time.Second,
		ReadTimeout:     time.Duration(10) * time.Second,
		WriteTimeout:    time.Duration(10) * time.Second,
	})

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Println("unable to ping redis db", err)
		return
	}

	connection, err := rmq.OpenConnection("user-management-service", "tcp", redisAddr, 1, errChan)
	if err != nil {
		log.Println("unable to open connection for rmq")
		return
	}

	emailQueue, err := connection.OpenQueue("email")

	conn, err := grpc.NewClient(serviceConfig.TMSHost)
	//if err != nil {
	//	log.Println("unable to create grpc client", err)
	//	return
	//}

	tmsClient := pb.NewTokenServiceClient(conn)

	handler := handlers.NewHandler(kmsClient, &tmsClient, serviceConfig, redisClient, emailQueue)

	routerGroup := router.Group("/user-service/v1")
	routerGroup.Handle(http.MethodPost, "/register", handler.Register)

	if err = router.Run(fmt.Sprintf(":%d", serviceConfig.Port)); err != nil {
		log.Println(err)
		return
	}
}
