package ui

import (
	"strings"

	"taper/storage"

	"github.com/charmbracelet/bubbles/textinput"
)

const completeFormFieldCount = 6

var completeOutcomes = []string{"completed", "dnf", "dns"}

type completeForm struct {
	inputs  [completeFormFieldCount]textinput.Model
	cursor  int
	outcome int
}

func newCompleteForm() completeForm {
	specs := []struct {
		placeholder string
		limit       int
		width       int
	}{
		{"", 0, 0}, // outcome uses cycling, not text
		{"finish time e.g. 3:28:14", 16, 20},
		{"what went well / lessons", 200, 50},
		{"photos album URL", 120, 40},
		{"recovery next 2 weeks", 200, 50},
		{"additional notes", 200, 50},
	}
	var f completeForm
	for i, spec := range specs {
		if i == 0 {
			continue
		}
		ti := textinput.New()
		ti.Placeholder = spec.placeholder
		ti.CharLimit = spec.limit
		ti.Width = spec.width
		f.inputs[i] = ti
	}
	return f
}

func (f *completeForm) focusField(n int) {
	f.cursor = n
	for i := 1; i < completeFormFieldCount; i++ {
		if i == n {
			f.inputs[i].Focus()
		} else {
			f.inputs[i].Blur()
		}
	}
}

func (f *completeForm) blurAll() {
	for i := 1; i < completeFormFieldCount; i++ {
		f.inputs[i].Blur()
	}
}

func (f *completeForm) nextField() {
	f.focusField((f.cursor + 1) % completeFormFieldCount)
}

func (f *completeForm) prevField() {
	f.focusField((f.cursor - 1 + completeFormFieldCount) % completeFormFieldCount)
}

func (f *completeForm) cycleOutcome() {
	f.outcome = (f.outcome + 1) % len(completeOutcomes)
}

func (f *completeForm) loadRace(r storage.Race) {
	f.outcome = 0
	if r.Status == "dnf" {
		f.outcome = 1
	} else if r.Status == "dns" {
		f.outcome = 2
	}
	f.inputs[1].SetValue(r.Result)
	f.inputs[2].SetValue(r.LessonsLearned)
	f.inputs[3].SetValue(r.PhotosLink)
	f.inputs[4].SetValue(r.RecoveryPlan)
	f.inputs[5].SetValue("")
	f.focusField(0)
}

func (f completeForm) outcomeLabel() string {
	return completeOutcomes[f.outcome]
}

func (f completeForm) toInput() storage.CompletionInput {
	return storage.CompletionInput{
		Outcome:        f.outcomeLabel(),
		ResultTime:     strings.TrimSpace(f.inputs[1].Value()),
		LessonsLearned: strings.TrimSpace(f.inputs[2].Value()),
		PhotosLink:     strings.TrimSpace(f.inputs[3].Value()),
		RecoveryPlan:   strings.TrimSpace(f.inputs[4].Value()),
		Notes:          strings.TrimSpace(f.inputs[5].Value()),
	}
}

func (f completeForm) fieldLabels() []string {
	return []string{
		"Outcome",
		"Finish time",
		"Lessons learned",
		"Photos link",
		"Recovery plan",
		"Notes",
	}
}
