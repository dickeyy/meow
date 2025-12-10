package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DiscordToken        string
	SpotifyClientID     string
	SpotifyClientSecret string
	PostgresURL         string
	RedisURL            string
	DefaultVolume       int
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		// .env file is optional in production
		fmt.Println("No .env file found, using environment variables")
	}

	cfg := &Config{
		DiscordToken:        os.Getenv("DISCORD_TOKEN"),
		SpotifyClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		SpotifyClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		PostgresURL:         os.Getenv("POSTGRES_URL"),
		RedisURL:            os.Getenv("REDIS_URL"),
		DefaultVolume:       50,
	}

	if vol := os.Getenv("DEFAULT_VOLUME"); vol != "" {
		if v, err := strconv.Atoi(vol); err == nil && v >= 0 && v <= 100 {
			cfg.DefaultVolume = v
		}
	}

	if cfg.DiscordToken == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN is required")
	}

	return cfg, nil
}
