package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dickeyy/meow/internal/embeds"
)

func handleVolume(s *discordgo.Session, i *discordgo.InteractionCreate, bot BotInterface) {
	session := bot.GetSession(i.GuildID)
	if session == nil {
		respond(s, i, embeds.Error("Error", "No active session. Start playing something first."))
		return
	}

	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		// Show current volume
		respond(s, i, embeds.Info("Volume", fmt.Sprintf("Current volume: **%d%%**", session.Volume())))
		return
	}

	level := int(options[0].IntValue())
	session.SetVolume(level)

	// Create volume bar visualization
	volumeBar := createVolumeBar(level)
	
	respond(s, i, embeds.Success("Volume", fmt.Sprintf("%s **%d%%**", volumeBar, level)))
}

func createVolumeBar(level int) string {
	filled := level / 10
	empty := 10 - filled
	
	bar := ""
	for i := 0; i < filled; i++ {
		bar += "="
	}
	for i := 0; i < empty; i++ {
		bar += "-"
	}
	
	return "[" + bar + "]"
}

