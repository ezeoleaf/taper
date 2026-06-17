package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseAndToggleChecklist(t *testing.T) {
	content := "# Pack\n\n- [ ] shoes\n- [x] bib\n- [ ] gels\n"
	items := ParseChecklist(content)
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if items[0].Checked || items[1].Text != "bib" || !items[1].Checked {
		t.Fatalf("unexpected parse: %+v", items)
	}

	toggled := toggleCheckboxLine("- [ ] shoes")
	if toggled != "- [ ] Running shoes" && toggled != "- [x] shoes" {
		if strings.Contains(toggled, "]]") {
			t.Fatalf("duplicate bracket: %q", toggled)
		}
	}
	if !strings.Contains(toggled, "[x]") {
		t.Fatalf("expected checked line, got %q", toggled)
	}
	untoggled := toggleCheckboxLine("- [x] bib")
	if strings.Contains(untoggled, "]]") {
		t.Fatalf("duplicate bracket on uncheck: %q", untoggled)
	}
}

func TestUpdateRace(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", root)

	db, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	race, err := CreateRace(&db, RaceInput{Name: "Ultra", Date: "2026-10-01"})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := UpdateRace(&db, race.ID, RaceInput{
		Name:     "Ultra 50k",
		Date:     "2026-10-01",
		Distance: "50k",
		GoalTime: "8:00:00",
		Status:   "registered",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Ultra 50k" || updated.GoalTime != "8:00:00" {
		t.Fatalf("unexpected update: %+v", updated)
	}
}

func TestCreateRaceFiles(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", root)

	db, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	race, err := CreateRace(&db, RaceInput{
		Name:     "Test Marathon",
		Date:     "2026-09-01",
		Distance: "42.2k",
		Location: "Berlin",
		Status:   "registered",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"log.md", "strategy.md", "packing.md"} {
		path := filepath.Join(root, "taper", "races", race.ID, name)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
	jsonPath := filepath.Join(root, "taper", "races.json")
	if _, err := os.Stat(jsonPath); err != nil {
		t.Fatalf("missing races.json: %v", err)
	}
}
