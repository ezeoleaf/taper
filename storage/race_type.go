package storage

import "strings"

// Race type tags.
const (
	TypeRoad  = "road"
	TypeTrail = "trail"
	TypeUltra = "ultra"
	TypeTri   = "tri"
)

// ValidRaceTypes lists supported race type values.
var ValidRaceTypes = []string{TypeRoad, TypeTrail, TypeUltra, TypeTri}

// NormalizeRaceType lowercases and maps aliases to a canonical type (empty if unknown/blank).
func NormalizeRaceType(raw string) string {
	s := strings.TrimSpace(strings.ToLower(raw))
	switch s {
	case "", "none":
		return ""
	case "road", "road race":
		return TypeRoad
	case "trail", "trail race":
		return TypeTrail
	case "ultra", "ultramarathon", "ultra marathon":
		return TypeUltra
	case "tri", "triathlon", "triathlon race":
		return TypeTri
	default:
		for _, t := range ValidRaceTypes {
			if s == t {
				return t
			}
		}
		return s
	}
}

// TypeLabel returns a display label for a race type, or empty.
func TypeLabel(t string) string {
	t = NormalizeRaceType(t)
	if t == "" {
		return ""
	}
	switch t {
	case TypeRoad:
		return "road"
	case TypeTrail:
		return "trail"
	case TypeUltra:
		return "ultra"
	case TypeTri:
		return "tri"
	default:
		return t
	}
}
