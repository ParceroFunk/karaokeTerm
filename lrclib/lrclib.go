package lrclib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// LyricResponse mirrors the LRCLIB API schema
type LyricResponse struct {
	ID           int     `json:"id"`
	Name         int     `json:"name"`
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    string  `json:"albumName"`
	Duration     float64 `json:"duration"`
	Instrumental bool    `json:"instrumental"`
	PlainLyrics  string  `json:"plainLyrics"`
	SyncedLyrics string  `json:"syncedLyrics"`
}

func main() {
	// 1. Build Query Parameters
	baseURL := "https://lrclib.net"
	params := url.Values{}
	params.Add("track_name", "The Chain")
	params.Add("duration", "270") // Duration in seconds is critical

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// 2. Prepare HTTP Request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		panic(err)
	}

	// Required: Set a unique User-Agent identifier
	req.Header.Set("User-Agent", "MyGoLyricApp/1.0.0 (https://github.com)")

	// 3. Execute Request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// 4. Handle Missing Track
	if resp.StatusCode == http.StatusNotFound {
		fmt.Println("Lyrics not found for this track signature.")
		return
	}

	// 5. Parse JSON Output
	var lyrics LyricResponse
	if err := json.NewDecoder(resp.Body).Decode(&lyrics); err != nil {
		panic(err)
	}

	// 6. Print Results
	fmt.Printf("Found: %s - %s\n\n", lyrics.ArtistName, lyrics.TrackName)
	if lyrics.Instrumental {
		fmt.Println("[Instrumental Track]")
	} else if lyrics.SyncedLyrics != "" {
		fmt.Println("--- Synced Lyrics (.lrc Format) ---")
		fmt.Println(lyrics.SyncedLyrics)
	} else {
		fmt.Println("--- Plain Lyrics ---")
		fmt.Println(lyrics.PlainLyrics)
	}
}
