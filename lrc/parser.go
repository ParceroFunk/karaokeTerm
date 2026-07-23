package lrc

import (
	"log"
	"strings"
	"time"
)

type LyricLine struct {
	Timestamp time.Duration
	Verse     string
}

// LrcLyricsParser returns a slice of the pair of timestamp and
// lyric line. Then gets processed for sync with media.
func LrcLyricsParser(lyrics string) []LyricLine {
	lines := strings.Split(lyrics, "\n")
	var lyricsLines []LyricLine

	// Example line: [00:00.96] Two, three, four
	// Bracketed timestamp is 10 chars: '[', 8 timestamp chars, ']'
	const splitIndex = 10
	for _, line := range lines {
		if len(line) <= splitIndex {
			if strings.TrimSpace(line) != "" {
				log.Printf("[WARNING] verse %q is too short for parsing", line)
			}
			continue
		}

		// Strip the surrounding brackets: line[1:9] -> "00:00.96"
		timestamp := line[1 : splitIndex-1]
		verse := strings.TrimSpace(line[splitIndex+1:])

		lyricsLines = append(lyricsLines, newLyricLine(timestamp, verse))
	}

	return lyricsLines
}

func newLyricLine(timestamp, verse string) LyricLine {
	// Handle timestamp parsing for Go time package.
	// "00:00.96" -> "00m00.96s" -- time.ParseDuration understands
	// fractional seconds natively, so only ":" needs replacing.
	timeForGo := strings.Replace(timestamp, ":", "m", 1) + "s"

	elapsed, err := time.ParseDuration(timeForGo)
	if err != nil {
		log.Printf("[WARNING] failed to parse timestamp %q: %v", timestamp, err)
	}

	return LyricLine{
		Timestamp: elapsed,
		Verse:     verse,
	}
}
