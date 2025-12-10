package embeds

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dickeyy/meow/internal/audio"
)

func NowPlaying(track *audio.Track, session *audio.Session) *discordgo.MessageEmbed {
	elapsed := session.Elapsed()
	total := track.Duration

	// Create a nicer progress bar
	progressBar := createProgressBar(elapsed, total)
	timeDisplay := fmt.Sprintf("`%s`  %s  `%s`",
		formatDuration(elapsed),
		progressBar,
		formatDuration(total),
	)

	description := fmt.Sprintf("**%s**\n%s\n\n%s",
		track.Title,
		track.Artist,
		timeDisplay,
	)

	embed := &discordgo.MessageEmbed{
		Title:       "Now Playing",
		Description: description,
		Color:       ColorDefault,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Meow",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if track.Thumbnail != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: track.Thumbnail,
		}
	}

	if track.Album != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Album",
			Value:  track.Album,
			Inline: true,
		})
	}

	if track.RequestedBy != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Requested by",
			Value:  fmt.Sprintf("<@%s>", track.RequestedBy),
			Inline: true,
		})
	}

	return embed
}

func createProgressBar(elapsed, total time.Duration) string {
	const barLength = 16

	if total == 0 {
		return repeatString("-", barLength)
	}

	progress := float64(elapsed) / float64(total)
	if progress > 1 {
		progress = 1
	}
	if progress < 0 {
		progress = 0
	}

	filled := int(progress * barLength)

	// Use unicode characters for a cleaner look
	bar := ""
	for i := 0; i < barLength; i++ {
		if i < filled {
			bar += "\u2501" // horizontal line (filled)
		} else if i == filled {
			bar += "\u25CF" // circle (current position)
		} else {
			bar += "\u2500" // light horizontal line (empty)
		}
	}

	return bar
}

func repeatString(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}

	total := int(d.Seconds())
	hours := total / 3600
	minutes := (total % 3600) / 60
	seconds := total % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}
