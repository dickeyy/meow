package embeds

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

// Discord's default dark theme embed color (no visible border)
const ColorDefault = 0x2B2D31

func PlayerButtons(isPaused bool) []discordgo.MessageComponent {
	var playPauseButton discordgo.Button

	if isPaused {
		playPauseButton = discordgo.Button{
			CustomID: "player_resume",
			Label:    "Play",
			Style:    discordgo.SuccessButton,
		}
	} else {
		playPauseButton = discordgo.Button{
			CustomID: "player_pause",
			Label:    "Pause",
			Style:    discordgo.SecondaryButton,
		}
	}

	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: "player_previous",
					Label:    "Previous",
					Style:    discordgo.SecondaryButton,
				},
				playPauseButton,
				discordgo.Button{
					CustomID: "player_skip",
					Label:    "Skip",
					Style:    discordgo.SecondaryButton,
				},
				discordgo.Button{
					CustomID: "player_queue",
					Label:    "Queue",
					Style:    discordgo.SecondaryButton,
				},
				discordgo.Button{
					CustomID: "player_stop",
					Label:    "Stop",
					Style:    discordgo.DangerButton,
				},
			},
		},
	}
}

func Success(title, description string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       ColorDefault,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Meow",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

func Error(title, description string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       ColorDefault,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Meow",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

func Info(title, description string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       ColorDefault,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Meow",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
}
