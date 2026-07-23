package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ParceroFunk/karaokeTerm/mpris"
	"github.com/godbus/dbus/v5"
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
	// Initialize read from dbus for playing media
	conn, err := dbus.SessionBus()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	player := mpris.NewMPRIS(conn)

	title, err := player.GetTitle()
	if err != nil {
		log.Fatalf("could not get title: %v", err)
	}
	fmt.Println("Currently playing song:", title)

	length, err := player.GetDuration()
	if err != nil {
		log.Fatalf("could not get title: %v", err)
	}
	fmt.Println("Duration of currently playing song:", length)

	// 1. Build Query Parameters
	baseURL := "https://lrclib.net/Search/"
	params := url.Values{}
	params.Add("track_name", title)
	params.Add("duration", strconv.FormatInt(length, 10)) // Duration in seconds is critical

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
