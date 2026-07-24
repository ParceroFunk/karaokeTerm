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
	conn, err := dbus.SessionBus()
	if err != nil {
		log.Fatalf("dbus connection failed: %v", err)
	}
	defer conn.Close()

	player := mpris.NewMPRIS(conn)

	meta, err := player.GetMetadata()
	if err != nil {
		log.Fatalf("could not get metadata: %v", err)
	}
	log.Printf("Currently playing song: %s", meta.Title)
	log.Printf("Artist: %s", meta.Artist)
	log.Printf("Duration: %d", meta.Duration)
	log.Printf("TrackID: %s", meta.TrackID)

	lyrics, err := lrclib.GetLyrics(meta.Title, meta.Artist, strconv.FormatInt(meta.Duration, 10))
	if err != nil {
		log.Fatalf("could not fetch lyrics: %v", err)
	}

	lyricsLines := lrc.LrcLyricsParser(lyrics)

	p := tea.NewProgram(ui.NewModel(player, lyricsLines, meta), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("ui error: %v", err)
	}
}
