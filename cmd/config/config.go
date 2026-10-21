package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	ServerPort  string
	DatabaseURL string
	RedisURL    string
}

func Load() (Config, error) {
	port := strings.TrimSpace(os.Getenv("SERVER_PORT"))
	if port == "" {
		port = "8080"
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		username := os.Getenv("POSTGRES_USERNAME")
		password := os.Getenv("POSTGRES_PASSWORD")
		dbName := os.Getenv("DB_NAME")

		if username == "" || dbName == "" {
			return Config{}, fmt.Errorf(
				"DATABASE_URL or POSTGRES_USERNAME and DB_NAME must be set",
			)
		}

		u := &url.URL{
			Scheme: "postgres",
			User:   url.UserPassword(username, password),
			Host:   "localhost:5432",
			Path:   "/" + dbName,
		}
		q := u.Query()
		q.Set("sslmode", "disable")
		u.RawQuery = q.Encode()

		databaseURL = u.String()
	}

	redisURL := strings.TrimSpace(os.Getenv("REDIS_URL"))
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	return Config{
		ServerPort:  port,
		DatabaseURL: databaseURL,
		RedisURL:    redisURL,
	}, nil
}
