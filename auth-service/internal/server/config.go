package server

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/MasLazu/dev-ops-porto/pkg/database"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

type bucketNames struct {
	profilePictures string
}

type s3Config struct {
	enpoint     string
	bucketNames bucketNames
}

type AwsConfig struct {
	awsConfig aws.Config
	s3        s3Config
}

type config struct {
	port                 int
	serviceName          string
	otlpDomain           string
	jwtSecret            []byte
	database             database.Config
	aws                  AwsConfig
	staticServiceEnpoint string
}

func getConfig(ctx context.Context) (config, error) {
	port, err := getIntEnv("PORT")
	if err != nil {
		return config{}, err
	}

	dbConfig, err := getDatabaseConfig()
	if err != nil {
		return config{}, fmt.Errorf("failed to get database config: %w", err)
	}

	awsConfig, err := getAwsConfig(ctx)
	if err != nil {
		return config{}, fmt.Errorf("failed to get S3 config: %w", err)
	}

	return config{
		port:                 port,
		otlpDomain:           os.Getenv("OTLP_DOMAIN"),
		jwtSecret:            []byte(os.Getenv("JWT_SECRET")),
		serviceName:          "auth-service",
		database:             dbConfig,
		staticServiceEnpoint: os.Getenv("PUBLIC_STATIC_SERVICE_ENDPOINT"),
		aws: AwsConfig{
			awsConfig: awsConfig,
			s3: s3Config{
				enpoint: os.Getenv("S3_ENDPOINT"),
				bucketNames: bucketNames{
					profilePictures: os.Getenv("S3_BUCKET_PROFILE_PICTURES"),
				},
			},
		},
	}, nil
}

func getIntEnv(key string) (int, error) {
	value := os.Getenv(key)

	i, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}

	return i, nil
}

func getAwsConfig(ctx context.Context) (aws.Config, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("us-west-2"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(os.Getenv("S3_ACCESS_KEY"), os.Getenv("S3_SECRET_KEY"), "")),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return awsCfg, nil
}

func getDatabaseConfig() (database.Config, error) {
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return database.Config{}, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	return database.Config{
		Host:              os.Getenv("DB_HOST"),
		Port:              port,
		Database:          os.Getenv("DB_DATABASE"),
		Username:          os.Getenv("DB_USERNAME"),
		Password:          os.Getenv("DB_PASSWORD"),
		Schema:            os.Getenv("DB_SCHEMA"),
		MigrationLocation: "file://migrations",
	}, nil
}
