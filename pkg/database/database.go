package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

type Config struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
	Schema   string
}

type Service struct {
	Pool   *sql.DB
	config Config
}

func New(config Config) (*Service, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable&search_path=%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.Schema,
	)
	if err := migrateSql(connStr); err != nil {
		return nil, fmt.Errorf("failed to migrate sql: %w", err)
	}

	db, err := openConnSql(connStr, config)
	if err != nil {
		log.Fatal(err)
	}

	return &Service{
		Pool:   db,
		config: config,
	}, nil
}

func migrateSql(connStr string) error {
	m, err := migrate.New("file://migrations", connStr)
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}
	return nil
}

func openConnSql(connStr string, c Config) (*sql.DB, error) {
	db, err := otelsql.Open("pgx", connStr, otelsql.WithAttributes(semconv.DBSystemPostgreSQL), otelsql.WithDBName(c.Database))
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (s *Service) Health(ctx context.Context) map[string]string {
	stats := make(map[string]string)

	// Ping the database
	err := s.Pool.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.Pool.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

func (s *Service) Close() error {
	log.Printf("Disconnected from database: %s", s.config.Database)
	return s.Pool.Close()
}
