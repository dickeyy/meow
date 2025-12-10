package embeds

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dickeyy/meow/internal/audio"
)

const MaxQueueDisplay = 10

func Queue(current *audio.Track, upcoming []*audio.Track, page int) *discordgo.MessageEmbed {
	description := ""

	if current != nil {
		description = fmt.Sprintf("**Now Playing:**\n%s - %s `[%s]`\n\n",
			current.Title,
			current.Artist,
			current.FormatDuration(),
		)
	}

	if len(upcoming) == 0 {
		description += "*No upcoming tracks*"
	} else {
		description += "**Up Next:**\n"

		start := (page - 1) * MaxQueueDisplay
		end := start + MaxQueueDisplay
		if end > len(upcoming) {
			end = len(upcoming)
		}
		if start >= len(upcoming) {
			start = 0
			end = min(MaxQueueDisplay, len(upcoming))
		}

		var totalDuration time.Duration
		for _, t := range upcoming {
			totalDuration += t.Duration
		}

		for i := start; i < end; i++ {
			track := upcoming[i]
			duration := track.FormatDuration()
			description += fmt.Sprintf("`%d.` %s - %s `[%s]`\n",
				i+1,
				track.Title,
				track.Artist,
				duration,
			)
		}

		if len(upcoming) > end {
			description += fmt.Sprintf("\n*...and %d more tracks*", len(upcoming)-end)
		}

		description += fmt.Sprintf("\n\n**Total:** %d tracks | %s",
			len(upcoming),
			formatDuration(totalDuration),
		)
	}

	return &discordgo.MessageEmbed{
		Title:       "Queue",
		Description: description,
		Color:       ColorDefault,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Meow",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
