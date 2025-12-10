package storage

import (
	"context"
	"fmt"
	"time"
)

type Storage struct {
	postgres *PostgresStore
	redis    *RedisStore
	ctx      context.Context
}

func New(ctx context.Context, postgresURL, redisURL string) (*Storage, error) {
	s := &Storage{ctx: ctx}

	if postgresURL != "" {
		pg, err := NewPostgresStore(ctx, postgresURL)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to postgres: %w", err)
		}
		s.postgres = pg
	}

	if redisURL != "" {
		r, err := NewRedisStore(ctx, redisURL)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to redis: %w", err)
		}
		s.redis = r
	}

	return s, nil
}

func (s *Storage) Close() {
	if s.postgres != nil {
		s.postgres.Close()
	}
	if s.redis != nil {
		s.redis.Close()
	}
}

func (s *Storage) GetGuildSettings(guildID string) (*GuildSettings, error) {
	if s.postgres == nil {
		return DefaultGuildSettings(guildID), nil
	}
	return s.postgres.GetGuildSettings(guildID)
}

func (s *Storage) SaveGuildSettings(settings *GuildSettings) error {
	if s.postgres == nil {
		return nil
	}
	return s.postgres.SaveGuildSettings(settings)
}

func (s *Storage) CacheStreamURL(trackID, streamURL string) error {
	if s.redis == nil {
		return nil
	}
	// Cache for 5 hours (YouTube URLs typically expire after 6)
	return s.redis.Set(s.ctx, "stream:"+trackID, streamURL, 5*time.Hour)
}

func (s *Storage) GetCachedStreamURL(trackID string) (string, error) {
	if s.redis == nil {
		return "", nil
	}
	return s.redis.Get(s.ctx, "stream:"+trackID)
}

