package main

import (
	"log"
	"strconv"

	"github.com/ParceroFunk/karaokeTerm/lrc"
	"github.com/ParceroFunk/karaokeTerm/lrclib"
	"github.com/ParceroFunk/karaokeTerm/mpris"
	"github.com/ParceroFunk/karaokeTerm/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/godbus/dbus/v5"
)

func main() {
	// Initialize read from dbus for playing media
	conn, err := dbus.SessionBus()
	if err != nil {
		log.Fatalf("dbus connection failed: %v", err)
	}
	defer conn.Close()

	// Get metadata from mpris package
	player := mpris.NewMPRIS(conn)
	title, artist, duration := getPlayingMediaMetadata(player)

	// Get the Lyrics from lrclib package
	lyrics, err := lrclib.GetLyrics(title, artist, duration)
	if err != nil {
		log.Fatalf("dbus connection failed: %v", err)
	}
	// fmt.Println(lyrics)

	// Parse Lyrics to sync them
	lyricsLines := lrc.LrcLyricsParser(lyrics)

	p := tea.NewProgram(ui.NewModel(player, lyricsLines), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("ui error: %v", err)
	}
}

func getPlayingMediaMetadata(player *mpris.MPRIS) (title, artist, length string) {
	mediaTitle, err := player.GetTitle()
	if err != nil {
		log.Fatalf("could not get title: %v", err)
	}
	log.Printf("Currently playing song: %s", mediaTitle)

	mediaArtist, err := player.GetArtist()
	if err != nil {
		log.Fatalf("could not get artist: %v", err)
	}
	log.Printf("Artist: %s", mediaArtist)

	mediaLength, err := player.GetDuration()
	if err != nil {
		log.Fatalf("could not get duration: %v", err)
	}
	log.Printf("Duration of currently playing song: %d", mediaLength)

	return mediaTitle, mediaArtist, strconv.FormatInt(mediaLength, 10)
}
