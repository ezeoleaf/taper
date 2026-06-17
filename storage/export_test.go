package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrateBackfill(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", root)

	db, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	race, err := CreateRace(&db, RaceInput{Name: "Trail 50k", Date: "2026-08-01"})
	if err != nil {
		t.Fatal(err)
	}
	// remove one seeded file
	nutrition := filepath.Join(root, "taper", "races", race.ID, "nutrition.md")
	if err := os.Remove(nutrition); err != nil {
		t.Fatal(err)
	}
	res, err := MigrateAll(&db)
	if err != nil {
		t.Fatal(err)
	}
	if res.FilesCreated < 1 {
		t.Fatalf("expected at least 1 file created, got %+v", res)
	}
	if _, err := os.Stat(nutrition); err != nil {
		t.Fatalf("nutrition.md not backfilled: %v", err)
	}
}

func TestRaceMatchesFilter(t *testing.T) {
	r := Race{Name: "Boston Marathon", Location: "Boston", Status: "registered"}
	if !RaceMatchesFilter(r, "boston") {
		t.Fatal("expected match")
	}
	if RaceMatchesFilter(r, "ultra") {
		t.Fatal("expected no match")
	}
}

func TestExportMarkdownPacket(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", root)

	db, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	race, err := CreateRace(&db, RaceInput{Name: "City 10k", Date: "2026-05-01", Distance: "10k"})
	if err != nil {
		t.Fatal(err)
	}
	path, err := ExportMarkdownPacket(db, race.ID)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	body := string(b)
	if !strings.Contains(body, "City 10k") || !strings.Contains(body, "Journal") {
		t.Fatalf("unexpected packet: %s", body[:min(200, len(body))])
	}
}
