package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

const defaultPort = 5432

type Config struct {
	Provider    string
	DatabaseURL string
	Host        string
	Port        int
	User        string
	Password    string
	Name        string
	SSLMode     string
}

func LoadConfigFromEnv() Config {
	_ = godotenv.Load()

	provider := strings.TrimSpace(os.Getenv("DB_PROVIDER"))
	if provider == "" {
		provider = "local"
	}

	port := defaultPort
	if rawPort := strings.TrimSpace(os.Getenv("DB_PORT")); rawPort != "" {
		if parsed, err := strconv.Atoi(rawPort); err == nil {
			port = parsed
		}
	}

	sslMode := strings.TrimSpace(os.Getenv("DB_SSLMODE"))
	if sslMode == "" {
		if strings.EqualFold(provider, "supabase") {
			sslMode = "require"
		} else {
			sslMode = "disable"
		}
	}

	return Config{
		Provider:    provider,
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),
		Host:        strings.TrimSpace(os.Getenv("DB_HOST")),
		Port:        port,
		User:        strings.TrimSpace(os.Getenv("DB_USER")),
		Password:    os.Getenv("DB_PASSWORD"),
		Name:        strings.TrimSpace(os.Getenv("DB_NAME")),
		SSLMode:     sslMode,
	}
}

func (c Config) ConnectionString() (string, error) {
	if c.DatabaseURL != "" {
		return c.DatabaseURL, nil
	}
	if c.Host == "" || c.User == "" || c.Name == "" {
		return "", errors.New("database configuration is incomplete; set DATABASE_URL or DB_HOST/DB_USER/DB_NAME")
	}

	connectionURL := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path:   c.Name,
	}
	query := connectionURL.Query()
	query.Set("sslmode", c.SSLMode)
	connectionURL.RawQuery = query.Encode()

	return connectionURL.String(), nil
}

func OpenPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	dsn, err := cfg.ConnectionString()
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func OpenSQLDB(ctx context.Context, cfg Config) (*sql.DB, error) {
	dsn, err := cfg.ConnectionString()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}
