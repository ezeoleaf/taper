package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"taper/storage"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const calendarCellW = 4

func renderCalendar(month time.Time, races []storage.Race, width int) string {
	now := time.Now()
	month = time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())

	var b strings.Builder
	b.WriteString(titleStyle.Render("Training Calendar"))
	b.WriteString("\n\n")
	b.WriteString(mutedStyle.Render(", / . change month · j/k scroll · Enter jump to race"))
	b.WriteString("\n\n")

	b.WriteString(renderCountdowns(races, now))
	b.WriteString("\n\n")
	b.WriteString(renderMonthGrid(month, races, now))
	if notes := renderMonthRaces(month, races); notes != "" {
		b.WriteString("\n\n")
		b.WriteString(notes)
	}
	b.WriteString("\n\n")
	b.WriteString(renderCalendarLegend())
	return b.String()
}

func renderCountdowns(races []storage.Race, now time.Time) string {
	upcoming := storage.UpcomingCountdown(races, now, 5)
	if len(upcoming) == 0 {
		return subtleBoxStyle.Render(mutedStyle.Render("No upcoming races — press n on Races tab to add one."))
	}

	var lines []string
	lines = append(lines, titleStyle.Render("Countdown"))
	for _, r := range upcoming {
		days := storage.DaysUntil(r, now)
		if days < 0 {
			continue
		}
		name := ansi.Truncate(r.Name, 26, "…")
		var when string
		switch {
		case days == 0:
			when = "today"
		case days <= storage.TaperDays:
			when = fmt.Sprintf("%dd taper", days)
		default:
			when = fmt.Sprintf("%dd", days)
		}
		line := padCell(name, 28) + padCell(when, 12)
		if typ := storage.TypeLabel(r.Type); typ != "" {
			line += mutedStyle.Render(typ)
		}
		lines = append(lines, line)
	}
	return subtleBoxStyle.Render(strings.Join(lines, "\n"))
}

func renderMonthGrid(month time.Time, races []storage.Race, now time.Time) string {
	year, mon := month.Year(), month.Month()
	first := time.Date(year, mon, 1, 0, 0, 0, 0, month.Location())
	daysInMonth := time.Date(year, mon+1, 0, 0, 0, 0, 0, month.Location()).Day()
	startWeekday := int(first.Weekday())

	var b strings.Builder
	b.WriteString(titleStyle.Render(first.Format("January 2006")))
	b.WriteString("\n\n")

	headers := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	for _, h := range headers {
		b.WriteString(padCell(mutedStyle.Render(h), calendarCellW))
	}
	b.WriteString("\n")

	var row strings.Builder
	for i := 0; i < startWeekday; i++ {
		row.WriteString(padCell("", calendarCellW))
	}

	for day := 1; day <= daysInMonth; day++ {
		d := time.Date(year, mon, day, 0, 0, 0, 0, month.Location())
		info := storage.DescribeDay(d, races, now)
		row.WriteString(formatCalendarCell(day, info))

		if (startWeekday+day)%7 == 0 {
			b.WriteString(row.String())
			b.WriteString("\n")
			row.Reset()
		}
	}
	if row.Len() > 0 {
		b.WriteString(row.String())
		b.WriteString("\n")
	}
	return b.String()
}

func formatCalendarCell(day int, info storage.DayInfo) string {
	label := fmt.Sprintf("%2d", day)
	switch {
	case info.IsRaceDay:
		label = fmt.Sprintf("*%d", day)
		return padCell(raceDayStyle.Render(label), calendarCellW)
	case info.IsToday:
		return padCell(todayStyle.Render(label), calendarCellW)
	case info.IsTaper:
		return padCell(taperDayStyle.Render(label), calendarCellW)
	default:
		return padCell(label, calendarCellW)
	}
}

func renderMonthRaces(month time.Time, races []storage.Race) string {
	var inMonth []storage.Race
	for _, r := range races {
		d := r.RaceDate()
		if d.IsZero() {
			continue
		}
		if d.Year() == month.Year() && d.Month() == month.Month() {
			inMonth = append(inMonth, r)
		}
	}
	if len(inMonth) == 0 {
		return ""
	}
	sort.Slice(inMonth, func(i, j int) bool {
		return inMonth[i].RaceDate().Before(inMonth[j].RaceDate())
	})

	var lines []string
	lines = append(lines, titleStyle.Render("This month"))
	for _, r := range inMonth {
		day := r.RaceDate().Day()
		meta := storage.TypeLabel(r.Type)
		if meta != "" {
			meta = " · " + meta
		}
		lines = append(lines, fmt.Sprintf("  %2d  %s%s", day, r.Name, mutedStyle.Render(meta)))
	}
	return strings.Join(lines, "\n")
}

func renderCalendarLegend() string {
	return mutedStyle.Render("Legend: ") +
		todayStyle.Render(padCell("15", calendarCellW)) +
		taperDayStyle.Render(padCell("14", calendarCellW)) +
		raceDayStyle.Render(padCell("*7", calendarCellW)) +
		mutedStyle.Render(" = today / taper / race")
}

func padCell(s string, width int) string {
	if s == "" {
		return strings.Repeat(" ", width)
	}
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}
