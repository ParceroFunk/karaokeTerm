// Package mpris manages the Linux player call for current playing media
// Properties are on the following link: https://www.freedesktop.org/wiki/Specifications/mpris-spec/metadata/?__goaway_challenge=meta-refresh&__goaway_id=89c89904e1313541ad6814801c76ffac&__goaway_referer=https%3A%2F%2Fwww.google.com%2F
package mpris

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/godbus/dbus/v5"
)

// DivFactorForLength converts mpris:length (microseconds) to seconds.
const DivFactorForLength = 1_000_000

type MPRIS struct {
	conn       *dbus.Conn
	playerName string // cached, avoids re-listing D-Bus names every call
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
	if m.playerName != "" {
		return m.playerName, nil
	}
	name, err := m.findPlayer()
	if err != nil {
		return "", err
	}
	m.playerName = name
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
		m.playerName = ""
		return dbus.Variant{}, fmt.Errorf("failed to get property %s: %w", name, err)
	}
	return val, nil
}

// TrackMetadata bundles everything read from a single Metadata fetch,
// avoiding redundant D-Bus round-trips for the same property.
type TrackMetadata struct {
	Title    string
	Artist   string
	Duration int64  // seconds
	TrackID  string // mpris:trackid — unique per track, cheap to compare for change detection
}

// GetMetadata fetches org.mpris.MediaPlayer2.Player.Metadata once and
// parses all fields from it, instead of one D-Bus call per field.
func (m *MPRIS) GetMetadata() (TrackMetadata, error) {
	v, err := m.getProperty("org.mpris.MediaPlayer2.Player.Metadata")
	if err != nil {
		return TrackMetadata{}, err
	}

	meta, ok := v.Value().(map[string]dbus.Variant)
	if !ok {
		return TrackMetadata{}, errors.New("unexpected type for Metadata")
	}

	md := TrackMetadata{}

	if title, ok := meta["xesam:title"]; ok {
		if str, ok := title.Value().(string); ok {
			md.Title = str
		}
	}

	if artist, ok := meta["xesam:artist"]; ok {
		if arr, ok := artist.Value().([]string); ok && len(arr) > 0 {
			md.Artist = arr[0]
		}
	}

	if length, ok := meta["mpris:length"]; ok {
		if us, ok := length.Value().(int64); ok {
			md.Duration = us / DivFactorForLength
		}
	}

	if trackID, ok := meta["mpris:trackid"]; ok {
		switch t := trackID.Value().(type) {
		case dbus.ObjectPath:
			md.TrackID = string(t)
		case string:
			md.TrackID = t
		}
	}

	if md.Title == "" {
		return md, errors.New("no title in metadata")
	}

	return md, nil
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
