package storage

import "time"

// TaperDays is the default taper window before race day (inclusive).
const TaperDays = 14

// DayInfo describes how a calendar day relates to races.
type DayInfo struct {
	IsToday   bool
	IsRaceDay bool
	IsTaper   bool
	Races     []Race
}

// DaysUntil returns whole days from today until race day (0 = today, negative = past).
func DaysUntil(r Race, now time.Time) int {
	d := r.RaceDate()
	if d.IsZero() {
		return -1
	}
	today := dateOnly(now)
	raceDay := dateOnly(d)
	return int(raceDay.Sub(today).Hours() / 24)
}

// IsInTaperWindow reports whether day is within the taper window before an upcoming race.
func IsInTaperWindow(day time.Time, r Race, now time.Time) bool {
	if !r.IsUpcoming(now) {
		return false
	}
	raceDay := dateOnly(r.RaceDate())
	if raceDay.IsZero() {
		return false
	}
	d := dateOnly(day)
	if !d.Before(raceDay) {
		return false
	}
	days := int(raceDay.Sub(d).Hours() / 24)
	return days >= 1 && days <= TaperDays
}

// DescribeDay returns calendar metadata for a single day.
func DescribeDay(day time.Time, races []Race, now time.Time) DayInfo {
	d := dateOnly(day)
	info := DayInfo{IsToday: d.Equal(dateOnly(now))}
	for _, r := range races {
		rd := r.RaceDate()
		if rd.IsZero() {
			continue
		}
		if dateOnly(rd).Equal(d) {
			info.IsRaceDay = true
			info.Races = append(info.Races, r)
		} else if IsInTaperWindow(d, r, now) {
			info.IsTaper = true
		}
	}
	return info
}

// UpcomingCountdown returns upcoming races sorted by date for countdown widgets.
func UpcomingCountdown(races []Race, now time.Time, limit int) []Race {
	upcoming, _ := PartitionRaces(races, now)
	if limit > 0 && len(upcoming) > limit {
		upcoming = upcoming[:limit]
	}
	return upcoming
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
