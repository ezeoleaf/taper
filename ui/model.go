package ui

import (
	"fmt"
	"strings"
	"time"

	"taper/storage"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type section int

const (
	sectionRaces section = iota
	sectionCalendar
	sectionJournal
	sectionChecklist
	sectionPlan
)

var tabItems = []string{"Races", "Calendar", "Journal", "Checklist", "Plan"}

const (
	chromeAppPadV       = 2
	chromeAppPadH       = 4
	chromeHeaderReserve = 2
	chromeFooterReserve = 3
)

type raceListEntry struct {
	race storage.Race
	past bool
}

type Model struct {
	width  int
	height int

	tabCursor int
	section   section

	db             storage.Database
	raceList       []raceListEntry
	raceListCursor int
	selectedRaceID string

	checklistItems  []storage.CheckItem
	checklistCursor int

	planDocCursor int

	calendarMonth time.Time

	readMode bool
	readKind storage.DocKind

	mainViewport viewport.Model
	lastContent string

	bundle *raceBundle

	raceForm     raceForm
	completeForm completeForm
	formMode     string // "", "new", "edit", "complete", "delete"
	editingID    string

	searchInput  textinput.Model
	searchActive bool

	status string
}

func NewModel() (Model, error) {
	db, err := storage.Load()
	if err != nil {
		return Model{}, err
	}

	m := Model{
		db:           db,
		mainViewport: viewport.New(0, 0),
		raceForm:     newRaceForm(),
		completeForm: newCompleteForm(),
		readKind:     storage.DocLog,
	}
	si := textinput.New()
	si.Placeholder = "filter races…"
	si.CharLimit = 60
	si.Width = 40
	m.searchInput = si
	now := time.Now()
	m.calendarMonth = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	m.rebuildRaceList()
	m.clampRaceCursor()
	m.refreshContent(true)
	return m, nil
}

func (m *Model) rebuildRaceList() {
	query := ""
	if m.searchActive || strings.TrimSpace(m.searchInput.Value()) != "" {
		query = m.searchInput.Value()
	}
	upcoming, past := storage.PartitionRaces(m.db.Races, time.Now())
	m.raceList = m.raceList[:0]
	for _, r := range upcoming {
		if storage.RaceMatchesFilter(r, query) {
			m.raceList = append(m.raceList, raceListEntry{race: r, past: false})
		}
	}
	for _, r := range past {
		if storage.RaceMatchesFilter(r, query) {
			m.raceList = append(m.raceList, raceListEntry{race: r, past: true})
		}
	}
	m.clampRaceCursor()
	if m.selectedRaceID != "" {
		found := false
		for _, e := range m.raceList {
			if e.race.ID == m.selectedRaceID {
				found = true
				break
			}
		}
		if !found {
			m.clearRaceLoad()
		}
	}
}

func (m *Model) selectedRace() (storage.Race, bool) {
	for _, r := range m.db.Races {
		if r.ID == m.selectedRaceID {
			return r, true
		}
	}
	return storage.Race{}, false
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) inOverlay() bool {
	return m.formMode != "" || m.searchActive
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m = m.applyLayout(msg.Width, msg.Height)
		m.ensureBundleRendered()
		m.refreshContent(true)

	case editorDoneMsg:
		if msg.err != nil {
			m.status = formatEditorErr(msg.err)
		} else {
			m.status = "Editor closed — reloaded."
		}
		m.reloadBundle()
		m.refreshContent(true)

	case tea.MouseMsg:
		if !m.inOverlay() {
			m.mainViewport, cmd = m.mainViewport.Update(msg)
		}

	case tea.KeyMsg:
		if m.formMode == "delete" {
			return m.handleDeleteConfirmKeys(msg)
		}
		if m.searchActive {
			return m.handleSearchKeys(msg)
		}
		if m.inOverlay() {
			switch m.formMode {
			case "new", "edit":
				return m.handleRaceFormKeys(msg)
			case "complete":
				return m.handleCompleteFormKeys(msg)
			}
		}
		m, cmd = m.handleKeys(msg)

	default:
		if m.searchActive {
			var c tea.Cmd
			m.searchInput, c = m.searchInput.Update(msg)
			m.rebuildRaceList()
			m.clampRaceCursor()
			m.refreshContent(true)
			return m, c
		}
		if m.formMode == "new" || m.formMode == "edit" {
			idx := m.raceForm.fieldIndex()
			m.raceForm.inputs[idx], cmd = m.raceForm.inputs[idx].Update(msg)
		} else if m.formMode == "complete" && m.completeForm.cursor > 0 {
			idx := m.completeForm.cursor
			m.completeForm.inputs[idx], cmd = m.completeForm.inputs[idx].Update(msg)
		}
	}

	if !m.inOverlay() {
		m.refreshContent(false)
	}
	return m, cmd
}

func (m Model) handleKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	key := msg.String()

	if m.shouldViewportScroll(key) {
		var vpCmd tea.Cmd
		m.mainViewport, vpCmd = m.mainViewport.Update(msg)
		return m, vpCmd
	}

	switch key {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "left", "h":
		if m.tabCursor > 0 {
			m.tryNavigateTab(m.tabCursor - 1)
		}
		return m, nil

	case "right", "l":
		if m.tabCursor < len(tabItems)-1 {
			m.tryNavigateTab(m.tabCursor + 1)
		}
		return m, nil

	case "e":
		return m, m.openCurrentEditorCmd()

	case "u":
		if m.section == sectionRaces && m.hasLoadedRace() {
			m.openEditRaceForm()
			return m, textinput.Blink
		}
		if m.section == sectionRaces {
			m.status = "Press Enter on the Races tab to load a race first."
		}
		return m, nil

	case "c":
		if m.section == sectionRaces && m.hasLoadedRace() {
			m.openCompleteForm()
			return m, textinput.Blink
		}
		if m.section == sectionRaces {
			m.status = "Press Enter on the Races tab to load a race first."
		}
		return m, nil

	case "r":
		if m.section == sectionChecklist {
			m.readMode = !m.readMode
			m.mainViewport.GotoTop()
			m.refreshContent(true)
		}
		return m, nil

	case "s":
		if m.section == sectionRaces {
			m.readMode = !m.readMode || m.readKind != storage.DocStrategy
			m.readKind = storage.DocStrategy
			m.mainViewport.GotoTop()
			m.refreshContent(true)
		}
		return m, nil

	case "n":
		if m.section == sectionRaces {
			m.openNewRaceForm()
			return m, textinput.Blink
		}
		return m, nil

	case "/":
		if m.section == sectionRaces {
			m.searchActive = true
			m.searchInput.Focus()
			return m, textinput.Blink
		}
		return m, nil

	case "d":
		if m.section == sectionRaces && m.hasLoadedRace() {
			m.formMode = "delete"
			return m, nil
		}
		if m.section == sectionRaces {
			m.status = "Press Enter on the Races tab to load a race first."
		}
		return m, nil

	case "x":
		if m.section == sectionRaces && m.hasLoadedRace() {
			path, err := storage.ExportMarkdownPacket(m.db, m.selectedRaceID)
			if err != nil {
				m.status = err.Error()
			} else {
				m.status = "Exported packet → " + path
			}
		}
		return m, nil

	case "X":
		if m.section == sectionRaces && m.hasLoadedRace() {
			path, err := storage.ExportPDFPacket(m.db, m.selectedRaceID)
			if err != nil {
				m.status = err.Error()
			} else {
				m.status = "Exported PDF → " + path
			}
		}
		return m, nil

	case ",", "<":
		if m.section == sectionCalendar {
			m.calendarMonth = m.calendarMonth.AddDate(0, -1, 0)
			m.mainViewport.GotoTop()
			m.refreshContent(true)
		}
		return m, nil

	case ".", ">":
		if m.section == sectionCalendar {
			m.calendarMonth = m.calendarMonth.AddDate(0, 1, 0)
			m.mainViewport.GotoTop()
			m.refreshContent(true)
		}
		return m, nil

	case "j", "down":
		m.handleDown()
		m.refreshContent(true)
		return m, nil

	case "k", "up":
		m.handleUp()
		m.refreshContent(true)
		return m, nil

	case "enter":
		if m.section == sectionCalendar {
			upcoming := storage.UpcomingCountdown(m.db.Races, time.Now(), 1)
			if len(upcoming) > 0 {
				for i, e := range m.raceList {
					if e.race.ID == upcoming[0].ID {
						m.raceListCursor = i
						break
					}
				}
				if err := m.commitRaceSelection(); err != nil {
					m.status = err.Error()
					return m, nil
				}
				m.tabCursor = int(sectionJournal)
				m.section = sectionJournal
				m.onTabChange()
				m.refreshContent(true)
				m.status = "Loaded " + upcoming[0].Name
			}
			return m, nil
		}
		if m.section == sectionRaces && len(m.raceList) > 0 {
			if err := m.commitRaceSelection(); err != nil {
				m.status = err.Error()
				return m, nil
			}
			m.tryNavigateTab(int(sectionJournal))
			if race, ok := m.selectedRace(); ok {
				m.status = "Loaded " + race.Name + " — journal open"
			}
		}
		return m, nil

	case " ":
		if m.section == sectionChecklist && !m.readMode {
			return m.toggleChecklistItem()
		}
		return m, nil
	}
	return m, nil
}

func (m Model) handleRaceFormKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.formMode = ""
		m.raceForm.blurAll()
		m.refreshContent(true)
		return m, nil
	case "tab":
		m.raceForm.nextField()
		return m, textinput.Blink
	case "shift+tab":
		m.raceForm.prevField()
		return m, textinput.Blink
	case "]":
		m.raceForm.nextSection()
		return m, textinput.Blink
	case "[":
		m.raceForm.prevSection()
		return m, textinput.Blink
	case "enter":
		return m.submitRaceForm()
	}
	idx := m.raceForm.fieldIndex()
	var c tea.Cmd
	m.raceForm.inputs[idx], c = m.raceForm.inputs[idx].Update(msg)
	return m, tea.Batch(c, textinput.Blink)
}

func (m Model) handleSearchKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searchActive = false
		m.searchInput.SetValue("")
		m.searchInput.Blur()
		m.rebuildRaceList()
		m.clampRaceCursor()
		m.refreshContent(true)
		return m, nil
	case "enter":
		m.searchActive = false
		m.searchInput.Blur()
		m.rebuildRaceList()
		m.clampRaceCursor()
		m.refreshContent(true)
		return m, nil
	}
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.rebuildRaceList()
	m.clampRaceCursor()
	m.refreshContent(true)
	return m, tea.Batch(cmd, textinput.Blink)
}

func (m Model) handleDeleteConfirmKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		name := ""
		if r, ok := m.selectedRace(); ok {
			name = r.Name
		}
		id := m.selectedRaceID
		if err := storage.DeleteRace(&m.db, id); err != nil {
			m.status = err.Error()
		} else {
			m.status = fmt.Sprintf("Deleted %q.", name)
			m.rebuildRaceList()
			m.clearRaceLoad()
		}
		m.formMode = ""
		m.refreshContent(true)
		return m, nil
	case "n", "N", "esc":
		m.formMode = ""
		m.refreshContent(true)
		return m, nil
	}
	return m, nil
}

func (m Model) handleCompleteFormKeys(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.formMode = ""
		m.completeForm.blurAll()
		m.refreshContent(true)
		return m, nil
	case "tab":
		m.completeForm.nextField()
		return m, textinput.Blink
	case "shift+tab":
		m.completeForm.prevField()
		return m, textinput.Blink
	case " ":
		if m.completeForm.cursor == 0 {
			m.completeForm.cycleOutcome()
			return m, nil
		}
	case "enter":
		return m.submitCompleteForm()
	}
	if m.completeForm.cursor == 0 {
		return m, nil
	}
	idx := m.completeForm.cursor
	var c tea.Cmd
	m.completeForm.inputs[idx], c = m.completeForm.inputs[idx].Update(msg)
	return m, tea.Batch(c, textinput.Blink)
}

func (m *Model) onTabChange() {
	switch m.section {
	case sectionJournal:
		m.readMode = true
		m.readKind = storage.DocLog
	case sectionCalendar:
		m.readMode = false
	case sectionChecklist:
		m.readMode = false
		m.readKind = storage.DocPacking
	case sectionPlan:
		m.readMode = true
		if m.planDocCursor >= len(storage.PlanDocs) {
			m.planDocCursor = 0
		}
		m.readKind = storage.PlanDocs[m.planDocCursor].Kind
	case sectionRaces:
		m.readMode = false
		m.readKind = storage.DocStrategy
	}
	m.mainViewport.GotoTop()
}

func (m Model) shouldViewportScroll(key string) bool {
	switch key {
	case "pgup", "pgdown", "ctrl+u", "ctrl+d", "home", "end":
		return true
	}
	if key == "j" || key == "down" || key == "k" || key == "up" {
		switch m.section {
		case sectionJournal, sectionCalendar:
			return true
		case sectionPlan:
			return false
		case sectionChecklist:
			return m.readMode
		case sectionRaces:
			return m.readMode && m.readKind == storage.DocStrategy
		}
	}
	return false
}

func (m *Model) handleDown() {
	switch m.section {
	case sectionRaces:
		if len(m.raceList) > 0 && m.raceListCursor < len(m.raceList)-1 {
			m.raceListCursor++
			m.clampRaceCursor()
		}
	case sectionChecklist:
		if !m.readMode && len(m.checklistItems) > 0 && m.checklistCursor < len(m.checklistItems)-1 {
			m.checklistCursor++
		}
	case sectionPlan:
		if m.planDocCursor < len(storage.PlanDocs)-1 {
			m.planDocCursor++
			m.readKind = storage.PlanDocs[m.planDocCursor].Kind
			m.mainViewport.GotoTop()
		}
	case sectionCalendar, sectionJournal:
		m.mainViewport.LineDown(1)
	}
}

func (m *Model) handleUp() {
	switch m.section {
	case sectionRaces:
		if m.raceListCursor > 0 {
			m.raceListCursor--
			m.clampRaceCursor()
		}
	case sectionChecklist:
		if !m.readMode && m.checklistCursor > 0 {
			m.checklistCursor--
		}
	case sectionPlan:
		if m.planDocCursor > 0 {
			m.planDocCursor--
			m.readKind = storage.PlanDocs[m.planDocCursor].Kind
			m.mainViewport.GotoTop()
		}
	case sectionCalendar, sectionJournal:
		m.mainViewport.LineUp(1)
	}
}

func (m *Model) refreshContent(force bool) {
	content := m.renderMainContent()
	if !force && content == m.lastContent {
		return
	}
	m.lastContent = content
	m.mainViewport.SetContent(content)
}

func (m Model) toggleChecklistItem() (Model, tea.Cmd) {
	if !m.hasLoadedRace() || len(m.checklistItems) == 0 {
		return m, nil
	}
	_, err := storage.ToggleChecklistItem(m.selectedRaceID, m.checklistCursor)
	if err != nil {
		m.status = err.Error()
		return m, nil
	}
	content, err := storage.ReadFile(m.selectedRaceID, storage.DocPacking)
	if err != nil {
		m.status = err.Error()
		return m, nil
	}
	m.bundle.applyPacking(content, m.bundleWidth())
	m.checklistItems = m.bundle.checklist
	m.status = "Toggled checklist item."
	m.refreshContent(true)
	return m, nil
}

func (m *Model) openCurrentEditorCmd() tea.Cmd {
	if !m.hasLoadedRace() {
		m.status = "Press Enter on the Races tab to load a race first."
		return nil
	}
	kind := m.currentDocKind()
	path, err := storage.EditorPath(m.selectedRaceID, kind)
	if err != nil {
		m.status = err.Error()
		return nil
	}
	m.status = "Opening " + editorName(m.db.Config.DefaultEditor) + "…"
	return OpenEditorCmd(path, m.db.Config.DefaultEditor)
}

func (m Model) currentDocKind() storage.DocKind {
	switch m.section {
	case sectionJournal:
		return storage.DocLog
	case sectionChecklist:
		return storage.DocPacking
	case sectionPlan:
		if m.planDocCursor >= 0 && m.planDocCursor < len(storage.PlanDocs) {
			return storage.PlanDocs[m.planDocCursor].Kind
		}
		return storage.DocStrategy
	default:
		if m.readMode {
			return m.readKind
		}
		return storage.DocStrategy
	}
}

func (m *Model) openNewRaceForm() {
	m.formMode = "new"
	m.editingID = ""
	m.raceForm.clear()
	m.raceForm.inputs[1].SetValue(time.Now().Format("2006-01-02"))
	m.raceForm.focusCurrent()
}

func (m *Model) openEditRaceForm() {
	race, ok := m.selectedRace()
	if !ok {
		return
	}
	m.formMode = "edit"
	m.editingID = race.ID
	m.raceForm.loadRace(race)
}

func (m *Model) openCompleteForm() {
	race, ok := m.selectedRace()
	if !ok {
		return
	}
	m.formMode = "complete"
	m.completeForm.loadRace(race)
}

func (m Model) submitRaceForm() (Model, tea.Cmd) {
	in := m.raceForm.toInput()
	switch m.formMode {
	case "new":
		race, err := storage.CreateRace(&m.db, in)
		if err != nil {
			m.status = err.Error()
			return m, nil
		}
		m.formMode = ""
		m.raceForm.blurAll()
		m.rebuildRaceList()
		for i, e := range m.raceList {
			if e.race.ID == race.ID {
				m.raceListCursor = i
				break
			}
		}
		if err := m.commitRaceSelection(); err != nil {
			m.status = err.Error()
			return m, nil
		}
		m.refreshContent(true)
		m.status = fmt.Sprintf("Created race %q.", race.Name)
	case "edit":
		race, err := storage.UpdateRace(&m.db, m.editingID, in)
		if err != nil {
			m.status = err.Error()
			return m, nil
		}
		m.formMode = ""
		m.raceForm.blurAll()
		m.rebuildRaceList()
		for i, e := range m.raceList {
			if e.race.ID == race.ID {
				m.raceListCursor = i
				break
			}
		}
		if m.hasLoadedRace() && m.selectedRaceID == race.ID {
			m.reloadBundle()
		}
		m.refreshContent(true)
		m.status = fmt.Sprintf("Updated race %q.", race.Name)
	}
	return m, nil
}

func (m Model) submitCompleteForm() (Model, tea.Cmd) {
	in := m.completeForm.toInput()
	race, err := storage.CompleteRace(&m.db, m.selectedRaceID, in)
	if err != nil {
		m.status = err.Error()
		return m, nil
	}
	m.formMode = ""
	m.completeForm.blurAll()
	m.rebuildRaceList()
	if m.hasLoadedRace() {
		m.reloadBundle()
	}
	m.refreshContent(true)
	m.status = fmt.Sprintf("Recorded %s for %q.", race.Status, race.Name)
	return m, nil
}

func (m Model) applyLayout(w, h int) Model {
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}
	m.width = w
	m.height = h
	contentW := max(40, w-chromeAppPadH)
	innerH := max(8, h-chromeAppPadV)
	vpH := max(6, innerH-chromeHeaderReserve-chromeFooterReserve)
	m.mainViewport.Width = contentW
	m.mainViewport.Height = vpH
	return m
}

func (m Model) View() string {
	switch m.formMode {
	case "new":
		return appStyle.Render(m.renderRaceFormOverlay("New Race"))
	case "edit":
		return appStyle.Render(m.renderRaceFormOverlay("Edit Race"))
	case "complete":
		return appStyle.Render(m.renderCompleteFormOverlay())
	case "delete":
		return appStyle.Render(m.renderDeleteConfirm())
	}
	header := m.renderHeaderTabs()
	body := m.mainViewport.View()
	footer := m.renderFooter()
	stack := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
	return appStyle.Render(stack)
}

func (m Model) renderHeaderTabs() string {
	var tabs []string
	for i, item := range tabItems {
		style := tabStyle
		if i == m.tabCursor {
			style = activeTabStyle
		} else if !m.raceTabEnabled(i) {
			style = disabledTabStyle
		}
		tabs = append(tabs, style.Render(item))
	}
	bar := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	w := max(20, m.width-chromeAppPadH)
	subtitle := ""
	switch m.section {
	case sectionCalendar:
		subtitle = mutedStyle.Render("  ·  " + m.calendarMonth.Format("Jan 2006"))
	case sectionRaces:
		// no subtitle
	default:
		if race, ok := m.selectedRace(); ok {
			subtitle = mutedStyle.Render("  ·  " + race.Name)
		}
	}
	return headerBarStyle.Width(w).Render(bar + subtitle)
}

func (m *Model) renderMainContent() string {
	switch m.section {
	case sectionRaces:
		return m.renderRaces()
	case sectionCalendar:
		return renderCalendar(m.calendarMonth, m.db.Races, m.mainViewport.Width)
	case sectionJournal:
		return m.renderJournal()
	case sectionChecklist:
		return m.renderChecklist()
	case sectionPlan:
		return m.renderPlan()
	default:
		return ""
	}
}

func (m *Model) renderRaces() string {
	if len(m.raceList) == 0 {
		return titleStyle.Render("Races") + "\n\n" +
			mutedStyle.Render("No races match. Press n to add one or / to change filter.")
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("Upcoming & Past Races"))
	b.WriteString("\n\n")
	if m.searchActive {
		b.WriteString(mutedStyle.Render("Filter: "))
		b.WriteString(m.searchInput.View())
		b.WriteString("\n\n")
	} else if q := strings.TrimSpace(m.searchInput.Value()); q != "" {
		b.WriteString(mutedStyle.Render("Filter: "+q+"  (/ edit · Esc clear)"))
		b.WriteString("\n\n")
	}
	b.WriteString(mutedStyle.Render("j/k highlight · Enter load race · / filter · n new · u edit · c complete · d delete · x export · X PDF"))
	b.WriteString("\n\n")

	lastPast := false
	for i, e := range m.raceList {
		if e.past && !lastPast {
			b.WriteString("\n" + mutedStyle.Render("── Past ──") + "\n")
			lastPast = true
		}
		line := m.formatRaceLine(e.race, e.past)
		if m.hasLoadedRace() && e.race.ID == m.selectedRaceID && i != m.raceListCursor {
			line += "  " + checkStyle.Render("loaded")
		}
		prefix := "   "
		if i == m.raceListCursor {
			prefix = " › "
			b.WriteString(navItemActiveStyle.Render(prefix + line))
		} else {
			b.WriteString(navItemStyle.Render(prefix + line))
		}
		b.WriteString("\n")
	}

	if m.hasLoadedRace() {
		if race, ok := m.selectedRace(); ok {
			b.WriteString("\n" + subtleBoxStyle.Render(m.renderRaceDetail(race)))
		}
	} else if len(m.raceList) > 0 {
		b.WriteString("\n" + mutedStyle.Render("Press Enter to load the highlighted race into memory (unlocks Journal, Checklist, Plan)."))
	}

	if m.readMode && m.readKind == storage.DocStrategy {
		b.WriteString("\n\n" + titleStyle.Render("Strategy") + "\n\n")
		b.WriteString(m.cachedMarkdown(storage.DocStrategy))
	}
	return b.String()
}

func (m *Model) renderRaceDetail(r storage.Race) string {
	lines := []string{titleStyle.Render(r.Name), fmt.Sprintf("Date:     %s", r.Date)}
	if d := daysUntil(r); d >= 0 && r.IsUpcoming(time.Now()) {
		if d == 0 {
			lines = append(lines, warnStyle.Render("          Race day!"))
		} else {
			lines = append(lines, fmt.Sprintf("          in %d days", d))
		}
	}
	add := func(label, val string) {
		if val != "" {
			lines = append(lines, fmt.Sprintf("%-10s %s", label+":", val))
		}
	}
	add("Distance", r.Distance)
	add("Type", storage.TypeLabel(r.Type))
	add("Location", r.Location)
	add("Goal", r.GoalTime)
	add("Status", r.Status)
	add("Result", r.Result)
	add("Reg", r.Confirmation)
	add("Bib", r.BibPickup)
	add("Hotel", r.Hotel)
	add("Health", r.HealthNotes)
	add("Notes", ansi.Truncate(r.Notes, 50, "…"))
	lines = append(lines, mutedStyle.Render("u edit · c complete race"))
	return strings.Join(lines, "\n")
}

func (m Model) formatRaceLine(r storage.Race, past bool) string {
	date := r.Date
	if date == "" {
		date = "????-??-??"
	}
	meta := strings.TrimSpace(r.Distance)
	if loc := strings.TrimSpace(r.Location); loc != "" {
		if meta != "" {
			meta += " · "
		}
		meta += loc
	}
	if t := storage.TypeLabel(r.Type); t != "" {
		if meta != "" {
			meta += " · "
		}
		meta += t
	}
	if st := strings.TrimSpace(r.Status); st != "" && st != "planned" {
		if meta != "" {
			meta += " · "
		}
		meta += st
	}
	line := fmt.Sprintf("%s  %s", date, r.Name)
	if meta != "" {
		line += "  (" + meta + ")"
	}
	if past && r.Result != "" {
		line += "  " + goodStyle.Render(r.Result)
		return line
	}
	if days := daysUntil(r); !past {
		if days == 0 {
			line += "  " + warnStyle.Render("TODAY")
		} else if days > 0 && days <= 14 {
			line += fmt.Sprintf("  %s", warnStyle.Render(fmt.Sprintf("in %dd", days)))
		}
	}
	return line
}

func daysUntil(r storage.Race) int {
	d := r.RaceDate()
	if d.IsZero() {
		return -1
	}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	raceDay := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, now.Location())
	return int(raceDay.Sub(today).Hours() / 24)
}

func (m *Model) renderChecklist() string {
	if !m.hasLoadedRace() {
		return titleStyle.Render("Checklist") + "\n\n" + mutedStyle.Render("Load a race with Enter on the Races tab first.")
	}
	if m.readMode {
		return titleStyle.Render("Checklist") + "\n\n" +
			mutedStyle.Render("r interactive · e edit · Space toggles in interactive mode") + "\n\n" +
			m.cachedMarkdown(storage.DocPacking)
	}
	if len(m.checklistItems) == 0 {
		return titleStyle.Render("Checklist") + "\n\n" +
			mutedStyle.Render("No checklist items. Press e to edit packing.md.")
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("Checklist") + "\n\n")
	b.WriteString(mutedStyle.Render("Space toggle · j/k move · r read mode · e edit"))
	b.WriteString("\n\n")

	done := 0
	for _, it := range m.checklistItems {
		if it.Checked {
			done++
		}
	}
	b.WriteString(subtleBoxStyle.Render(fmt.Sprintf("%d / %d packed", done, len(m.checklistItems))) + "\n\n")

	for i, item := range m.checklistItems {
		box := "[ ]"
		textStyle := navItemStyle
		if item.Checked {
			box = "[x]"
			textStyle = goodStyle
		}
		line := fmt.Sprintf("%s %s", box, item.Text)
		prefix := "   "
		if i == m.checklistCursor {
			b.WriteString(navItemActiveStyle.Render(" › " + line))
		} else {
			b.WriteString(textStyle.Render(prefix + line))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m *Model) renderJournal() string {
	if !m.hasLoadedRace() {
		return titleStyle.Render("Journal") + "\n\n" + mutedStyle.Render("Load a race with Enter on the Races tab first.")
	}
	head := titleStyle.Render("Journal")
	if race, ok := m.selectedRace(); ok {
		head += mutedStyle.Render(" — " + race.Name)
	}
	return head + "\n\n" + mutedStyle.Render("e edit log.md · j/k scroll") + "\n\n" +
		m.cachedMarkdown(storage.DocLog)
}

func (m *Model) renderPlan() string {
	if !m.hasLoadedRace() {
		return titleStyle.Render("Plan") + "\n\n" + mutedStyle.Render("Load a race with Enter on the Races tab first.")
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("Race Plan") + "\n\n")
	b.WriteString(mutedStyle.Render("j/k pick section · e edit · PgUp/PgDn scroll preview"))
	b.WriteString("\n\n")

	for i, doc := range storage.PlanDocs {
		prefix := "   "
		if i == m.planDocCursor {
			prefix = " › "
			b.WriteString(navItemActiveStyle.Render(prefix + doc.Title))
		} else {
			b.WriteString(navItemStyle.Render(prefix + doc.Title))
		}
		b.WriteString(mutedStyle.Render("  (" + doc.File + ")"))
		b.WriteString("\n")
	}

	b.WriteString("\n\n")
	if m.planDocCursor >= 0 && m.planDocCursor < len(storage.PlanDocs) {
		b.WriteString(titleStyle.Render(storage.PlanDocs[m.planDocCursor].Title))
		b.WriteString("\n\n")
		b.WriteString(m.cachedMarkdown(storage.PlanDocs[m.planDocCursor].Kind))
	}
	return b.String()
}

func (m Model) renderRaceFormOverlay(title string) string {
	spec := m.raceForm.sectionSpec()
	head := titleStyle.Render(title)
	section := mutedStyle.Render(fmt.Sprintf("Section: %s  ([ ] prev/next section · Tab field · Enter save · Esc cancel)", spec.name))
	var b strings.Builder
	b.WriteString(head + "\n\n" + section + "\n\n")
	labels := m.raceForm.labels()
	for i := 0; i < spec.count; i++ {
		idx := spec.start + i
		line := fmt.Sprintf("%-14s %s", labels[idx]+":", m.raceForm.inputs[idx].View())
		if i == m.raceForm.field {
			b.WriteString(navItemActiveStyle.Render("› " + line))
		} else {
			b.WriteString(navItemStyle.Render("  " + line))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m Model) renderCompleteFormOverlay() string {
	head := titleStyle.Render("Complete Race")
	hint := mutedStyle.Render("Space cycle outcome · Tab field · Enter save · Esc cancel")
	var b strings.Builder
	b.WriteString(head + "\n\n" + hint + "\n\n")
	labels := m.completeForm.fieldLabels()
	for i, label := range labels {
		var line string
		if i == 0 {
			style := navItemStyle
			if m.completeForm.cursor == 0 {
				style = navItemActiveStyle
			}
			line = style.Render(fmt.Sprintf("%-16s %s (Space to change)", label+":", m.completeForm.outcomeLabel()))
		} else {
			line = fmt.Sprintf("%-16s %s", label+":", m.completeForm.inputs[i].View())
			if i == m.completeForm.cursor {
				line = navItemActiveStyle.Render("› " + line)
			} else {
				line = navItemStyle.Render("  " + line)
			}
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

func (m Model) renderDeleteConfirm() string {
	name := m.selectedRaceID
	if r, ok := m.selectedRace(); ok {
		name = r.Name
	}
	head := titleStyle.Render("Delete Race?")
	body := warnStyle.Render(fmt.Sprintf("Permanently delete %q and all markdown files?", name))
	hint := mutedStyle.Render("y confirm · n cancel · Esc cancel")
	return head + "\n\n" + body + "\n\n" + hint
}

func (m Model) renderFooter() string {
	w := max(20, m.width-chromeAppPadH)
	keys := "h/l tabs · j/k · / filter · u edit · c complete · d delete · x export · X PDF · q quit"
	line1 := ansi.Truncate(keys, w, "…")
	st := strings.TrimSpace(m.status)
	if st == "" {
		st = " "
	}
	line2 := ansi.Truncate(st, w, "…")
	return footerStyle.Width(w).Render(mutedStyle.Render(line1) + "\n" + mutedStyle.Render(line2))
}
