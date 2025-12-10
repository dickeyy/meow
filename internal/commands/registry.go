package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dickeyy/meow/internal/audio"
	"github.com/dickeyy/meow/internal/config"
	"github.com/dickeyy/meow/internal/services/artwork"
	"github.com/dickeyy/meow/internal/services/spotify"
	"github.com/dickeyy/meow/internal/services/youtube"
	"github.com/dickeyy/meow/internal/storage"
)

// BotInterface defines what the commands package needs from the bot
type BotInterface interface {
	Discord() *discordgo.Session
	YouTube() *youtube.Extractor
	Spotify() *spotify.Client
	Artwork() *artwork.ITunesClient
	Storage() *storage.Storage
	Config() *config.Config
	GetSession(guildID string) *audio.Session
	GetOrCreateSession(guildID string) *audio.Session
	RemoveSession(guildID string)
}

type CommandHandler func(s *discordgo.Session, i *discordgo.InteractionCreate, bot BotInterface)
type ComponentHandler func(s *discordgo.Session, i *discordgo.InteractionCreate, bot BotInterface)

type Registry struct {
	session           *discordgo.Session
	bot               BotInterface
	commands          []*discordgo.ApplicationCommand
	handlers          map[string]CommandHandler
	componentHandlers map[string]ComponentHandler
	registeredCmds    []*discordgo.ApplicationCommand
}

func NewRegistry(session *discordgo.Session, bot BotInterface) *Registry {
	r := &Registry{
		session:           session,
		bot:               bot,
		commands:          make([]*discordgo.ApplicationCommand, 0),
		handlers:          make(map[string]CommandHandler),
		componentHandlers: make(map[string]ComponentHandler),
	}

	r.registerCommands()
	return r
}

func (r *Registry) registerCommands() {
	// Play command
	r.addCommand(&discordgo.ApplicationCommand{
		Name:        "play",
		Description: "Play a song or playlist from YouTube, Spotify, or other sources",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "query",
				Description: "URL or search query",
				Required:    true,
			},
		},
	}, handlePlay)

	// Pause command
	r.addCommand(&discordgo.ApplicationCommand{
		Name:        "pause",
		Description: "Pause the current track",
	}, handlePause)

	// Resume command
	r.addCommand(&discordgo.ApplicationCommand{
		Name:        "resume",
		Description: "Resume playback",
	}, handleResume)

	// Skip command
	r.addCommand(&discordgo.ApplicationCommand{
		Name:        "skip",
		Description: "Skip to the next track",
	}, handleSkip)

	// Previous command
	r.addCommand(&discordgo.ApplicationCommand{
		Name:        "previous",
		Description: "Go back to the previous track",
	}, handlePrevious)

	// Stop command
	r.addCommand(&discordgo.ApplicationCommand{
		Name:        "stop",
		Description: "Stop playback and clear the queue",
	}, handleStop)

	// Volume command
	r.addCommand(&discordgo.ApplicationCommand{
		Name:        "volume",
		Description: "Set the playback volume",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "level",
				Description: "Volume level (0-100)",
				Required:    true,
				MinValue:    floatPtr(0),
				MaxValue:    100,
			},
		},
	}, handleVolume)

	// Queue command
	r.addCommand(&discordgo.ApplicationCommand{
		Name:        "queue",
		Description: "View or manage the queue",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "view",
				Description: "View the current queue",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "move",
				Description: "Move a track in the queue",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "from",
						Description: "Position to move from",
						Required:    true,
						MinValue:    floatPtr(1),
					},
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "to",
						Description: "Position to move to",
						Required:    true,
						MinValue:    floatPtr(1),
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "Remove a track from the queue",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "position",
						Description: "Position of the track to remove",
						Required:    true,
						MinValue:    floatPtr(1),
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "clear",
				Description: "Clear all tracks from the queue",
			},
		},
	}, handleQueue)

	// Shuffle command
	r.addCommand(&discordgo.ApplicationCommand{
		Name:        "shuffle",
		Description: "Shuffle the queue",
	}, handleShuffle)

	// Now playing command
	r.addCommand(&discordgo.ApplicationCommand{
		Name:        "nowplaying",
		Description: "Show the currently playing track",
	}, handleNowPlaying)

	// Register component handlers
	r.componentHandlers["player_pause"] = handlePlayerPause
	r.componentHandlers["player_resume"] = handlePlayerResume
	r.componentHandlers["player_skip"] = handlePlayerSkip
	r.componentHandlers["player_previous"] = handlePlayerPrevious
	r.componentHandlers["player_stop"] = handlePlayerStop
	r.componentHandlers["player_queue"] = handlePlayerQueue
}

func (r *Registry) addCommand(cmd *discordgo.ApplicationCommand, handler CommandHandler) {
	r.commands = append(r.commands, cmd)
	r.handlers[cmd.Name] = handler
}

func (r *Registry) RegisterCommands() error {
	for _, cmd := range r.commands {
		registered, err := r.session.ApplicationCommandCreate(r.session.State.User.ID, "", cmd)
		if err != nil {
			return fmt.Errorf("failed to register command %s: %w", cmd.Name, err)
		}
		r.registeredCmds = append(r.registeredCmds, registered)
		fmt.Printf("Registered command: /%s\n", cmd.Name)
	}
	return nil
}

func (r *Registry) UnregisterCommands() {
	for _, cmd := range r.registeredCmds {
		err := r.session.ApplicationCommandDelete(r.session.State.User.ID, "", cmd.ID)
		if err != nil {
			fmt.Printf("Failed to unregister command %s: %v\n", cmd.Name, err)
		}
	}
}

func (r *Registry) HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if handler, exists := r.handlers[i.ApplicationCommandData().Name]; exists {
		handler(s, i, r.bot)
	}
}

func (r *Registry) HandleComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID
	if handler, exists := r.componentHandlers[customID]; exists {
		handler(s, i, r.bot)
	}
}

func floatPtr(f float64) *float64 {
	return &f
}

