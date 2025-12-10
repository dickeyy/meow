package youtube

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/dickeyy/meow/internal/audio"
)

var (
	youtubeRegex  = regexp.MustCompile(`(?:youtube\.com\/(?:watch\?v=|playlist\?list=|embed\/)|youtu\.be\/)([a-zA-Z0-9_-]+)`)
	playlistRegex = regexp.MustCompile(`[?&]list=([a-zA-Z0-9_-]+)`)
)

const commandTimeout = 30 * time.Second

type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

type ytdlpOutput struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Uploader     string  `json:"uploader"`
	Channel      string  `json:"channel"`
	Duration     float64 `json:"duration"`
	Thumbnail    string  `json:"thumbnail"`
	WebpageURL   string  `json:"webpage_url"`
	URL          string  `json:"url"`
	ExtractorKey string  `json:"extractor_key"`
	Album        string  `json:"album"`
	Artist       string  `json:"artist"`
	Track        string  `json:"track"`
}

type ytdlpPlaylist struct {
	ID      string        `json:"id"`
	Title   string        `json:"title"`
	Entries []ytdlpOutput `json:"entries"`
}

func (e *Extractor) IsPlaylist(url string) bool {
	return playlistRegex.MatchString(url)
}

func (e *Extractor) IsYouTubeURL(url string) bool {
	return youtubeRegex.MatchString(url)
}

func (e *Extractor) Extract(url string, requestedBy string) ([]*audio.Track, error) {
	if e.IsPlaylist(url) {
		return e.extractPlaylist(url, requestedBy)
	}

	track, err := e.extractSingle(url, requestedBy)
	if err != nil {
		return nil, err
	}
	return []*audio.Track{track}, nil
}

func (e *Extractor) runCommand(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "yt-dlp", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	fmt.Printf("[yt-dlp] Running: yt-dlp %s\n", strings.Join(args, " "))

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("yt-dlp timed out after %v", commandTimeout)
	}
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return nil, fmt.Errorf("yt-dlp failed: %s", errMsg)
	}

	return stdout.Bytes(), nil
}

func (e *Extractor) extractSingle(url string, requestedBy string) (*audio.Track, error) {
	output, err := e.runCommand(
		"-j",
		"-f", "bestaudio[ext=m4a]/bestaudio/best",
		"--no-playlist",
		url,
	)
	if err != nil {
		return nil, err
	}

	var info ytdlpOutput
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("failed to parse yt-dlp output: %w (output: %s)", err, string(output))
	}

	return e.infoToTrack(&info, requestedBy), nil
}

func (e *Extractor) extractPlaylist(url string, requestedBy string) ([]*audio.Track, error) {
	output, err := e.runCommand(
		"-j",
		"--flat-playlist",
		url,
	)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	tracks := make([]*audio.Track, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		var info ytdlpOutput
		if err := json.Unmarshal([]byte(line), &info); err != nil {
			continue
		}

		track := &audio.Track{
			ID:          info.ID,
			Title:       info.Title,
			Artist:      info.Uploader,
			URL:         info.WebpageURL,
			Source:      audio.SourceYouTube,
			RequestedBy: requestedBy,
		}

		if track.URL == "" && info.ID != "" {
			track.URL = "https://www.youtube.com/watch?v=" + info.ID
		}

		tracks = append(tracks, track)
	}

	if len(tracks) == 0 {
		return nil, fmt.Errorf("no tracks found in playlist")
	}

	return tracks, nil
}

func (e *Extractor) GetStreamURL(track *audio.Track) (string, error) {
	url := track.URL
	if url == "" && track.ID != "" {
		url = "https://www.youtube.com/watch?v=" + track.ID
	}

	fmt.Printf("[yt-dlp] Getting stream URL for: %s\n", url)

	output, err := e.runCommand(
		"-f", "bestaudio[ext=m4a]/bestaudio/best",
		"-g",
		"--no-playlist",
		url,
	)
	if err != nil {
		return "", err
	}

	streamURL := strings.TrimSpace(string(output))
	if streamURL == "" {
		return "", fmt.Errorf("no stream URL found")
	}

	fmt.Printf("[yt-dlp] Got stream URL (length: %d)\n", len(streamURL))

	if track.Duration == 0 {
		e.enrichTrackMetadata(track)
	}

	return streamURL, nil
}

func (e *Extractor) enrichTrackMetadata(track *audio.Track) {
	url := track.URL
	if url == "" && track.ID != "" {
		url = "https://www.youtube.com/watch?v=" + track.ID
	}

	output, err := e.runCommand(
		"-j",
		"--no-playlist",
		url,
	)
	if err != nil {
		fmt.Printf("[yt-dlp] Failed to enrich metadata: %v\n", err)
		return
	}

	var info ytdlpOutput
	if err := json.Unmarshal(output, &info); err != nil {
		return
	}

	if track.Title == "" {
		track.Title = info.Title
	}
	if track.Artist == "" {
		track.Artist = info.Uploader
		if info.Artist != "" {
			track.Artist = info.Artist
		}
	}
	if track.Duration == 0 && info.Duration > 0 {
		track.Duration = time.Duration(info.Duration) * time.Second
	}
	if track.Thumbnail == "" {
		track.Thumbnail = info.Thumbnail
	}
	if track.Album == "" {
		track.Album = info.Album
	}
}

func (e *Extractor) Search(query string, requestedBy string) (*audio.Track, error) {
	fmt.Printf("[yt-dlp] Searching for: %s\n", query)

	// Use YouTube Music search for better audio-only results
	// Falls back to regular YouTube search if ytmusic fails
	output, err := e.runCommand(
		"-j",
		"-f", "bestaudio[ext=m4a]/bestaudio/best",
		"--no-playlist",
		"--default-search", "ytsearch",
		"--match-filter", "!is_live",
		query,
	)
	if err != nil {
		return nil, err
	}

	var info ytdlpOutput
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("failed to parse yt-dlp output: %w", err)
	}

	fmt.Printf("[yt-dlp] Found: %s by %s\n", info.Title, info.Uploader)

	return e.infoToTrack(&info, requestedBy), nil
}

func (e *Extractor) infoToTrack(info *ytdlpOutput, requestedBy string) *audio.Track {
	artist := info.Uploader
	if info.Artist != "" {
		artist = info.Artist
	}
	if info.Channel != "" && artist == "" {
		artist = info.Channel
	}

	title := info.Title
	if info.Track != "" {
		title = info.Track
	}

	return &audio.Track{
		ID:          info.ID,
		Title:       title,
		Artist:      artist,
		Album:       info.Album,
		Duration:    time.Duration(info.Duration) * time.Second,
		URL:         info.WebpageURL,
		StreamURL:   info.URL,
		Thumbnail:   info.Thumbnail,
		Source:      audio.SourceYouTube,
		RequestedBy: requestedBy,
	}
}
