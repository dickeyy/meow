package bot

import (
	"context"
	"fmt"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/dickeyy/meow/internal/audio"
	"github.com/dickeyy/meow/internal/commands"
	"github.com/dickeyy/meow/internal/config"
	"github.com/dickeyy/meow/internal/services/artwork"
	"github.com/dickeyy/meow/internal/services/spotify"
	"github.com/dickeyy/meow/internal/services/youtube"
	"github.com/dickeyy/meow/internal/storage"
)

type Bot struct {
	session    *discordgo.Session
	config     *config.Config
	ctx        context.Context
	sessions   map[string]*audio.Session // guildID -> audio session
	sessionsMu sync.RWMutex
	youtube    *youtube.Extractor
	spotify    *spotify.Client
	artwork    *artwork.ITunesClient
	storage    *storage.Storage
	commands   *commands.Registry
}

func New(ctx context.Context, cfg *config.Config) (*Bot, error) {
	session, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create discord session: %w", err)
	}

	session.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsGuildMessages

	b := &Bot{
		session:  session,
		config:   cfg,
		ctx:      ctx,
		sessions: make(map[string]*audio.Session),
		youtube:  youtube.NewExtractorWithCookies(cfg.YouTubeCookiesPath),
		artwork:  artwork.NewITunesClient(),
	}

	// Initialize Spotify client if credentials provided
	if cfg.SpotifyClientID != "" && cfg.SpotifyClientSecret != "" {
		b.spotify = spotify.NewClient(ctx, cfg.SpotifyClientID, cfg.SpotifyClientSecret)
	}

	// Initialize storage if database URLs provided
	if cfg.PostgresURL != "" || cfg.RedisURL != "" {
		store, err := storage.New(ctx, cfg.PostgresURL, cfg.RedisURL)
		if err != nil {
			fmt.Printf("Warning: Failed to initialize storage: %v\n", err)
		} else {
			b.storage = store
		}
	}

	// Initialize command registry
	b.commands = commands.NewRegistry(b.session, b)

	// Register handlers
	session.AddHandler(b.handleReady)
	session.AddHandler(b.handleInteractionCreate)
	session.AddHandler(b.handleVoiceStateUpdate)

	return b, nil
}

func (b *Bot) Start() error {
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("failed to open discord session: %w", err)
	}

	// Register slash commands
	if err := b.commands.RegisterCommands(); err != nil {
		return fmt.Errorf("failed to register commands: %w", err)
	}

	return nil
}

func (b *Bot) Stop() {
	// Disconnect from all voice channels
	b.sessionsMu.Lock()
	for _, s := range b.sessions {
		s.Stop()
	}
	b.sessionsMu.Unlock()

	// Remove slash commands
	b.commands.UnregisterCommands()

	// Close storage connections
	if b.storage != nil {
		b.storage.Close()
	}

	b.session.Close()
}

func (b *Bot) GetSession(guildID string) *audio.Session {
	b.sessionsMu.RLock()
	defer b.sessionsMu.RUnlock()
	return b.sessions[guildID]
}

func (b *Bot) GetOrCreateSession(guildID string) *audio.Session {
	b.sessionsMu.Lock()
	defer b.sessionsMu.Unlock()

	if s, exists := b.sessions[guildID]; exists {
		return s
	}

	s := audio.NewSession(guildID, b.config.DefaultVolume)
	b.sessions[guildID] = s
	return s
}

func (b *Bot) RemoveSession(guildID string) {
	b.sessionsMu.Lock()
	defer b.sessionsMu.Unlock()

	if s, exists := b.sessions[guildID]; exists {
		s.Stop()
		delete(b.sessions, guildID)
	}
}

func (b *Bot) Discord() *discordgo.Session {
	return b.session
}

func (b *Bot) YouTube() *youtube.Extractor {
	return b.youtube
}

func (b *Bot) Spotify() *spotify.Client {
	return b.spotify
}

func (b *Bot) Artwork() *artwork.ITunesClient {
	return b.artwork
}

func (b *Bot) Storage() *storage.Storage {
	return b.storage
}

func (b *Bot) Config() *config.Config {
	return b.config
}
