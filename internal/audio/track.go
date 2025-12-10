package audio

import "time"

type TrackSource string

const (
	SourceYouTube TrackSource = "youtube"
	SourceSpotify TrackSource = "spotify"
	SourceDirect  TrackSource = "direct"
	SourceUnknown TrackSource = "unknown"
)

type Track struct {
	ID          string
	Title       string
	Artist      string
	Album       string
	Duration    time.Duration
	URL         string        // Original URL
	StreamURL   string        // Direct stream URL (from yt-dlp)
	Thumbnail   string        // Album art / thumbnail URL
	Source      TrackSource
	RequestedBy string        // User ID who requested the track
	PlaylistID  string        // If part of a playlist
}

func (t *Track) FormatDuration() string {
	if t.Duration == 0 {
		return "Live"
	}

	total := int(t.Duration.Seconds())
	hours := total / 3600
	minutes := (total % 3600) / 60
	seconds := total % 60

	if hours > 0 {
		return formatTime(hours, minutes, seconds)
	}
	return formatTimeMinSec(minutes, seconds)
}

func formatTime(h, m, s int) string {
	return pad(h) + ":" + pad(m) + ":" + pad(s)
}

func formatTimeMinSec(m, s int) string {
	return pad(m) + ":" + pad(s)
}

func pad(n int) string {
	if n < 10 {
		return "0" + itoa(n)
	}
	return itoa(n)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}

