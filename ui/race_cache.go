package ui

import (
	"strings"

	"taper/storage"
)

type raceBundle struct {
	raceID    string
	width     int
	raw       map[storage.DocKind]string
	glamour   map[storage.DocKind]string
	checklist []storage.CheckItem
}

func loadRaceBundle(raceID string, width int) (*raceBundle, error) {
	if width < 20 {
		width = 20
	}
	b := &raceBundle{
		raceID:   raceID,
		width:    width,
		raw:     make(map[storage.DocKind]string, len(storage.AllDocKinds)),
		glamour: make(map[storage.DocKind]string, len(storage.AllDocKinds)),
	}
	for _, kind := range storage.AllDocKinds {
		content, err := storage.ReadFile(raceID, kind)
		if err != nil {
			return nil, err
		}
		b.raw[kind] = content
		b.glamour[kind] = renderMarkdown(content, width)
	}
	b.checklist = storage.ParseChecklist(b.raw[storage.DocPacking])
	return b, nil
}

func (b *raceBundle) rerender(width int) {
	if width < 20 {
		width = 20
	}
	if b.width == width {
		return
	}
	b.width = width
	for _, kind := range storage.AllDocKinds {
		b.glamour[kind] = renderMarkdown(b.raw[kind], width)
	}
}

func (b *raceBundle) markdown(kind storage.DocKind) string {
	if out, ok := b.glamour[kind]; ok {
		return out
	}
	return mutedStyle.Render("(empty — press e to edit)")
}

func (b *raceBundle) reloadFromDisk(width int) error {
	nb, err := loadRaceBundle(b.raceID, width)
	if err != nil {
		return err
	}
	*b = *nb
	return nil
}

func (b *raceBundle) applyPacking(content string, width int) {
	b.raw[storage.DocPacking] = content
	b.glamour[storage.DocPacking] = renderMarkdown(content, width)
	b.checklist = storage.ParseChecklist(content)
}

func (m *Model) bundleWidth() int {
	return max(20, m.mainViewport.Width-2)
}

func (m *Model) hasLoadedRace() bool {
	return m.bundle != nil && m.selectedRaceID != "" && m.bundle.raceID == m.selectedRaceID
}

func (m *Model) commitRaceSelection() error {
	if len(m.raceList) == 0 {
		m.selectedRaceID = ""
		m.bundle = nil
		m.checklistItems = nil
		return nil
	}
	m.clampRaceCursor()
	id := m.raceList[m.raceListCursor].race.ID
	w := m.bundleWidth()
	bundle, err := loadRaceBundle(id, w)
	if err != nil {
		return err
	}
	m.selectedRaceID = id
	m.bundle = bundle
	m.checklistItems = bundle.checklist
	if m.checklistCursor >= len(m.checklistItems) {
		m.checklistCursor = 0
	}
	m.planDocCursor = 0
	invalidateMarkdownRenderer()
	return nil
}

func (m *Model) reloadBundle() {
	if !m.hasLoadedRace() {
		return
	}
	w := m.bundleWidth()
	if err := m.bundle.reloadFromDisk(w); err != nil {
		m.status = err.Error()
		return
	}
	m.checklistItems = m.bundle.checklist
}

func (m *Model) ensureBundleRendered() {
	if m.bundle != nil {
		m.bundle.rerender(m.bundleWidth())
	}
}

func (m *Model) cachedMarkdown(kind storage.DocKind) string {
	if !m.hasLoadedRace() {
		return mutedStyle.Render("(load a race with Enter on the Races tab)")
	}
	m.ensureBundleRendered()
	return m.bundle.markdown(kind)
}

func (m *Model) clearRaceLoad() {
	m.selectedRaceID = ""
	m.bundle = nil
	m.checklistItems = nil
	m.checklistCursor = 0
	m.planDocCursor = 0
}

func requiresLoadedRace(s section) bool {
	switch s {
	case sectionJournal, sectionChecklist, sectionPlan:
		return true
	default:
		return false
	}
}

func (m *Model) tryNavigateTab(next int) bool {
	if next < 0 || next >= len(tabItems) {
		return false
	}
	if requiresLoadedRace(section(next)) && !m.hasLoadedRace() {
		m.status = "Press Enter on the Races tab to load a race first."
		if m.section != sectionRaces {
			m.tabCursor = int(sectionRaces)
			m.section = sectionRaces
			m.refreshContent(true)
		}
		return false
	}
	m.tabCursor = next
	m.section = section(next)
	m.onTabChange()
	m.refreshContent(true)
	return true
}

func (m *Model) clampRaceCursor() {
	if len(m.raceList) == 0 {
		m.raceListCursor = 0
		return
	}
	if m.raceListCursor < 0 {
		m.raceListCursor = 0
	}
	if m.raceListCursor >= len(m.raceList) {
		m.raceListCursor = len(m.raceList) - 1
	}
}

func (m Model) raceTabEnabled(i int) bool {
	if !requiresLoadedRace(section(i)) {
		return true
	}
	return m.hasLoadedRace()
}

func (m Model) loadedRaceName() string {
	if race, ok := m.selectedRace(); ok {
		return race.Name
	}
	return strings.TrimSpace(m.selectedRaceID)
}
