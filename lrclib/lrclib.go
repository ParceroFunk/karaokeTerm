package lrclib

import (
	"encoding/json"
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

func GetLyrics(title, artist, duration string) (string, error) {
	// 1. Build Query Parameters
	// Correct base is https://lrclib.net/api — /get for exact match, /search for fuzzy.
	// baseURL := "https://lrclib.net/api/get"
	params := url.Values{}
	params.Add("track_name", title)
	params.Add("artist_name", artist)
	params.Add("duration", duration) // Duration in seconds is critical

	fullURL := baseURL + "?" + params.Encode()
	log.Printf("Requesting lyrics: %s", fullURL)

	// 2. Prepare HTTP Request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		log.Fatalf("failed to build request: %v", err)
		return "", err
	}

	// Required: Set a unique User-Agent identifier
	req.Header.Set("User-Agent", "MyGoLyricApp/1.0.0 (https://github.com)")

	// 3. Execute Request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("request to lrclib failed: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	log.Printf("lrclib response status: %s", resp.Status)

	// Read the raw body first so we can log it on decode failure
	// instead of losing it after a failed Decode() panic.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed to read response body: %v", err)
		return "", err
	}

	// 4. Handle Missing Track
	if resp.StatusCode == http.StatusNotFound {
		log.Printf("Lyrics not found for this track signature. Body: %s", body)
		return "", nil
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("lrclib returned unexpected status: %s, body: %s", resp.Status, body)
		return "", err
	}

	// 5. Parse JSON Output
	var lyrics LyricResponse
	if err = json.Unmarshal(body, &lyrics); err != nil {
		log.Fatalf("failed to decode lrclib response: %v\nraw body: %s", err, body)
		return "", err
	}

	// 6. Return Results
	log.Printf("Found: %s - %s", lyrics.ArtistName, lyrics.TrackName)
	if lyrics.Instrumental {
		log.Println("[Instrumental Track]")
		return "[Instrumental Track]", nil
	} else if lyrics.SyncedLyrics != "" {
		log.Println("--- Synced Lyrics (.lrc Format) ---")
		return lyrics.SyncedLyrics, nil
	} else {
		log.Println("--- Plain Lyrics ---")
		return lyrics.PlainLyrics, nil
	}
}
