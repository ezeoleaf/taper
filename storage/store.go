package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	pathpkg "path/filepath"
	"sort"
	"strings"
	"time"
)

// Config holds global taper settings stored in races.json.
type Config struct {
	DefaultEditor string `json:"default_editor,omitempty"`
}

// Race describes a single race entry.
type Race struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Date     string `json:"date"` // YYYY-MM-DD
	Distance string `json:"distance,omitempty"`
	Type     string `json:"type,omitempty"` // road, trail, ultra, tri
	Location string `json:"location,omitempty"`
	GoalTime string `json:"goal_time,omitempty"`
	Status   string `json:"status,omitempty"` // planned, registered, completed, dns, dnf
	Result   string `json:"result_time,omitempty"`
	Notes    string `json:"notes,omitempty"`

	EntryFee     string `json:"entry_fee,omitempty"`
	Confirmation string `json:"confirmation,omitempty"`
	BibPickup    string `json:"bib_pickup,omitempty"`

	Hotel     string `json:"hotel,omitempty"`
	Flights   string `json:"flights,omitempty"`
	Transport string `json:"transport,omitempty"`

	LastLongRun string `json:"last_long_run,omitempty"`
	PeakWeek    string `json:"peak_week,omitempty"`
	HealthNotes string `json:"health_notes,omitempty"`

	PhotosLink     string `json:"photos_link,omitempty"`
	LessonsLearned string `json:"lessons_learned,omitempty"`
	RecoveryPlan   string `json:"recovery_plan,omitempty"`
}

// Database is the on-disk races.json document.
type Database struct {
	Config Config `json:"config"`
	Races  []Race `json:"races"`
}

// Load reads races.json, creating defaults when missing.
func Load() (Database, error) {
	if err := EnsureLayout(); err != nil {
		return Database{}, err
	}
	path, err := RacesJSONPath()
	if err != nil {
		return Database{}, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			db := Database{Races: []Race{}}
			if saveErr := Save(db); saveErr != nil {
				return Database{}, saveErr
			}
			return db, nil
		}
		return Database{}, err
	}
	var db Database
	if err := json.Unmarshal(b, &db); err != nil {
		return Database{}, fmt.Errorf("parse races.json: %w", err)
	}
	if db.Races == nil {
		db.Races = []Race{}
	}
	if _, err := MigrateAll(&db); err != nil {
		return db, err
	}
	return db, nil
}

// Save writes races.json.
func Save(db Database) error {
	if err := EnsureLayout(); err != nil {
		return err
	}
	path, err := RacesJSONPath()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

// ParseDate parses a race date string in local time.
func ParseDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", strings.TrimSpace(s), time.Local)
}

// RaceDate returns the parsed date or zero time on error.
func (r Race) RaceDate() time.Time {
	t, err := ParseDate(r.Date)
	if err != nil {
		return time.Time{}
	}
	return t
}

// IsUpcoming reports whether the race is today or in the future.
func (r Race) IsUpcoming(now time.Time) bool {
	d := r.RaceDate()
	if d.IsZero() {
		return true
	}
	day := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, now.Location())
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return !day.Before(today)
}

// PartitionRaces splits races into upcoming and past lists, each sorted by date.
func PartitionRaces(races []Race, now time.Time) (upcoming, past []Race) {
	for _, r := range races {
		if r.IsUpcoming(now) {
			upcoming = append(upcoming, r)
		} else {
			past = append(past, r)
		}
	}
	sort.Slice(upcoming, func(i, j int) bool {
		di, dj := upcoming[i].RaceDate(), upcoming[j].RaceDate()
		if di.IsZero() {
			return false
		}
		if dj.IsZero() {
			return true
		}
		return di.Before(dj)
	})
	sort.Slice(past, func(i, j int) bool {
		di, dj := past[i].RaceDate(), past[j].RaceDate()
		if di.IsZero() {
			return false
		}
		if dj.IsZero() {
			return true
		}
		return di.After(dj)
	})
	return upcoming, past
}

// RaceInput is the editable metadata for a race (id set only on update).
type RaceInput struct {
	Name     string
	Date     string
	Distance string
	Type     string
	Location string
	GoalTime string
	Status   string
	Result   string
	Notes    string

	EntryFee     string
	Confirmation string
	BibPickup    string

	Hotel     string
	Flights   string
	Transport string

	LastLongRun string
	PeakWeek    string
	HealthNotes string

	PhotosLink     string
	LessonsLearned string
	RecoveryPlan   string
}

// CompletionInput captures post-race outcome from the completion dialog.
type CompletionInput struct {
	Outcome        string // completed, dnf, dns
	ResultTime     string
	LessonsLearned string
	PhotosLink     string
	RecoveryPlan   string
	Notes          string
}

func trimFields(in *RaceInput) {
	in.Name = strings.TrimSpace(in.Name)
	in.Date = strings.TrimSpace(in.Date)
	in.Distance = strings.TrimSpace(in.Distance)
	in.Type = NormalizeRaceType(in.Type)
	in.Location = strings.TrimSpace(in.Location)
	in.GoalTime = strings.TrimSpace(in.GoalTime)
	in.Status = strings.TrimSpace(strings.ToLower(in.Status))
	in.Result = strings.TrimSpace(in.Result)
	in.Notes = strings.TrimSpace(in.Notes)
	in.EntryFee = strings.TrimSpace(in.EntryFee)
	in.Confirmation = strings.TrimSpace(in.Confirmation)
	in.BibPickup = strings.TrimSpace(in.BibPickup)
	in.Hotel = strings.TrimSpace(in.Hotel)
	in.Flights = strings.TrimSpace(in.Flights)
	in.Transport = strings.TrimSpace(in.Transport)
	in.LastLongRun = strings.TrimSpace(in.LastLongRun)
	in.PeakWeek = strings.TrimSpace(in.PeakWeek)
	in.HealthNotes = strings.TrimSpace(in.HealthNotes)
	in.PhotosLink = strings.TrimSpace(in.PhotosLink)
	in.LessonsLearned = strings.TrimSpace(in.LessonsLearned)
	in.RecoveryPlan = strings.TrimSpace(in.RecoveryPlan)
}

func raceFromInput(id string, in RaceInput) Race {
	return Race{
		ID:             id,
		Name:           in.Name,
		Date:           in.Date,
		Distance:       in.Distance,
		Type:           in.Type,
		Location:       in.Location,
		GoalTime:       in.GoalTime,
		Status:         in.Status,
		Result:         in.Result,
		Notes:          in.Notes,
		EntryFee:       in.EntryFee,
		Confirmation:   in.Confirmation,
		BibPickup:      in.BibPickup,
		Hotel:          in.Hotel,
		Flights:        in.Flights,
		Transport:      in.Transport,
		LastLongRun:    in.LastLongRun,
		PeakWeek:       in.PeakWeek,
		HealthNotes:    in.HealthNotes,
		PhotosLink:     in.PhotosLink,
		LessonsLearned: in.LessonsLearned,
		RecoveryPlan:   in.RecoveryPlan,
	}
}

func normalizeRaceInput(in RaceInput) (RaceInput, error) {
	trimFields(&in)
	if in.Name == "" {
		return in, errors.New("race name is required")
	}
	if in.Date == "" {
		in.Date = time.Now().Format("2006-01-02")
	}
	if _, err := ParseDate(in.Date); err != nil {
		return in, fmt.Errorf("invalid date %q (use YYYY-MM-DD)", in.Date)
	}
	if in.Status == "" {
		in.Status = "planned"
	}
	return in, nil
}

// CreateRace adds a race, creates its directory, and seeds markdown files.
func CreateRace(db *Database, in RaceInput) (Race, error) {
	norm, err := normalizeRaceInput(in)
	if err != nil {
		return Race{}, err
	}

	id := uniqueID(db, slugify(norm.Name, norm.Date))
	race := raceFromInput(id, norm)
	if err := ensureRaceFiles(race.ID); err != nil {
		return Race{}, err
	}
	db.Races = append(db.Races, race)
	if err := Save(*db); err != nil {
		return Race{}, err
	}
	return race, nil
}

// UpdateRace updates metadata for an existing race (id is immutable).
func UpdateRace(db *Database, id string, in RaceInput) (Race, error) {
	norm, err := normalizeRaceInput(in)
	if err != nil {
		return Race{}, err
	}
	id = sanitizeID(id)
	for i, r := range db.Races {
		if r.ID != id {
			continue
		}
		updated := raceFromInput(r.ID, norm)
		db.Races[i] = updated
		if err := Save(*db); err != nil {
			return Race{}, err
		}
		return updated, nil
	}
	return Race{}, fmt.Errorf("race %q not found", id)
}

// CompleteRace records a race outcome and post-race metadata.
func CompleteRace(db *Database, id string, in CompletionInput) (Race, error) {
	in.Outcome = strings.TrimSpace(strings.ToLower(in.Outcome))
	if in.Outcome == "" {
		in.Outcome = "completed"
	}
	switch in.Outcome {
	case "completed", "dnf", "dns":
	default:
		return Race{}, fmt.Errorf("outcome must be completed, dnf, or dns")
	}
	in.ResultTime = strings.TrimSpace(in.ResultTime)
	in.LessonsLearned = strings.TrimSpace(in.LessonsLearned)
	in.PhotosLink = strings.TrimSpace(in.PhotosLink)
	in.RecoveryPlan = strings.TrimSpace(in.RecoveryPlan)
	in.Notes = strings.TrimSpace(in.Notes)

	id = sanitizeID(id)
	for i, r := range db.Races {
		if r.ID != id {
			continue
		}
		r.Status = in.Outcome
		if in.Outcome == "completed" && in.ResultTime != "" {
			r.Result = in.ResultTime
		}
		if in.LessonsLearned != "" {
			r.LessonsLearned = in.LessonsLearned
		}
		if in.PhotosLink != "" {
			r.PhotosLink = in.PhotosLink
		}
		if in.RecoveryPlan != "" {
			r.RecoveryPlan = in.RecoveryPlan
		}
		if in.Notes != "" {
			if r.Notes != "" {
				r.Notes = r.Notes + "\n" + in.Notes
			} else {
				r.Notes = in.Notes
			}
		}
		db.Races[i] = r
		if err := Save(*db); err != nil {
			return Race{}, err
		}
		return r, nil
	}
	return Race{}, fmt.Errorf("race %q not found", id)
}

// DeleteRace removes a race from the database and its directory.
func DeleteRace(db *Database, id string) error {
	id = sanitizeID(id)
	var kept []Race
	found := false
	for _, r := range db.Races {
		if r.ID == id {
			found = true
			continue
		}
		kept = append(kept, r)
	}
	if !found {
		return fmt.Errorf("race %q not found", id)
	}
	dir, err := RaceDir(id)
	if err != nil {
		return err
	}
	_ = os.RemoveAll(dir)
	db.Races = kept
	return Save(*db)
}

func slugify(name, date string) string {
	base := sanitizeID(strings.ToLower(name))
	if base == "race" || base == "" {
		base = "race"
	}
	return sanitizeID(fmt.Sprintf("%s-%s", base, strings.ReplaceAll(date, "/", "-")))
}

func uniqueID(db *Database, base string) string {
	used := make(map[string]struct{}, len(db.Races))
	for _, r := range db.Races {
		used[r.ID] = struct{}{}
	}
	id := base
	for i := 2; ; i++ {
		if _, ok := used[id]; !ok {
			return id
		}
		id = fmt.Sprintf("%s-%d", base, i)
	}
}

func pathForRaceFile(id, name string) (string, error) {
	return DocPath(id, name)
}

const defaultLog = `# Race Log

## Pre-race

-

## Race day

-

## Post-race

-
`

const defaultStrategy = `# Race Strategy

## Pacing

-

## Nutrition

-

## Mental cues

-
`

const defaultPacking = `# Packing List

## Gear

- [ ] Running shoes
- [ ] Race bib & pins
- [ ] Watch / HR strap
- [ ] Gels / fuel
- [ ] Hydration vest or belt

## Clothing

- [ ] Race kit
- [ ] Warm-up layers
- [ ] Hat / sunglasses

## Travel

- [ ] ID & registration
- [ ] Hotel confirmation
- [ ] Transport tickets
`

// ReadFile reads a race markdown file by kind.
func ReadFile(raceID string, kind DocKind) (string, error) {
	path, err := docPathForKind(raceID, kind)
	if err != nil {
		return "", err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	return string(b), nil
}

// WriteFile writes a race markdown file.
func WriteFile(raceID string, kind DocKind, content string) error {
	path, err := docPathForKind(raceID, kind)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(pathpkg.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// EditorPath returns the filesystem path for external editor integration.
func EditorPath(raceID string, kind DocKind) (string, error) {
	return docPathForKind(raceID, kind)
}

func docPathForKind(raceID string, kind DocKind) (string, error) {
	switch kind {
	case DocLog:
		return LogPath(raceID)
	case DocStrategy:
		return StrategyPath(raceID)
	case DocPacking:
		return PackingPath(raceID)
	case DocNutrition:
		return DocPath(raceID, "nutrition.md")
	case DocGear:
		return DocPath(raceID, "gear.md")
	case DocWeather:
		return DocPath(raceID, "weather.md")
	case DocCrew:
		return DocPath(raceID, "crew.md")
	case DocSplits:
		return DocPath(raceID, "splits.md")
	default:
		return "", fmt.Errorf("unknown file kind %q", kind)
	}
}
