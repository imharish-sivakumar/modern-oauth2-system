package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"token-management-service/config"
	"token-management-service/domain"
	"token-management-service/grpcserver"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/imharish-sivakumar/modern-oauth2-system/service-utils/constants"
	gc "github.com/imharish-sivakumar/modern-oauth2-system/service-utils/globalconfig"
	utilsLog "github.com/imharish-sivakumar/modern-oauth2-system/service-utils/log"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var (
	// Derived from ldflags -X.
	buildRevision string
	buildVersion  string
	buildTime     string

	// general options.
	versionFlag bool
	helpFlag    bool

	// program controller.
	done    = make(chan struct{})
	grpcErr = make(chan error)

	// config.
	serviceConfig *config.AppConfig
	globalConfig  *gc.GlobalConfig

	// redis credentials.
	redisPassword string
	redisHost     string
	redisPort     string

	// global context
	ctx context.Context
)

func init() {
	ctx = context.Background()
	err := godotenv.Load()
	if err != nil {
		return
	}
	flag.BoolVar(&versionFlag, "version", false, "show current version and exit")
	flag.BoolVar(&helpFlag, "help", false, "show usage and exit")
}

func setBuildVariables() {
	if buildRevision == "" {
		buildRevision = "dev"
	}
	if buildVersion == "" {
		buildVersion = "dev"
	}
	if buildTime == "" {
		buildTime = time.Now().UTC().Format(time.RFC3339)
	}
}

func parseFlags() {
	flag.Parse()

	if helpFlag {
		flag.Usage()
		os.Exit(0)
	}

	if versionFlag {
		slog.InfoContext(ctx, "pre-flight check", slog.String("buildRevision", buildRevision), slog.Any("buildTime", buildTime))
		os.Exit(0)
	}
}

func handleInterrupts() {
	slog.InfoContext(ctx, "start handle interrupts")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	sig := <-interrupt
	slog.InfoContext(ctx, "caught sig", slog.Any("signal", sig))
	// close resource here
	done <- struct{}{}
}

func getSecrets() error {
	cfg, err := awsconig.LoadDefaultConfig(ctx)
	if err != nil {
		log.Println("unable to load config", err)
		return err
	}
	secretClient := secretsmanager.NewFromConfig(cfg)

	value, err := secretClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(serviceConfig.SecretKey),
	})
	if err != nil {
		log.Println("unable to get aws secrets ", err)
		return err
	}
	secretString := []byte(*value.SecretString)
	var data config.AppSecretKeys
	err = json.Unmarshal(secretString, &data)
	if err != nil {
		log.Println("unable to unmarshal db password from aws secrets ", err)
		return err
	}
	for clientID, oauthClient := range serviceConfig.CISAuth.Clients {
		key := strings.Split(oauthClient.Secret, ":")[1]
		slog.InfoContext(ctx, "key is", slog.Any("key", key))
		oauthClient.Secret = data[serviceConfig.CISAuth.SecretKeys[key]]
		serviceConfig.CISAuth.Clients[clientID] = oauthClient // Update the map with the modified secureClient
	}
	redisPassword = data["REDIS_DB_PASSWORD"]
	redisHost = data["REDIS_DB_HOST"]
	redisPort = data["REDIS_DB_PORT"]

	return nil
}

func main() {
	setBuildVariables()
	parseFlags()
	go handleInterrupts()

	var err error
	serviceConfig, err = config.Load()
	if err != nil {
		slog.ErrorContext(ctx, "Unable to parse application config", slog.Any(constants.Error, err))
		return
	}

	globalConfig, err = gc.Load()
	if err != nil {
		slog.ErrorContext(ctx, "unable to get global config", slog.Any(constants.Error, err))
		return
	}

	utilsLog.InitializeLogger(serviceConfig.Environment, serviceConfig.Name)
	defer utilsLog.Close()

	if err := getSecrets(); err != nil {
		slog.ErrorContext(ctx, "failed to get secrets", slog.Any(constants.Error, err))
		return
	}

	//using one common client across the service to add logger and tracing into http client
	httpClient := http.DefaultClient

	//redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     strings.Join([]string{redisHost, redisPort}, ":"),
		Password: redisPassword,
	})

	ping := redisClient.Ping(context.Background())
	if ping.Err() != nil {
		slog.ErrorContext(ctx, "unable to connect redis client", slog.Any(constants.Error, ping.Err()))
		return
	}

	auth2 := domain.NewOAuth2(httpClient, redisClient, &serviceConfig.CISAuth)

	// grpc server
	grpcHandler := grpcserver.NewGRPCHandler(auth2)

	grpcServer := grpcserver.NewGRPCServer(strings.Join([]string{"", strconv.Itoa(serviceConfig.CISAuth.GRPCPort)}, ":"), grpcHandler)

	go func(appConfig *config.App, ch chan error) {
		slog.InfoContext(ctx, "Token Service gRPC Server has started at PORT", slog.Int("gRPC Port", appConfig.GRPCPort))
		ch <- grpcServer.ListenAndServe()
	}(&serviceConfig.CISAuth, grpcErr)

	select {
	case err := <-grpcErr:
		slog.ErrorContext(ctx, "(cisAuth gRPC Server) ListenAndServe error", slog.Any(constants.Error, err))
	case <-done:
		slog.InfoContext(ctx, "shutting down server ...")
	}
	time.AfterFunc(1*time.Second, func() {
		close(done)
		close(grpcErr)
	})
}
