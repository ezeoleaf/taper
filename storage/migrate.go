package storage

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// MigrateResult summarizes a migration/backfill pass.
type MigrateResult struct {
	RacesProcessed int
	FilesCreated   int
}

// MigrateAll ensures every race directory exists and seeds any missing markdown files.
func MigrateAll(db *Database) (MigrateResult, error) {
	var res MigrateResult
	for _, r := range db.Races {
		res.RacesProcessed++
		n, err := backfillRaceFiles(r.ID)
		if err != nil {
			return res, fmt.Errorf("migrate race %q: %w", r.ID, err)
		}
		res.FilesCreated += n
	}
	return res, nil
}

// backfillRaceFiles creates the race directory and any missing markdown seeds.
// Returns the number of files created.
func backfillRaceFiles(id string) (int, error) {
	dir, err := RaceDir(id)
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, err
	}
	created := 0
	for name, content := range AllDocSeeds() {
		path, err := DocPath(id, name)
		if err != nil {
			return created, err
		}
		if _, err := os.Stat(path); err == nil {
			continue
		} else if !errors.Is(err, fs.ErrNotExist) {
			return created, err
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return created, err
		}
		created++
	}
	return created, nil
}

// ensureRaceFiles is kept for create-race; delegates to backfill.
func ensureRaceFiles(id string) error {
	_, err := backfillRaceFiles(id)
	return err
}
