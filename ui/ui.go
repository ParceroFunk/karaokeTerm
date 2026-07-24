package ui

import (
	"fmt"
	"log"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ParceroFunk/karaokeTerm/lrc"
	"github.com/ParceroFunk/karaokeTerm/mpris"
)

var (
	currentStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

type tickMsg time.Time

type Model struct {
	Player     *mpris.MPRIS
	Lines      []lrc.LyricLine
	currentIdx int
	termHeight int
}

func NewModel(player *mpris.MPRIS, lines []lrc.LyricLine) Model {
	return Model{Player: player, Lines: lines, termHeight: 20}
}

func (m Model) Init() tea.Cmd {
	return tick()
}

func tick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
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
		posUs, err := m.Player.GetPosition()
		if err != nil {
			log.Printf("position error: %v", err)
			return m, tick()
		}
		pos := time.Duration(posUs) * time.Microsecond
		m.currentIdx = findCurrentLine(m.Lines, pos)
		return m, tick()
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
	b.WriteString(fmt.Sprintf("\n[q to quit]"))
	return b.String()
}
