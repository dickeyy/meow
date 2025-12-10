package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dickeyy/meow/internal/audio"
	"github.com/dickeyy/meow/internal/embeds"
)

func handlePlay(s *discordgo.Session, i *discordgo.InteractionCreate, bot BotInterface) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		fmt.Printf("[play] Failed to defer response: %v\n", err)
		return
	}

	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		respondError(s, i, "Please provide a URL or search query")
		return
	}

	query := options[0].StringValue()
	userID := i.Member.User.ID

	fmt.Printf("[play] Query: %s\n", query)

	voiceState, err := findUserVoiceState(s, i.GuildID, userID)
	if err != nil || voiceState == nil {
		respondError(s, i, "You need to be in a voice channel to play music")
		return
	}

	fmt.Printf("[play] User is in voice channel: %s\n", voiceState.ChannelID)

	session := bot.GetOrCreateSession(i.GuildID)

	vc := session.VoiceConnection()
	if vc == nil || vc.ChannelID != voiceState.ChannelID {
		fmt.Printf("[play] Joining voice channel...\n")
		vc, err = s.ChannelVoiceJoin(i.GuildID, voiceState.ChannelID, false, true)
		if err != nil {
			respondError(s, i, "Failed to join voice channel: "+err.Error())
			return
		}
		session.SetVoiceConnection(vc)
		session.SetChannelID(i.ChannelID)
		fmt.Printf("[play] Joined voice channel\n")
	}

	var tracks []*audio.Track

	if bot.Spotify() != nil && bot.Spotify().IsSpotifyURL(query) {
		fmt.Printf("[play] Extracting Spotify URL...\n")
		tracks, err = bot.Spotify().Extract(query, userID)
		if err != nil {
			fmt.Printf("[play] Spotify extract failed: %v\n", err)
			respondError(s, i, "Failed to get Spotify tracks: "+err.Error())
			return
		}
	} else if bot.YouTube().IsYouTubeURL(query) || strings.HasPrefix(query, "http") {
		fmt.Printf("[play] Extracting URL...\n")
		tracks, err = bot.YouTube().Extract(query, userID)
		if err != nil {
			fmt.Printf("[play] Extract failed: %v\n", err)
			respondError(s, i, "Failed to extract tracks: "+err.Error())
			return
		}
		// Try to get better artwork from iTunes for YouTube tracks
		for _, track := range tracks {
			fetchArtwork(bot, track)
		}
	} else {
		fmt.Printf("[play] Searching YouTube for: %s\n", query)
		track, err := bot.YouTube().Search(query, userID)
		if err != nil {
			fmt.Printf("[play] Search failed: %v\n", err)
			respondError(s, i, "Search failed: "+err.Error())
			return
		}
		// Try to get better artwork from iTunes
		fetchArtwork(bot, track)
		tracks = []*audio.Track{track}
	}

	if len(tracks) == 0 {
		respondError(s, i, "No tracks found")
		return
	}

	fmt.Printf("[play] Found %d tracks\n", len(tracks))

	wasEmpty := session.Queue().IsEmpty()
	session.Queue().Add(tracks...)

	if wasEmpty {
		firstTrack := session.Queue().Current()
		fmt.Printf("[play] First track: %s\n", firstTrack.Title)

		// Store original thumbnail (from Spotify) before YouTube search
		originalThumbnail := firstTrack.Thumbnail

		if firstTrack.Source == audio.SourceSpotify {
			searchQuery := bot.Spotify().GetSearchQuery(firstTrack)
			fmt.Printf("[play] Searching YouTube for Spotify track: %s\n", searchQuery)
			ytTrack, err := bot.YouTube().Search(searchQuery, userID)
			if err != nil {
				fmt.Printf("[play] YouTube search for Spotify track failed: %v\n", err)
				respondError(s, i, "Failed to find track on YouTube: "+err.Error())
				return
			}
			firstTrack.StreamURL = ytTrack.StreamURL
			// Keep the original Spotify thumbnail, don't use YouTube's
			if firstTrack.Duration == 0 {
				firstTrack.Duration = ytTrack.Duration
			}
		}

		// Restore original thumbnail if it was set (Spotify)
		if originalThumbnail != "" {
			firstTrack.Thumbnail = originalThumbnail
		}

		if firstTrack.StreamURL == "" {
			fmt.Printf("[play] Getting stream URL...\n")
			streamURL, err := bot.YouTube().GetStreamURL(firstTrack)
			if err != nil {
				fmt.Printf("[play] Failed to get stream URL: %v\n", err)
				respondError(s, i, "Failed to get stream URL: "+err.Error())
				return
			}
			firstTrack.StreamURL = streamURL
		}

		fmt.Printf("[play] Got stream URL, starting playback...\n")

		session.OnTrackChange = func(track *audio.Track) {
			fmt.Printf("[player] Track changed to: %s\n", track.Title)
			
			// Store original thumbnail before YouTube search
			originalThumb := track.Thumbnail
			
			if track.Source == audio.SourceSpotify && track.StreamURL == "" {
				searchQuery := bot.Spotify().GetSearchQuery(track)
				ytTrack, err := bot.YouTube().Search(searchQuery, userID)
				if err == nil {
					track.StreamURL = ytTrack.StreamURL
					// Don't overwrite Spotify thumbnail
					if track.Duration == 0 {
						track.Duration = ytTrack.Duration
					}
				}
				// Restore Spotify thumbnail
				if originalThumb != "" {
					track.Thumbnail = originalThumb
				}
			} else if track.StreamURL == "" {
				streamURL, _ := bot.YouTube().GetStreamURL(track)
				track.StreamURL = streamURL
			}

			// For non-Spotify tracks without artwork, try iTunes
			if track.Thumbnail == "" || track.Source != audio.SourceSpotify {
				fetchArtwork(bot, track)
			}

			sendNowPlayingEmbed(s, session.ChannelID(), track, session)
		}

		player := audio.NewPlayer()
		go func() {
			if err := player.Play(session, s); err != nil {
				fmt.Printf("[player] Playback error: %v\n", err)
			}
		}()

		// Delete the deferred response since we'll send the Now Playing embed from OnTrackChange
		s.InteractionResponseDelete(i.Interaction)
	} else {
		var content string
		if len(tracks) == 1 {
			content = fmt.Sprintf("Added to queue: **%s**", tracks[0].Title)
		} else {
			content = fmt.Sprintf("Added **%d** tracks to queue", len(tracks))
		}

		embed := embeds.Success("Queue Updated", content)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{embed},
		})
	}
}

func fetchArtwork(bot BotInterface, track *audio.Track) {
	if bot.Artwork() == nil {
		return
	}
	
	// Only fetch if we don't already have good artwork (Spotify)
	if track.Source == audio.SourceSpotify && track.Thumbnail != "" {
		return
	}

	artworkURL, err := bot.Artwork().GetAlbumArt(track.Artist, track.Title)
	if err != nil {
		fmt.Printf("[artwork] Failed to get artwork for %s: %v\n", track.Title, err)
		return
	}
	
	fmt.Printf("[artwork] Got iTunes artwork for: %s\n", track.Title)
	track.Thumbnail = artworkURL
}

func findUserVoiceState(s *discordgo.Session, guildID, userID string) (*discordgo.VoiceState, error) {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		return nil, err
	}

	for _, vs := range guild.VoiceStates {
		if vs.UserID == userID {
			return vs, nil
		}
	}

	return nil, nil
}

func respondError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	fmt.Printf("[play] Error: %s\n", message)
	embed := embeds.Error("Error", message)
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
}

func sendNowPlayingEmbed(s *discordgo.Session, channelID string, track *audio.Track, session *audio.Session) {
	embed := embeds.NowPlaying(track, session)
	components := embeds.PlayerButtons(session.IsPaused())

	s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
}
