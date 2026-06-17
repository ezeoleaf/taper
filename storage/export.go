package storage

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ExportDoc describes a section in the race packet.
type ExportDoc struct {
	Title string
	Kind  DocKind
}

// AllExportDocs is the ordered list of docs included in a race packet.
var AllExportDocs = []ExportDoc{
	{Title: "Journal", Kind: DocLog},
	{Title: "Strategy & pacing", Kind: DocStrategy},
	{Title: "Packing list", Kind: DocPacking},
	{Title: "Nutrition plan", Kind: DocNutrition},
	{Title: "Gear choices", Kind: DocGear},
	{Title: "Weather contingency", Kind: DocWeather},
	{Title: "Crew & support", Kind: DocCrew},
	{Title: "Splits / checkpoints", Kind: DocSplits},
}

// ExportMarkdownPacket writes a single markdown file combining metadata and all race docs.
func ExportMarkdownPacket(db Database, raceID string) (string, error) {
	race, ok := findRace(db, raceID)
	if !ok {
		return "", fmt.Errorf("race %q not found", raceID)
	}
	body, err := buildPacketMarkdown(race)
	if err != nil {
		return "", err
	}
	dir, err := RaceDir(raceID)
	if err != nil {
		return "", err
	}
	out := filepath.Join(dir, "packet.md")
	if err := os.WriteFile(out, []byte(body), 0o644); err != nil {
		return "", err
	}
	return out, nil
}

// ExportPDFPacket writes packet.md then converts to packet.pdf via pandoc when available.
func ExportPDFPacket(db Database, raceID string) (string, error) {
	mdPath, err := ExportMarkdownPacket(db, raceID)
	if err != nil {
		return "", err
	}
	if _, err := exec.LookPath("pandoc"); err != nil {
		return "", fmt.Errorf("pandoc not found (install pandoc for PDF export, markdown saved at %s)", mdPath)
	}
	pdfPath := strings.TrimSuffix(mdPath, ".md") + ".pdf"
	cmd := exec.Command("pandoc", mdPath, "-o", pdfPath, "--from=markdown", "--pdf-engine=pdflatex")
	if _, err := cmd.CombinedOutput(); err != nil {
		cmd = exec.Command("pandoc", mdPath, "-o", pdfPath)
		if out, err2 := cmd.CombinedOutput(); err2 != nil {
			return mdPath, fmt.Errorf("pandoc failed: %v (%s); markdown at %s", err2, strings.TrimSpace(string(out)), mdPath)
		}
	}
	return pdfPath, nil
}

func findRace(db Database, id string) (Race, bool) {
	id = sanitizeID(id)
	for _, r := range db.Races {
		if r.ID == id {
			return r, true
		}
	}
	return Race{}, false
}

func buildPacketMarkdown(r Race) (string, error) {
	var b strings.Builder
	b.WriteString("# ")
	b.WriteString(r.Name)
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("> Exported %s · taper\n\n", time.Now().Format("2006-01-02 15:04")))
	b.WriteString("## Race details\n\n")
	b.WriteString(renderMetaTable(r))
	b.WriteString("\n")

	for _, doc := range AllExportDocs {
		content, err := ReadFile(r.ID, doc.Kind)
		if err != nil {
			return "", err
		}
		content = strings.TrimSpace(content)
		if content == "" {
			continue
		}
		b.WriteString("---\n\n")
		b.WriteString("## ")
		b.WriteString(doc.Title)
		b.WriteString("\n\n")
		b.WriteString(content)
		b.WriteString("\n\n")
	}
	return b.String(), nil
}

func renderMetaTable(r Race) string {
	rows := [][2]string{
		{"Date", r.Date},
		{"Distance", r.Distance},
		{"Type", TypeLabel(r.Type)},
		{"Location", r.Location},
		{"Goal time", r.GoalTime},
		{"Status", r.Status},
		{"Result", r.Result},
		{"Entry fee", r.EntryFee},
		{"Confirmation", r.Confirmation},
		{"Bib pickup", r.BibPickup},
		{"Hotel", r.Hotel},
		{"Flights", r.Flights},
		{"Transport", r.Transport},
		{"Last long run", r.LastLongRun},
		{"Peak week", r.PeakWeek},
		{"Health / taper", r.HealthNotes},
		{"Photos", r.PhotosLink},
		{"Lessons learned", r.LessonsLearned},
		{"Recovery plan", r.RecoveryPlan},
		{"Notes", r.Notes},
	}
	var b strings.Builder
	b.WriteString("| Field | Value |\n| ----- | ----- |\n")
	for _, row := range rows {
		if strings.TrimSpace(row[1]) == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("| %s | %s |\n", row[0], strings.ReplaceAll(row[1], "|", "\\|")))
	}
	return b.String()
}

// RaceMatchesFilter reports whether a race matches a case-insensitive search query.
func RaceMatchesFilter(r Race, query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return true
	}
	hay := strings.ToLower(strings.Join([]string{
		r.Name, r.Date, r.Distance, r.Type, r.Location, r.Status,
		r.GoalTime, r.Confirmation, r.Hotel, r.Notes,
	}, " "))
	return strings.Contains(hay, q)
}
