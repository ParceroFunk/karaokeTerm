package ui

import (
	"log"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ParceroFunk/karaokeTerm/lrc"
	"github.com/ParceroFunk/karaokeTerm/lrclib"
	"github.com/ParceroFunk/karaokeTerm/mpris"
)

var (
	currentStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

type tickMsg time.Time

// lyricsMsg is returned by fetchLyricsCmd once the async lrclib fetch + parse completes.
type lyricsMsg struct {
	identity string // identity of the track this fetch was for, to discard stale results
	lines    []lrc.LyricLine
	err      error
}

type Model struct {
	Player     *mpris.MPRIS
	Lines      []lrc.LyricLine
	currentIdx int
	termHeight int
	identity   string // current track identity, see trackIdentity()
	loading    bool
}

func NewModel(player *mpris.MPRIS, lines []lrc.LyricLine, meta mpris.TrackMetadata) Model {
	return Model{
		Player:     player,
		Lines:      lines,
		termHeight: 20,
		identity:   trackIdentity(meta),
	}
}

// trackIdentity builds a stable identifier for "is this a different track".
// mpris:trackid is unreliable on some players (notably browser-based MPRIS
// bridges like Chromium, which often reuse the same trackid for every track
// in a session) so title+artist is used as the primary signal, with trackID
// appended as a tiebreaker for players that do implement it correctly.
func trackIdentity(meta mpris.TrackMetadata) string {
	return meta.Title + "::" + meta.Artist + "::" + meta.TrackID
}

func (m Model) Init() tea.Cmd {
	return tick()
}

func tick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// fetchLyricsCmd fetches and parses lyrics for a track asynchronously so the
// UI loop isn't blocked on the lrclib HTTP call.
func fetchLyricsCmd(identity, title, artist string, duration int64) tea.Cmd {
	return func() tea.Msg {
		lyrics, err := lrclib.GetLyrics(title, artist, strconv.FormatInt(duration, 10))
		if err != nil {
			return lyricsMsg{identity: identity, err: err}
		}
		return lyricsMsg{identity: identity, lines: lrc.LrcLyricsParser(lyrics)}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termHeight = msg.Height
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case tickMsg:
		meta, err := m.Player.GetMetadata()
		if err != nil {
			log.Printf("metadata error: %v", err)
			return m, tick()
		}

		newIdentity := trackIdentity(meta)

		// Track changed: reset state and kick off an async lyrics refetch.
		if newIdentity != m.identity {
			log.Printf("track changed: %q -> %q", m.identity, newIdentity)
			m.identity = newIdentity
			m.Lines = nil
			m.currentIdx = 0
			m.loading = true
			return m, tea.Batch(fetchLyricsCmd(newIdentity, meta.Title, meta.Artist, meta.Duration), tick())
		}

		posUs, err := m.Player.GetPosition()
		if err != nil {
			log.Printf("position error: %v", err)
			return m, tick()
		}
		pos := time.Duration(posUs) * time.Microsecond
		m.currentIdx = findCurrentLine(m.Lines, pos)
		return m, tick()

	case lyricsMsg:
		// Discard results for a track we've since moved on from (e.g. rapid
		// skips while a slow lrclib request was in flight).
		if msg.identity != m.identity {
			return m, nil
		}
		m.loading = false
		if msg.err != nil {
			log.Printf("lyrics fetch error: %v", msg.err)
			return m, nil
		}
		m.Lines = msg.lines
		m.currentIdx = 0
		return m, nil
	}
	return m, nil
}

// findCurrentLine returns the index of the last line whose Timestamp <= pos.
func findCurrentLine(lines []lrc.LyricLine, pos time.Duration) int {
	idx := 0
	for i, l := range lines {
		if l.Timestamp <= pos {
			idx = i
		} else {
			break
		}
	}
	return idx
}

func (m Model) View() string {
	if m.loading {
		return "Loading lyrics...\n"
	}
	if len(m.Lines) == 0 {
		return "No lyrics loaded.\n"
	}

	half := m.termHeight / 2
	start := m.currentIdx - half
	end := m.currentIdx + half

	var b strings.Builder
	for i := start; i < end; i++ {
		if i < 0 || i >= len(m.Lines) {
			b.WriteString("\n")
			continue
		}
		if i == m.currentIdx {
			b.WriteString(currentStyle.Render(m.Lines[i].Verse) + "\n")
		} else {
			b.WriteString(dimStyle.Render(m.Lines[i].Verse) + "\n")
		}
	}
	b.WriteString("\n[q to quit]")
	return b.String()
}
