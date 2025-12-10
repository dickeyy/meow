package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/dickeyy/meow/internal/embeds"
)

func handlePause(s *discordgo.Session, i *discordgo.InteractionCreate, bot BotInterface) {
	session := bot.GetSession(i.GuildID)
	if session == nil || session.IsStopped() {
		respond(s, i, embeds.Error("Error", "Nothing is playing"))
		return
	}

	if session.IsPaused() {
		respond(s, i, embeds.Info("Paused", "Playback is already paused"))
		return
	}

	session.Pause()
	respond(s, i, embeds.Success("Paused", "Playback paused"))
}

func handleResume(s *discordgo.Session, i *discordgo.InteractionCreate, bot BotInterface) {
	session := bot.GetSession(i.GuildID)
	if session == nil || session.IsStopped() {
		respond(s, i, embeds.Error("Error", "Nothing is playing"))
		return
	}

	if !session.IsPaused() {
		respond(s, i, embeds.Info("Playing", "Playback is not paused"))
		return
	}

	session.Resume()
	respond(s, i, embeds.Success("Resumed", "Playback resumed"))
}

func handleSkip(s *discordgo.Session, i *discordgo.InteractionCreate, bot BotInterface) {
	session := bot.GetSession(i.GuildID)
	if session == nil || session.IsStopped() {
		respond(s, i, embeds.Error("Error", "Nothing is playing"))
		return
	}

	if !session.Queue().HasNext() {
		session.Stop()
		respond(s, i, embeds.Info("Queue Empty", "No more tracks in queue"))
		return
	}

	session.Skip()
	
	next := session.Queue().Peek(1)
	if len(next) > 0 {
		respond(s, i, embeds.Success("Skipped", "Now playing: **"+next[0].Title+"**"))
	} else {
		respond(s, i, embeds.Success("Skipped", "Playing next track"))
	}
}

func handlePrevious(s *discordgo.Session, i *discordgo.InteractionCreate, bot BotInterface) {
	session := bot.GetSession(i.GuildID)
	if session == nil || session.IsStopped() {
		respond(s, i, embeds.Error("Error", "Nothing is playing"))
		return
	}

	if !session.Queue().HasPrevious() {
		respond(s, i, embeds.Error("Error", "No previous track"))
		return
	}

	prev := session.Queue().Previous()
	if prev != nil {
		session.Skip() // This will trigger playback of the previous track (now at front)
		respond(s, i, embeds.Success("Previous", "Now playing: **"+prev.Title+"**"))
	} else {
		respond(s, i, embeds.Error("Error", "Failed to go to previous track"))
	}
}

func handleStop(s *discordgo.Session, i *discordgo.InteractionCreate, bot BotInterface) {
	session := bot.GetSession(i.GuildID)
	if session == nil {
		respond(s, i, embeds.Error("Error", "Nothing is playing"))
		return
	}

	session.Stop()
	
	// Disconnect from voice
	if vc := session.VoiceConnection(); vc != nil {
		vc.Disconnect()
	}
	
	bot.RemoveSession(i.GuildID)
	respond(s, i, embeds.Success("Stopped", "Playback stopped and queue cleared"))
}

func handleShuffle(s *discordgo.Session, i *discordgo.InteractionCreate, bot BotInterface) {
	session := bot.GetSession(i.GuildID)
	if session == nil || session.Queue().IsEmpty() {
		respond(s, i, embeds.Error("Error", "Nothing in the queue"))
		return
	}

	if session.Queue().UpcomingLen() < 2 {
		respond(s, i, embeds.Info("Queue", "Not enough tracks to shuffle"))
		return
	}

	session.Queue().Shuffle()
	respond(s, i, embeds.Success("Shuffled", "Queue has been shuffled"))
}

func handleNowPlaying(s *discordgo.Session, i *discordgo.InteractionCreate, bot BotInterface) {
	session := bot.GetSession(i.GuildID)
	if session == nil || session.Queue().IsEmpty() {
		respond(s, i, embeds.Error("Error", "Nothing is playing"))
		return
	}

	track := session.Queue().Current()
	if track == nil {
		respond(s, i, embeds.Error("Error", "Nothing is playing"))
		return
	}

	embed := embeds.NowPlaying(track, session)
	components := embeds.PlayerButtons(session.IsPaused())

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}

func respond(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// Component handlers for button interactions
func handlePlayerPause(s *discordgo.Session, i *discordgo.InteractionCreate, bot BotInterface) {
	session := bot.GetSession(i.GuildID)
	if session == nil || session.IsStopped() {
		respondComponent(s, i, embeds.Error("Error", "Nothing is playing"))
		return
	}

	session.Pause()
	
	track := session.Queue().Current()
	if track != nil {
		embed := embeds.NowPlaying(track, session)
		components := embeds.PlayerButtons(true)
		updateMessage(s, i, embed, components)
	}
}

func handlePlayerResume(s *discordgo.Session, i *discordgo.InteractionCreate, bot BotInterface) {
	session := bot.GetSession(i.GuildID)
	if session == nil || session.IsStopped() {
		respondComponent(s, i, embeds.Error("Error", "Nothing is playing"))
		return
	}

	session.Resume()
	
	track := session.Queue().Current()
	if track != nil {
		embed := embeds.NowPlaying(track, session)
		components := embeds.PlayerButtons(false)
		updateMessage(s, i, embed, components)
	}
}

func handlePlayerSkip(s *discordgo.Session, i *discordgo.InteractionCreate, bot BotInterface) {
	session := bot.GetSession(i.GuildID)
	if session == nil || session.IsStopped() {
		respondComponent(s, i, embeds.Error("Error", "Nothing is playing"))
		return
	}

	session.Skip()
	acknowledgeComponent(s, i)
}

func handlePlayerPrevious(s *discordgo.Session, i *discordgo.InteractionCreate, bot BotInterface) {
	session := bot.GetSession(i.GuildID)
	if session == nil || session.IsStopped() {
		respondComponent(s, i, embeds.Error("Error", "Nothing is playing"))
		return
	}

	if !session.Queue().HasPrevious() {
		respondComponent(s, i, embeds.Error("Error", "No previous track"))
		return
	}

	session.Queue().Previous()
	session.Skip()
	acknowledgeComponent(s, i)
}

func handlePlayerStop(s *discordgo.Session, i *discordgo.InteractionCreate, bot BotInterface) {
	session := bot.GetSession(i.GuildID)
	if session == nil {
		respondComponent(s, i, embeds.Error("Error", "Nothing is playing"))
		return
	}

	session.Stop()
	
	if vc := session.VoiceConnection(); vc != nil {
		vc.Disconnect()
	}
	
	bot.RemoveSession(i.GuildID)

	embed := embeds.Success("Stopped", "Playback stopped and queue cleared")
	updateMessage(s, i, embed, []discordgo.MessageComponent{})
}

func handlePlayerQueue(s *discordgo.Session, i *discordgo.InteractionCreate, bot BotInterface) {
	session := bot.GetSession(i.GuildID)
	if session == nil || session.Queue().IsEmpty() {
		respondComponent(s, i, embeds.Error("Error", "Queue is empty"))
		return
	}

	current := session.Queue().Current()
	upcoming := session.Queue().Upcoming()

	embed := embeds.Queue(current, upcoming, 1)
	
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

func respondComponent(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

func acknowledgeComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
}

func updateMessage(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed, components []discordgo.MessageComponent) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}

