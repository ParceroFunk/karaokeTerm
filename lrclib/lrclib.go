package lrclib

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

// LyricResponse mirrors the LRCLIB API schema
type LyricResponse struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    string  `json:"albumName"`
	Duration     float64 `json:"duration"`
	Instrumental bool    `json:"instrumental"`
	PlainLyrics  string  `json:"plainLyrics"`
	SyncedLyrics string  `json:"syncedLyrics"`
}

const baseURL = "https://lrclib.net/api/get"

// GetLyricsData fetches the full LRCLIB response (needed for synced parsing later).
func GetLyricsData(title, artist, duration string) (*LyricResponse, error) {
	params := url.Values{}
	params.Add("track_name", title)
	params.Add("artist_name", artist)
	params.Add("duration", duration)

	fullURL := baseURL + "?" + params.Encode()
	log.Printf("Requesting lyrics: %s", fullURL)

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("User-Agent", "MyGoLyricApp/1.0.0 (https://github.com)")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to lrclib failed: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("lrclib response status: %s", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("Lyrics not found for this track signature. Body: %s", body)
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lrclib returned unexpected status: %s, body: %s", resp.Status, body)
	}

	var lyrics LyricResponse
	if err := json.Unmarshal(body, &lyrics); err != nil {
		return nil, fmt.Errorf("failed to decode lrclib response: %w, raw body: %s", err, body)
	}

	log.Printf("Found: %s - %s", lyrics.ArtistName, lyrics.TrackName)
	return &lyrics, nil
}

// GetLyrics keeps the old string-only interface for convenience.
func GetLyrics(title, artist, duration string) (string, error) {
	lyrics, err := GetLyricsData(title, artist, duration)
	if err != nil {
		return "", err
	}
	if lyrics == nil {
		return "", nil
	}
	if lyrics.Instrumental {
		return "[Instrumental Track]", nil
	}
	if lyrics.SyncedLyrics != "" {
		return lyrics.SyncedLyrics, nil
	}
	return lyrics.PlainLyrics, nil
}
