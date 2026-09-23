package tui

import (
	"fmt"
	"image"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type filterKind int

const (
	noFilter filterKind = iota
	difficultyFilter
	statusFilter
	tagsFilter
)

const (
	filterApplyLabel  = "[Apply]"
	filterCancelLabel = "[Cancel]"
)

func (f filterKind) title() string {
	switch f {
	case difficultyFilter:
		return "Difficulty"
	case statusFilter:
		return "Status"
	case tagsFilter:
		return "Tags"
	default:
		return ""
	}
}

type filterItem struct {
	filterOption
	keywords string
	checked  bool
}

func (i *filterItem) FilterValue() string { return i.label + " " + i.value + " " + i.keywords }

func (m *model) openFilter(kind filterKind) tea.Cmd {
	if m.filter == kind {
		m.closeFilter()
		return m.syncPreview()
	}
	m.filter = kind
	m.search.Blur()
	m.filterSearch.Blur()
	m.filterSearch.Reset()
	m.populateFilter()
	m.resize()
	if kind == tagsFilter && !m.tagsLoaded && !m.tagsLoading {
		return m.loadTags()
	}
	return nil
}

func (m *model) closeFilter() {
	m.filter = noFilter
	m.filterSearch.Blur()
	m.filterSearch.Reset()
	m.resize()
}

func (m *model) populateFilter() {
	m.filterItems = nil
	if m.filter == tagsFilter {
		for _, tag := range m.availableTags {
			m.filterItems = append(m.filterItems, &filterItem{
				filterOption: filterOption{label: tag.Name, value: tag.Slug},
				keywords:     tag.NameTranslated, checked: slices.Contains(m.tags, tag.Slug),
			})
		}
	} else {
		options, selected := difficulties, m.difficulty
		if m.filter == statusFilter {
			options, selected = statuses, m.status
		}
		for i, option := range options {
			m.filterItems = append(m.filterItems, &filterItem{filterOption: option, checked: i == selected})
		}
	}
	m.filterChoices()
	for i, entry := range m.filterList.Items() {
		if entry.(*filterItem).checked {
			m.filterList.Select(i)
			break
		}
	}
}

// Filter locally so a delayed search cannot replace a newer selection draft.
func (m *model) filterChoices() {
	items := m.filterItems
	if query := m.filterSearch.Value(); query != "" {
		targets := make([]string, len(items))
		for i, entry := range items {
			targets[i] = entry.FilterValue()
		}
		items = nil
		for _, rank := range list.DefaultFilter(query, targets) {
			items = append(items, m.filterItems[rank.Index])
		}
	}
	m.filterList.SetItems(items)
	m.filterList.ResetSelected()
}

func (m *model) applyFilter() tea.Cmd {
	if m.filter == tagsFilter {
		if m.tagsLoading || m.tagsErr != nil {
			return nil
		}
		m.tags = nil
		for _, entry := range m.filterItems {
			if choice := entry.(*filterItem); choice.checked {
				m.tags = append(m.tags, choice.value)
			}
		}
		if len(m.tags) == len(m.availableTags) {
			m.tags = nil
		}
	} else {
		choice, ok := m.filterList.SelectedItem().(*filterItem)
		if !ok {
			return nil
		}
		options, selected := difficulties, &m.difficulty
		if m.filter == statusFilter {
			options, selected = statuses, &m.status
		}
		*selected = slices.IndexFunc(options, func(option filterOption) bool { return option.value == choice.value })
	}
	m.closeFilter()
	return m.loadQuestions(true)
}

func (m *model) clearFilters() tea.Cmd {
	m.difficulty, m.status = 0, 0
	m.query, m.tags = "", nil
	m.search.Blur()
	m.search.Reset()
	m.closeFilter()
	return m.loadQuestions(true)
}

func (m *model) updateFilter(msg tea.Msg) tea.Cmd {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			m.closeFilter()
			return m.syncPreview()
		case "enter":
			return m.applyFilter()
		case "tab", "shift+tab":
			m.closeFilter()
			m.previewFocused = m.previewVisible() && !m.previewFocused
			return m.syncPreview()
		case "up", "down", "pgup", "pgdown":
			m.filterSearch.Blur()
			var cmd tea.Cmd
			m.filterList, cmd = m.filterList.Update(msg)
			return cmd
		}
	}
	if m.filterSearch.Focused() {
		var cmd tea.Cmd
		query := m.filterSearch.Value()
		m.filterSearch, cmd = m.filterSearch.Update(msg)
		if m.filterSearch.Value() != query {
			m.filterChoices()
		}
		return cmd
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "q":
			m.closeFilter()
			return m.syncPreview()
		case "left", "right":
			m.closeFilter()
			m.previewFocused = keyMsg.String() == "right" && m.previewVisible()
			return m.syncPreview()
		case "d":
			return m.openFilter(difficultyFilter)
		case "s":
			return m.openFilter(statusFilter)
		case "t":
			return m.openFilter(tagsFilter)
		case "c":
			return m.clearFilters()
		case "r":
			if m.filter == tagsFilter && m.tagsErr != nil && !m.tagsLoading {
				return m.loadTags()
			}
		case "/":
			if m.filter == tagsFilter {
				return m.filterSearch.Focus()
			}
			return nil
		case " ":
			if choice, ok := m.filterList.SelectedItem().(*filterItem); ok && m.filter == tagsFilter {
				choice.checked = !choice.checked
			}
			return nil
		}
	}
	var cmd tea.Cmd
	m.filterList, cmd = m.filterList.Update(msg)
	return cmd
}

type filterControl struct {
	kind   filterKind
	label  string
	bounds image.Rectangle
}

// Wrap whole controls, using the same rectangles for drawing and clicking.
func (m *model) filterControls() []filterControl {
	tags := "All"
	if len(m.tags) > 0 {
		tags = fmt.Sprintf("%d selected", len(m.tags))
	}
	controls := []filterControl{
		{kind: difficultyFilter, label: "[d Difficulty: " + difficulties[m.difficulty].label + " ▾]"},
		{kind: statusFilter, label: "[s Status: " + statuses[m.status].label + " ▾]"},
		{kind: tagsFilter, label: "[t Tags: " + tags + " ▾]"},
		{kind: noFilter, label: "[c Clear]"},
	}
	width := max(1, m.width-2*pickPadding)
	x, y := pickPadding, 0
	for i := range controls {
		control := &controls[i]
		control.label = strings.TrimRight(fitCell(control.label, min(width, lipgloss.Width(control.label))), " ")
		w := lipgloss.Width(control.label)
		if x+w > pickPadding+width && x > pickPadding {
			x, y = pickPadding, y+1
		}
		control.bounds = image.Rect(x, y, x+w, y+1)
		x += w + 1
	}
	return controls
}

func (m *model) filterBarView() string {
	controls := m.filterControls()
	lines := make([]string, controls[len(controls)-1].bounds.Max.Y)
	for _, control := range controls {
		style := pickMutedStyle
		if control.kind != noFilter && m.filter == control.kind {
			style = pickActiveStyle
		} else if control.kind == difficultyFilter && m.difficulty != 0 ||
			control.kind == statusFilter && m.status != 0 || control.kind == tagsFilter && len(m.tags) > 0 {
			style = pickAccentStyle
		}
		line := &lines[control.bounds.Min.Y]
		*line += strings.Repeat(" ", control.bounds.Min.X-pickPadding-lipgloss.Width(*line)) + style.Render(control.label)
	}
	return strings.Join(lines, "\n")
}

type filterLayout struct {
	bounds  image.Rectangle
	options image.Rectangle
	search  image.Rectangle
	apply   image.Rectangle
	cancel  image.Rectangle
}

func (m *model) dropdownLayout() filterLayout {
	var anchor image.Rectangle
	for _, control := range m.filterControls() {
		if control.kind == m.filter {
			anchor = control.bounds
			break
		}
	}
	width, rows, extraRows := max(26, anchor.Dx()), len(m.filterItems), 2
	if m.filter == tagsFilter {
		width, rows, extraRows = 42, 8, 4
	}
	width = min(width, max(1, m.width-2*pickPadding))
	x := max(pickPadding, min(anchor.Min.X, m.width-pickPadding-width))
	y := anchor.Max.Y
	rows = max(1, min(rows, m.height-pickFooterHeight-y-extraRows))
	bounds := image.Rect(x, y, x+width, y+rows+extraRows)
	layout := filterLayout{
		bounds:  bounds,
		options: image.Rect(x+1, y+1, bounds.Max.X-1, y+1+rows),
	}
	if m.filter == tagsFilter {
		layout.search = image.Rect(x+1, y+1, bounds.Max.X-1, y+2)
		layout.options = layout.options.Add(image.Pt(0, 1))
		buttonY := layout.options.Max.Y
		layout.apply = image.Rect(x+1, buttonY, x+1+lipgloss.Width(filterApplyLabel), buttonY+1)
		layout.cancel = image.Rect(layout.apply.Max.X+1, buttonY, layout.apply.Max.X+1+lipgloss.Width(filterCancelLabel), buttonY+1)
	}
	return layout
}

func (m *model) dropdownView() string {
	layout := m.dropdownLayout()
	width, height := layout.options.Dx(), layout.options.Dy()
	body := m.filterList.View()
	switch {
	case m.filter == tagsFilter && m.tagsLoading:
		body = "Loading tags…"
	case m.filter == tagsFilter && m.tagsErr != nil:
		body = "Could not load tags. Press r to retry."
	case len(m.filterList.VisibleItems()) == 0:
		body = "No matching options."
	}
	body = lipgloss.NewStyle().Width(width).Height(height).MaxWidth(width).MaxHeight(height).Render(body)
	if m.filter == tagsFilter {
		search := pickMutedStyle.Render("/ Find tags")
		if m.filterSearch.Focused() || m.filterSearch.Value() != "" {
			search = m.filterSearch.View()
		}
		selected := 0
		for _, entry := range m.filterItems {
			if entry.(*filterItem).checked {
				selected++
			}
		}
		buttons := filterApplyLabel + " " + filterCancelLabel
		footer := splitLine(buttons, pickMutedStyle.Render(fmt.Sprintf("%d selected", selected)), width)
		body = fitCell(search, width) + "\n" + body + "\n" + footer
	}
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(pickAccent).Render(body)
}

func (m *model) overlayFilter(view string) string {
	bounds := m.dropdownLayout().bounds
	lines := strings.Split(view, "\n")
	for i, row := range strings.Split(m.dropdownView(), "\n") {
		y := bounds.Min.Y + i
		if y >= len(lines) {
			break
		}
		line := fitCell(lines[y], m.width)
		left := fitCell(ansi.Cut(line, 0, bounds.Min.X), bounds.Min.X)
		rightWidth := m.width - bounds.Max.X
		right := ansi.Cut(line, bounds.Max.X, m.width)
		// A wide character crossing the menu edge must leave blank cells,
		// rather than shifting the rest of the background to the right.
		for start := bounds.Max.X; ansi.StringWidth(right) > rightWidth; start++ {
			right = ansi.Cut(line, start+1, m.width)
		}
		right = strings.Repeat(" ", rightWidth-ansi.StringWidth(right)) + right
		lines[y] = left + row + right
	}
	return strings.Join(lines, "\n")
}
