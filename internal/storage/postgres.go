package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
	ctx  context.Context
}

func NewPostgresStore(ctx context.Context, url string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &PostgresStore{pool: pool, ctx: ctx}

	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
}

func (s *PostgresStore) migrate() error {
	query := `
		CREATE TABLE IF NOT EXISTS guild_settings (
			guild_id VARCHAR(255) PRIMARY KEY,
			default_volume INTEGER DEFAULT 50,
			dj_role_id VARCHAR(255) DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`

	_, err := s.pool.Exec(s.ctx, query)
	return err
}

func (s *PostgresStore) Close() {
	s.pool.Close()
}

func (s *PostgresStore) GetGuildSettings(guildID string) (*GuildSettings, error) {
	query := `
		SELECT guild_id, default_volume, dj_role_id, created_at, updated_at 
		FROM guild_settings 
		WHERE guild_id = $1
	`

	settings := &GuildSettings{}
	err := s.pool.QueryRow(s.ctx, query, guildID).Scan(
		&settings.GuildID,
		&settings.DefaultVolume,
		&settings.DJRoleID,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if err != nil {
		// Return default settings if not found
		return DefaultGuildSettings(guildID), nil
	}

	return settings, nil
}

func (s *PostgresStore) SaveGuildSettings(settings *GuildSettings) error {
	settings.UpdatedAt = time.Now()

	query := `
		INSERT INTO guild_settings (guild_id, default_volume, dj_role_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (guild_id) DO UPDATE SET
			default_volume = EXCLUDED.default_volume,
			dj_role_id = EXCLUDED.dj_role_id,
			updated_at = EXCLUDED.updated_at
	`

	_, err := s.pool.Exec(s.ctx, query,
		settings.GuildID,
		settings.DefaultVolume,
		settings.DJRoleID,
		settings.CreatedAt,
		settings.UpdatedAt,
	)

	return err
}

