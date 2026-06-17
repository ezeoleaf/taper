package storage

import (
	"testing"
	"time"
)

func TestNormalizeRaceType(t *testing.T) {
	if got := NormalizeRaceType("Triathlon"); got != TypeTri {
		t.Fatalf("got %q", got)
	}
	if got := NormalizeRaceType(""); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestIsInTaperWindow(t *testing.T) {
	now := time.Date(2026, 3, 1, 12, 0, 0, 0, time.Local)
	race := Race{Name: "Test", Date: "2026-03-15"}
	taperDay := time.Date(2026, 3, 10, 0, 0, 0, 0, time.Local)
	if !IsInTaperWindow(taperDay, race, now) {
		t.Fatal("expected taper day")
	}
	raceDay := time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local)
	if IsInTaperWindow(raceDay, race, now) {
		t.Fatal("race day is not taper")
	}
}

func TestDescribeDayRace(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	races := []Race{{Name: "A", Date: "2026-06-10", Type: TypeTrail}}
	day := time.Date(2026, 6, 10, 0, 0, 0, 0, time.Local)
	info := DescribeDay(day, races, now)
	if !info.IsRaceDay || len(info.Races) != 1 {
		t.Fatalf("unexpected %+v", info)
	}
}
