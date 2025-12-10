package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dickeyy/meow/internal/embeds"
)

func handleQueue(s *discordgo.Session, i *discordgo.InteractionCreate, bot BotInterface) {
	session := bot.GetSession(i.GuildID)
	
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		respond(s, i, embeds.Error("Error", "Please specify a subcommand"))
		return
	}

	subCmd := options[0]

	switch subCmd.Name {
	case "view":
		handleQueueView(s, i, bot, session)
	case "move":
		handleQueueMove(s, i, bot, session, subCmd.Options)
	case "remove":
		handleQueueRemove(s, i, bot, session, subCmd.Options)
	case "clear":
		handleQueueClear(s, i, bot, session)
	}
}

func handleQueueView(s *discordgo.Session, i *discordgo.InteractionCreate, bot BotInterface, session interface{}) {
	audioSession := bot.GetSession(i.GuildID)
	if audioSession == nil || audioSession.Queue().IsEmpty() {
		respond(s, i, embeds.Info("Queue", "The queue is empty"))
		return
	}

	current := audioSession.Queue().Current()
	upcoming := audioSession.Queue().Upcoming()

	embed := embeds.Queue(current, upcoming, 1)
	respond(s, i, embed)
}

func handleQueueMove(s *discordgo.Session, i *discordgo.InteractionCreate, bot BotInterface, session interface{}, options []*discordgo.ApplicationCommandInteractionDataOption) {
	audioSession := bot.GetSession(i.GuildID)
	if audioSession == nil || audioSession.Queue().IsEmpty() {
		respond(s, i, embeds.Error("Error", "The queue is empty"))
		return
	}

	var from, to int64
	for _, opt := range options {
		switch opt.Name {
		case "from":
			from = opt.IntValue()
		case "to":
			to = opt.IntValue()
		}
	}

	if !audioSession.Queue().Move(int(from), int(to)) {
		respond(s, i, embeds.Error("Error", "Invalid positions. Make sure both positions are within the queue range."))
		return
	}

	respond(s, i, embeds.Success("Queue Updated", fmt.Sprintf("Moved track from position %d to %d", from, to)))
}

func handleQueueRemove(s *discordgo.Session, i *discordgo.InteractionCreate, bot BotInterface, session interface{}, options []*discordgo.ApplicationCommandInteractionDataOption) {
	audioSession := bot.GetSession(i.GuildID)
	if audioSession == nil || audioSession.Queue().IsEmpty() {
		respond(s, i, embeds.Error("Error", "The queue is empty"))
		return
	}

	var position int64
	for _, opt := range options {
		if opt.Name == "position" {
			position = opt.IntValue()
		}
	}

	removed := audioSession.Queue().Remove(int(position))
	if removed == nil {
		respond(s, i, embeds.Error("Error", "Invalid position. Make sure the position is within the queue range."))
		return
	}

	respond(s, i, embeds.Success("Track Removed", fmt.Sprintf("Removed **%s** from the queue", removed.Title)))
}

func handleQueueClear(s *discordgo.Session, i *discordgo.InteractionCreate, bot BotInterface, session interface{}) {
	audioSession := bot.GetSession(i.GuildID)
	if audioSession == nil || audioSession.Queue().IsEmpty() {
		respond(s, i, embeds.Info("Queue", "The queue is already empty"))
		return
	}

	// Keep the current track, clear the rest
	current := audioSession.Queue().Current()
	audioSession.Queue().Clear()
	if current != nil {
		audioSession.Queue().Add(current)
	}

	respond(s, i, embeds.Success("Queue Cleared", "Upcoming tracks have been removed"))
}

