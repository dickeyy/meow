package spotify

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/dickeyy/meow/internal/audio"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

var (
	trackRegex    = regexp.MustCompile(`spotify\.com\/track\/([a-zA-Z0-9]+)`)
	playlistRegex = regexp.MustCompile(`spotify\.com\/playlist\/([a-zA-Z0-9]+)`)
	albumRegex    = regexp.MustCompile(`spotify\.com\/album\/([a-zA-Z0-9]+)`)
)

type Client struct {
	client *spotify.Client
	ctx    context.Context
}

func NewClient(ctx context.Context, clientID, clientSecret string) *Client {
	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     spotifyauth.TokenURL,
	}

	token, err := config.Token(ctx)
	if err != nil {
		fmt.Printf("Failed to get Spotify token: %v\n", err)
		return nil
	}

	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)

	return &Client{
		client: client,
		ctx:    ctx,
	}
}

func (c *Client) IsSpotifyURL(url string) bool {
	return strings.Contains(url, "spotify.com")
}

func (c *Client) IsTrack(url string) bool {
	return trackRegex.MatchString(url)
}

func (c *Client) IsPlaylist(url string) bool {
	return playlistRegex.MatchString(url)
}

func (c *Client) IsAlbum(url string) bool {
	return albumRegex.MatchString(url)
}

func (c *Client) Extract(url string, requestedBy string) ([]*audio.Track, error) {
	if c.client == nil {
		return nil, fmt.Errorf("spotify client not initialized")
	}

	if c.IsTrack(url) {
		track, err := c.extractTrack(url, requestedBy)
		if err != nil {
			return nil, err
		}
		return []*audio.Track{track}, nil
	}

	if c.IsPlaylist(url) {
		return c.extractPlaylist(url, requestedBy)
	}

	if c.IsAlbum(url) {
		return c.extractAlbum(url, requestedBy)
	}

	return nil, fmt.Errorf("unsupported Spotify URL")
}

func (c *Client) extractTrack(url string, requestedBy string) (*audio.Track, error) {
	matches := trackRegex.FindStringSubmatch(url)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid track URL")
	}

	trackID := spotify.ID(matches[1])
	track, err := c.client.GetTrack(c.ctx, trackID)
	if err != nil {
		return nil, fmt.Errorf("failed to get track: %w", err)
	}

	return c.spotifyTrackToAudioTrack(track, requestedBy), nil
}

func (c *Client) extractPlaylist(url string, requestedBy string) ([]*audio.Track, error) {
	matches := playlistRegex.FindStringSubmatch(url)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid playlist URL")
	}

	playlistID := spotify.ID(matches[1])
	
	var tracks []*audio.Track
	offset := 0
	limit := 100

	for {
		items, err := c.client.GetPlaylistTracks(c.ctx, playlistID, spotify.Limit(limit), spotify.Offset(offset))
		if err != nil {
			return nil, fmt.Errorf("failed to get playlist tracks: %w", err)
		}

		for _, item := range items.Tracks {
			track := item.Track
			if track.ID != "" {
				tracks = append(tracks, c.playlistTrackToAudioTrack(&track, requestedBy))
			}
		}

		if len(items.Tracks) < limit {
			break
		}
		offset += limit
	}

	if len(tracks) == 0 {
		return nil, fmt.Errorf("no tracks found in playlist")
	}

	return tracks, nil
}

func (c *Client) extractAlbum(url string, requestedBy string) ([]*audio.Track, error) {
	matches := albumRegex.FindStringSubmatch(url)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid album URL")
	}

	albumID := spotify.ID(matches[1])
	album, err := c.client.GetAlbum(c.ctx, albumID)
	if err != nil {
		return nil, fmt.Errorf("failed to get album: %w", err)
	}

	var tracks []*audio.Track
	for _, item := range album.Tracks.Tracks {
		track := &audio.Track{
			ID:          string(item.ID),
			Title:       item.Name,
			Artist:      artistsToString(item.Artists),
			Album:       album.Name,
			Duration:    item.TimeDuration(),
			Source:      audio.SourceSpotify,
			RequestedBy: requestedBy,
		}

		if len(album.Images) > 0 {
			track.Thumbnail = album.Images[0].URL
		}

		tracks = append(tracks, track)
	}

	// Handle pagination for large albums
	if album.Tracks.Next != "" {
		offset := len(album.Tracks.Tracks)
		for {
			page, err := c.client.GetAlbumTracks(c.ctx, albumID, spotify.Limit(50), spotify.Offset(offset))
			if err != nil {
				break
			}

			for _, item := range page.Tracks {
				track := &audio.Track{
					ID:          string(item.ID),
					Title:       item.Name,
					Artist:      artistsToString(item.Artists),
					Album:       album.Name,
					Duration:    item.TimeDuration(),
					Source:      audio.SourceSpotify,
					RequestedBy: requestedBy,
				}

				if len(album.Images) > 0 {
					track.Thumbnail = album.Images[0].URL
				}

				tracks = append(tracks, track)
			}

			if page.Next == "" {
				break
			}
			offset += 50
		}
	}

	if len(tracks) == 0 {
		return nil, fmt.Errorf("no tracks found in album")
	}

	return tracks, nil
}

func (c *Client) spotifyTrackToAudioTrack(track *spotify.FullTrack, requestedBy string) *audio.Track {
	t := &audio.Track{
		ID:          string(track.ID),
		Title:       track.Name,
		Artist:      artistsToString(track.Artists),
		Album:       track.Album.Name,
		Duration:    track.TimeDuration(),
		Source:      audio.SourceSpotify,
		RequestedBy: requestedBy,
	}

	if len(track.Album.Images) > 0 {
		t.Thumbnail = track.Album.Images[0].URL
	}

	return t
}

func (c *Client) playlistTrackToAudioTrack(track *spotify.FullTrack, requestedBy string) *audio.Track {
	t := &audio.Track{
		ID:          string(track.ID),
		Title:       track.Name,
		Artist:      artistsToString(track.Artists),
		Album:       track.Album.Name,
		Duration:    track.TimeDuration(),
		Source:      audio.SourceSpotify,
		RequestedBy: requestedBy,
	}

	if len(track.Album.Images) > 0 {
		t.Thumbnail = track.Album.Images[0].URL
	}

	return t
}

func artistsToString(artists []spotify.SimpleArtist) string {
	names := make([]string, len(artists))
	for i, a := range artists {
		names[i] = a.Name
	}
	return strings.Join(names, ", ")
}

func (c *Client) GetSearchQuery(track *audio.Track) string {
	// Add "audio" to prefer audio-only uploads over music videos
	// YouTube's "Topic" channels have clean audio versions
	return fmt.Sprintf("%s %s audio", track.Artist, track.Title)
}

