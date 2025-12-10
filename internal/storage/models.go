package storage

import "time"

type GuildSettings struct {
	GuildID       string    `json:"guild_id"`
	DefaultVolume int       `json:"default_volume"`
	DJRoleID      string    `json:"dj_role_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func DefaultGuildSettings(guildID string) *GuildSettings {
	return &GuildSettings{
		GuildID:       guildID,
		DefaultVolume: 50,
		DJRoleID:      "",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

