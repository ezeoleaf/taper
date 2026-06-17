package ui

import (
	"taper/storage"

	"github.com/charmbracelet/bubbles/textinput"
)

const raceFormFieldCount = 21

type formSection struct {
	name  string
	start int
	count int
}

var raceFormSections = []formSection{
	{name: "Core", start: 0, count: 7},
	{name: "Registration", start: 7, count: 3},
	{name: "Travel", start: 10, count: 3},
	{name: "Training", start: 13, count: 3},
	{name: "Post-race", start: 16, count: 5},
}

type raceForm struct {
	inputs  [raceFormFieldCount]textinput.Model
	section int
	field   int
}

func newRaceForm() raceForm {
	specs := []struct {
		label       string
		placeholder string
		limit       int
		width       int
	}{
		{"Name", "Race name", 80, 40},
		{"Date", "YYYY-MM-DD", 10, 16},
		{"Distance", "e.g. marathon, 50k", 40, 30},
		{"Type", "road · trail · ultra · tri", 16, 20},
		{"Location", "City, country", 60, 40},
		{"Goal time", "e.g. 3:30:00", 16, 16},
		{"Status", "planned · registered · completed", 20, 30},
		{"Entry fee", "e.g. $120", 24, 20},
		{"Confirmation #", "registration id", 40, 30},
		{"Bib pickup", "time & location", 80, 40},
		{"Hotel", "name & dates", 80, 40},
		{"Flights", "arrival / departure", 80, 40},
		{"Transport", "to start line", 80, 40},
		{"Last long run", "YYYY-MM-DD", 10, 16},
		{"Peak week", "notes or date range", 60, 30},
		{"Health / taper", "niggles, flags", 120, 40},
		{"Result time", "finish time", 16, 16},
		{"Photos link", "album URL", 120, 40},
		{"Lessons learned", "key takeaways", 120, 40},
		{"Recovery plan", "next 2 weeks", 120, 40},
		{"Notes", "anything else", 200, 50},
	}
	var f raceForm
	for i, spec := range specs {
		ti := textinput.New()
		ti.Placeholder = spec.placeholder
		ti.CharLimit = spec.limit
		ti.Width = spec.width
		f.inputs[i] = ti
	}
	return f
}

func (f *raceForm) sectionSpec() formSection {
	if f.section < 0 || f.section >= len(raceFormSections) {
		return raceFormSections[0]
	}
	return raceFormSections[f.section]
}

func (f *raceForm) fieldIndex() int {
	spec := f.sectionSpec()
	idx := spec.start + f.field
	if idx < 0 || idx >= raceFormFieldCount {
		return 0
	}
	return idx
}

func (f *raceForm) focusCurrent() {
	spec := f.sectionSpec()
	if f.field >= spec.count {
		f.field = spec.count - 1
	}
	if f.field < 0 {
		f.field = 0
	}
	for i := range f.inputs {
		f.inputs[i].Blur()
	}
	f.inputs[f.fieldIndex()].Focus()
}

func (f *raceForm) nextField() {
	spec := f.sectionSpec()
	f.field++
	if f.field >= spec.count {
		f.field = 0
	}
	f.focusCurrent()
}

func (f *raceForm) prevField() {
	spec := f.sectionSpec()
	f.field--
	if f.field < 0 {
		f.field = spec.count - 1
	}
	f.focusCurrent()
}

func (f *raceForm) nextSection() {
	f.section = (f.section + 1) % len(raceFormSections)
	f.field = 0
	f.focusCurrent()
}

func (f *raceForm) prevSection() {
	f.section--
	if f.section < 0 {
		f.section = len(raceFormSections) - 1
	}
	f.field = 0
	f.focusCurrent()
}

func (f *raceForm) blurAll() {
	for i := range f.inputs {
		f.inputs[i].Blur()
	}
}

func (f *raceForm) clear() {
	for i := range f.inputs {
		f.inputs[i].SetValue("")
	}
	f.section = 0
	f.field = 0
}

func (f *raceForm) loadRace(r storage.Race) {
	vals := []string{
		r.Name, r.Date, r.Distance, r.Type, r.Location, r.GoalTime, r.Status,
		r.EntryFee, r.Confirmation, r.BibPickup,
		r.Hotel, r.Flights, r.Transport,
		r.LastLongRun, r.PeakWeek, r.HealthNotes,
		r.Result, r.PhotosLink, r.LessonsLearned, r.RecoveryPlan, r.Notes,
	}
	for i, v := range vals {
		f.inputs[i].SetValue(v)
	}
	f.section = 0
	f.field = 0
	f.focusCurrent()
}

func (f raceForm) toInput() storage.RaceInput {
	return storage.RaceInput{
		Name: f.inputs[0].Value(), Date: f.inputs[1].Value(),
		Distance: f.inputs[2].Value(), Type: f.inputs[3].Value(),
		Location: f.inputs[4].Value(), GoalTime: f.inputs[5].Value(),
		Status: f.inputs[6].Value(),
		EntryFee: f.inputs[7].Value(), Confirmation: f.inputs[8].Value(),
		BibPickup: f.inputs[9].Value(),
		Hotel: f.inputs[10].Value(), Flights: f.inputs[11].Value(),
		Transport: f.inputs[12].Value(),
		LastLongRun: f.inputs[13].Value(), PeakWeek: f.inputs[14].Value(),
		HealthNotes: f.inputs[15].Value(),
		Result: f.inputs[16].Value(), PhotosLink: f.inputs[17].Value(),
		LessonsLearned: f.inputs[18].Value(), RecoveryPlan: f.inputs[19].Value(),
		Notes: f.inputs[20].Value(),
	}
}

func (f raceForm) labels() []string {
	return []string{
		"Name", "Date", "Distance", "Type", "Location", "Goal time", "Status",
		"Entry fee", "Confirmation #", "Bib pickup",
		"Hotel", "Flights", "Transport",
		"Last long run", "Peak week", "Health / taper",
		"Result time", "Photos link", "Lessons learned", "Recovery plan", "Notes",
	}
}
