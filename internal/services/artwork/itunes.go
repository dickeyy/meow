package artwork

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ITunesClient struct {
	client *http.Client
}

type itunesResponse struct {
	ResultCount int            `json:"resultCount"`
	Results     []itunesResult `json:"results"`
}

type itunesResult struct {
	ArtworkUrl100 string `json:"artworkUrl100"`
	ArtworkUrl60  string `json:"artworkUrl60"`
	ArtworkUrl30  string `json:"artworkUrl30"`
	TrackName     string `json:"trackName"`
	ArtistName    string `json:"artistName"`
	CollectionName string `json:"collectionName"`
}

func NewITunesClient() *ITunesClient {
	return &ITunesClient{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetAlbumArt searches iTunes for album artwork and returns a high-res URL
func (c *ITunesClient) GetAlbumArt(artist, title string) (string, error) {
	// Build search query
	query := fmt.Sprintf("%s %s", artist, title)
	searchURL := fmt.Sprintf(
		"https://itunes.apple.com/search?term=%s&media=music&entity=song&limit=1",
		url.QueryEscape(query),
	)

	resp, err := c.client.Get(searchURL)
	if err != nil {
		return "", fmt.Errorf("iTunes request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("iTunes returned status %d", resp.StatusCode)
	}

	var result itunesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse iTunes response: %w", err)
	}

	if result.ResultCount == 0 || len(result.Results) == 0 {
		return "", fmt.Errorf("no results found")
	}

	// Get the artwork URL and upscale it to 600x600
	artworkURL := result.Results[0].ArtworkUrl100
	if artworkURL == "" {
		return "", fmt.Errorf("no artwork URL in result")
	}

	// iTunes returns 100x100 by default, we can change it to get higher res
	// Replace "100x100" with "600x600" for better quality
	highResURL := strings.Replace(artworkURL, "100x100", "600x600", 1)

	return highResURL, nil
}

