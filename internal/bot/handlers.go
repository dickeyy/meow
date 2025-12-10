package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func (b *Bot) handleReady(s *discordgo.Session, r *discordgo.Ready) {
	fmt.Printf("Logged in as %s#%s\n", r.User.Username, r.User.Discriminator)

	err := s.UpdateStatusComplex(discordgo.UpdateStatusData{
		Activities: []*discordgo.Activity{
			{
				Name: "/play",
				Type: discordgo.ActivityTypeListening,
			},
		},
		Status: "online",
	})
	if err != nil {
		fmt.Printf("Failed to update status: %v\n", err)
	}
}

func (b *Bot) handleInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		b.commands.HandleCommand(s, i)
	case discordgo.InteractionMessageComponent:
		b.commands.HandleComponent(s, i)
	}
}

func (b *Bot) handleVoiceStateUpdate(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	// Check if it's the bot leaving a voice channel
	if v.UserID == s.State.User.ID && v.ChannelID == "" {
		b.RemoveSession(v.GuildID)
		return
	}

	// Check if bot is alone in voice channel
	session := b.GetSession(v.GuildID)
	if session == nil {
		return
	}

	vc := session.VoiceConnection()
	if vc == nil {
		return
	}

	// Get the channel the bot is in
	guild, err := s.State.Guild(v.GuildID)
	if err != nil {
		return
	}

	// Count users in the bot's voice channel
	userCount := 0
	for _, vs := range guild.VoiceStates {
		if vs.ChannelID == vc.ChannelID && vs.UserID != s.State.User.ID {
			userCount++
		}
	}

	// If bot is alone, disconnect
	if userCount == 0 {
		session.Stop()
		vc.Disconnect()
		b.RemoveSession(v.GuildID)
	}
}

