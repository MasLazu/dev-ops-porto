package server

import (
	"fmt"
	"os"
	"strconv"

	"github.com/MasLazu/dev-ops-porto/pkg/database"
)

type config struct {
	port                     int
	otlpDomain               string
	grpcMissionServiceDomain string
	database                 database.Config
	serviceName              string
	jwtSecret                []byte
	grpcAuthServiceDomain    string
}

func getConfig() (config, error) {
	port, err := getIntEnv("PORT")
	if err != nil {
		return config{}, err
	}

	dbConfig, err := getDatabaseConfig()
	if err != nil {
		return config{}, fmt.Errorf("failed to get database config: %w", err)
	}

	return config{
		port:                     port,
		otlpDomain:               os.Getenv("OTLP_DOMAIN"),
		jwtSecret:                []byte(os.Getenv("JWT_SECRET")),
		database:                 dbConfig,
		grpcMissionServiceDomain: os.Getenv("GRPC_MISSION_SERVICE_DOMAIN"),
		serviceName:              "theme-service",
		grpcAuthServiceDomain:    os.Getenv("GRPC_AUTH_SERVICE_DOMAIN"),
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
