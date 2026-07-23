// Package mpris manages the Linux player call for current playing media
// Properties are on the following link: https://www.freedesktop.org/wiki/Specifications/mpris-spec/metadata/?__goaway_challenge=meta-refresh&__goaway_id=89c89904e1313541ad6814801c76ffac&__goaway_referer=https%3A%2F%2Fwww.google.com%2F
package mpris

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/godbus/dbus/v5"
)

const DivFactorForLength = 1e6

type MPRIS struct {
	conn       *dbus.Conn
	PlayerName string // cached, avoids re-listing D-Bus names every call
}

func NewMPRIS(conn *dbus.Conn) *MPRIS {
	return &MPRIS{conn: conn}
}

// findPlayer lists D-Bus names and returns the first MPRIS player found.
func (m *MPRIS) findPlayer() (string, error) {
	var names []string
	if err := m.conn.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&names); err != nil {
		return "", fmt.Errorf("failed to list dbus names: %w", err)
	}

	for _, n := range names {
		if strings.HasPrefix(n, "org.mpris.MediaPlayer2.") {
			log.Printf("MPRIS player found: %s", n)
			return n, nil
		}
	}

	return "", errors.New("no MPRIS player found")
}

// ensurePlayer returns a cached player name, re-resolving if not set.
func (m *MPRIS) ensurePlayer() (string, error) {
	if m.PlayerName != "" {
		return m.PlayerName, nil
	}
	name, err := m.findPlayer()
	if err != nil {
		return "", err
	}
	m.PlayerName = name
	return name, nil
}

func (m *MPRIS) getProperty(name string) (dbus.Variant, error) {
	playerName, err := m.ensurePlayer()
	if err != nil {
		return dbus.Variant{}, err
	}

	obj := m.conn.Object(playerName, dbus.ObjectPath("/org/mpris/MediaPlayer2"))
	val, err := obj.GetProperty(name)
	if err != nil {
		// Player likely closed/changed; drop cache so next call re-resolves.
		m.PlayerName = ""
		return dbus.Variant{}, fmt.Errorf("failed to get property %s: %w", name, err)
	}
	return val, nil
}

// GetStatus returns the playback status (Playing, Paused, Stopped).
func (m *MPRIS) GetStatus() (string, error) {
	v, err := m.getProperty("org.mpris.MediaPlayer2.Player.PlaybackStatus")
	if err != nil {
		return "", err
	}
	status, ok := v.Value().(string)
	if !ok {
		return "", errors.New("unexpected type for PlaybackStatus")
	}
	return status, nil
}

// GetTitle returns the currently playing track title.
func (m *MPRIS) GetTitle() (string, error) {
	v, err := m.getProperty("org.mpris.MediaPlayer2.Player.Metadata")
	if err != nil {
		return "", err
	}

	meta, ok := v.Value().(map[string]dbus.Variant)
	if !ok {
		return "", errors.New("unexpected type for Metadata")
	}

	title, ok := meta["xesam:title"]
	if !ok {
		return "", errors.New("no title in metadata")
	}
	str, ok := title.Value().(string)
	if !ok {
		return "", errors.New("unexpected type for title")
	}
	return str, nil
}

// GetPosition returns playback position in microseconds.
// Needed for syncing lyrics to timestamps.
func (m *MPRIS) GetPosition() (int64, error) {
	v, err := m.getProperty("org.mpris.MediaPlayer2.Player.Position")
	if err != nil {
		return 0, err
	}
	pos, ok := v.Value().(int64)
	if !ok {
		return 0, errors.New("unexpected type for Position")
	}
	return pos, nil
}

// GetDuration returns the track duration in seconds.
// Needed for querying data in the lrclib API.
func (m *MPRIS) GetDuration() (int64, error) {
	v, err := m.getProperty("org.mpris.MediaPlayer2.Player.Metadata")
	if err != nil {
		return 0, err
	}

	meta, ok := v.Value().(map[string]dbus.Variant)
	if !ok {
		return 0, errors.New("unexpected type for Metadata")
	}

	title, ok := meta["mpris:length"]
	if !ok {
		return 0, errors.New("no duration in metadata")
	}
	result, ok := title.Value().(int64)
	if !ok {
		return 0, errors.New("unexpected type for duration")
	}

	seconds := int64(math.Floor(float64(result) / float64(DivFactorForLength)))
	return seconds, nil
}
