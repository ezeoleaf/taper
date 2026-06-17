// Package storage manages taper config and per-race markdown files under the user config directory.
//
// Layout:
//
//	~/.config/taper/
//	  races.json
//	  races/
//	    <id>/
//	      log.md, strategy.md, packing.md, nutrition.md, gear.md,
//	      weather.md, crew.md, splits.md
package storage

import (
	"os"
	pathpkg "path/filepath"
	"strings"
)

const (
	racesFile = "races.json"
	racesDir  = "races"
)

// ConfigDir returns the app config root, e.g. ~/.config/taper.
func ConfigDir() (string, error) {
	if xdg := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); xdg != "" {
		return pathpkg.Join(xdg, "taper"), nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return pathpkg.Join(base, "taper"), nil
}

// RacesJSONPath returns the path to races.json.
func RacesJSONPath() (string, error) {
	root, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return pathpkg.Join(root, racesFile), nil
}

// RacesRoot returns the directory containing per-race subdirectories.
func RacesRoot() (string, error) {
	root, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return pathpkg.Join(root, racesDir), nil
}

// RaceDir returns the directory for a specific race id.
func RaceDir(id string) (string, error) {
	root, err := RacesRoot()
	if err != nil {
		return "", err
	}
	return pathpkg.Join(root, sanitizeID(id)), nil
}

// LogPath, StrategyPath, and PackingPath return markdown file paths for a race.
func LogPath(id string) (string, error) {
	return raceFilePath(id, "log.md")
}

func StrategyPath(id string) (string, error) {
	return raceFilePath(id, "strategy.md")
}

func PackingPath(id string) (string, error) {
	return raceFilePath(id, "packing.md")
}

func DocPath(id, filename string) (string, error) {
	return raceFilePath(id, filename)
}

func raceFilePath(id, filename string) (string, error) {
	dir, err := RaceDir(id)
	if err != nil {
		return "", err
	}
	return pathpkg.Join(dir, filename), nil
}

// EnsureLayout creates ~/.config/taper and races/ if missing.
func EnsureLayout() error {
	root, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	races, err := RacesRoot()
	if err != nil {
		return err
	}
	return os.MkdirAll(races, 0o755)
}

func sanitizeID(id string) string {
	id = strings.TrimSpace(id)
	var b strings.Builder
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('-')
		}
	}
	s := b.String()
	if s == "" {
		return "race"
	}
	return s
}
