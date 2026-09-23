package tui

import (
	"fmt"
	"image"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"

	"github.com/j178/leetgo/leetcode"
)

const (
	pickMinWidth        = 36
	pickMinHeight       = 12
	pickPreviewMinWidth = 80
	pickPadding         = 1
	pickFooterHeight    = 3
	pickDivider         = " │ "
)

var (
	pickAccent = lipgloss.AdaptiveColor{Light: "#007F8B", Dark: "#22D3EE"}
	pickMuted  = lipgloss.AdaptiveColor{Light: "#667085", Dark: "#8492A6"}
	pickText   = lipgloss.AdaptiveColor{Light: "#172033", Dark: "#E2E8F0"}
	pickStripe = lipgloss.AdaptiveColor{Light: "#F1F5F9", Dark: "#1E2531"}
	pickGreen  = lipgloss.AdaptiveColor{Light: "#15803D", Dark: "#4ADE80"}
	pickYellow = lipgloss.AdaptiveColor{Light: "#A16207", Dark: "#FACC15"}
	pickRed    = lipgloss.AdaptiveColor{Light: "#BE123C", Dark: "#FB7185"}

	pickAccentStyle = lipgloss.NewStyle().Foreground(pickAccent)
	pickMutedStyle  = lipgloss.NewStyle().Foreground(pickMuted)
	pickActiveStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#07151C")).Background(pickAccent).Bold(true)
	pickRowStyle    = lipgloss.NewStyle().Foreground(pickText)
	pickSelectStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#172554", Dark: "#FFFFFF"}).
			Background(lipgloss.AdaptiveColor{Light: "#BFDBFE", Dark: "#1D4ED8"})
)

const pickHelp = `Tab / ← / →      Switch panes
↑ / ↓ or j / k   Navigate / scroll
Enter            Pick question
/                Search
m                Load more

d / s / t        Difficulty / Status / Tags
Space            Toggle tag
Enter / Esc      Apply / cancel filter
c                Clear filters and search

Click to select; wheel to scroll.

Esc / q          Close help
Ctrl+C           Quit`

type rowDelegate struct{}

func (rowDelegate) Height() int                         { return 1 }
func (rowDelegate) Spacing() int                        { return 0 }
func (rowDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }
func (rowDelegate) Render(w io.Writer, m list.Model, index int, entry list.Item) {
	cursor := "  "
	style := pickRowStyle
	if index%2 == 1 {
		style = style.Background(pickStripe)
	}
	if index == m.Index() {
		cursor = "▸ "
		style = pickSelectStyle
	}
	var row string
	switch entry := entry.(type) {
	case *item:
		q := (*leetcode.QuestionData)(entry)
		row = style.Render(cursor) + questionCells(q.QuestionFrontendId, q.GetTitle(), q.Difficulty, q.Status, m.Width()-2, style)
	case *filterItem:
		check := "[ ] "
		if entry.checked {
			check = "[✓] "
		}
		row = style.Render(cursor + fitCell(check+entry.label, m.Width()-2))
	}
	_, _ = fmt.Fprint(w, row)
}

func fitCell(text string, width int) string {
	if width <= 0 {
		return ""
	}
	text = strings.ReplaceAll(text, "\n", " ")
	if lipgloss.Width(text) > width {
		text = truncate.StringWithTail(text, uint(width), "…")
	}
	return text + strings.Repeat(" ", max(0, width-lipgloss.Width(text)))
}

func questionCells(id, title, difficulty, status string, width int, style lipgloss.Style) string {
	idWidth, difficultyWidth, statusWidth := 7, 8, 12
	if width < 62 {
		idWidth, statusWidth = 5, 0
	}
	if width >= 100 {
		idWidth = 12
	}
	difficultyColor := pickAccent
	switch strings.ToUpper(difficulty) {
	case "EASY":
		difficulty, difficultyColor = "Easy", pickGreen
	case "MEDIUM":
		difficulty, difficultyColor = "Medium", pickYellow
	case "HARD":
		difficulty, difficultyColor = "Hard", pickRed
	}
	statusColor := pickMuted
	switch status {
	case "ac", "AC":
		status, statusColor = "✓ Accepted", pickGreen
	case "notac", "TRIED":
		status, statusColor = "• Tried", pickYellow
	case "", "NOT_STARTED":
		status = "—"
	case "STATUS":
		statusColor = pickAccent
	}
	titleWidth := width - idWidth - difficultyWidth - 2
	if statusWidth > 0 {
		titleWidth -= statusWidth + 1
	}
	row := style.Render(fitCell(id, idWidth)+" "+fitCell(title, titleWidth)+" ") +
		style.Foreground(difficultyColor).Render(fitCell(difficulty, difficultyWidth))
	if statusWidth > 0 {
		row += style.Foreground(statusColor).Render(" " + fitCell(status, statusWidth))
	}
	return row
}

func splitLine(left, right string, width int) string {
	rightWidth := lipgloss.Width(right)
	if lipgloss.Width(left)+rightWidth+2 > width {
		return fitCell(left, width)
	}
	return fitCell(left, width-rightWidth) + right
}

func paneTitle(title string, focused bool) string {
	if focused {
		return pickAccentStyle.Bold(true).Render("● " + title)
	}
	return pickMutedStyle.Render("○ " + title)
}

func (m *model) headerView() string {
	width := m.layout().body.Dx()
	search := pickMutedStyle.Render("/ Search by title or question ID")
	if m.query != "" {
		search = "/ " + m.query
	}
	if m.search.Focused() {
		search = m.search.View()
	}
	heading := paneTitle("QUESTIONS", !m.previewFocused)
	columns := "  " + questionCells("ID", "TITLE", "LEVEL", "STATUS", m.list.Width()-2, pickAccentStyle.Bold(true))
	if m.previewVisible() {
		heading = m.joinPreview(fitCell(heading, m.list.Width()), m.previewHeader())
		columns = m.joinPreview(columns, strings.Repeat(" ", m.preview.viewport.Width))
	}
	if m.panel != "" {
		heading = pickAccentStyle.Bold(true).Render(m.panel)
		search = pickMutedStyle.Render("↑↓ scroll · esc to return")
		columns = ""
	}
	return strings.Join([]string{
		m.filterBarView(),
		fitCell(search, width),
		pickMutedStyle.Render(strings.Repeat("─", width)),
		fitCell(heading, width),
		fitCell(columns, width),
	}, "\n")
}

func (m *model) footerView() string {
	width := m.layout().body.Dx()
	var summary string
	count := fmt.Sprintf("%d loaded / %d total", len(m.list.Items()), m.total)
	if len(m.list.Items()) > 0 {
		start, end := m.list.Paginator.GetSliceBounds(len(m.list.Items()))
		count = fmt.Sprintf("%d–%d / %d", start+1, end, m.total)
		if m.hasMore {
			count += "  [m] more"
		}
	}
	if m.loading {
		count = "Loading questions…"
	}
	hints := []key.Binding{pickHint("↑↓", "move"), pickHint("enter", "pick"), pickHint("/", "search")}
	if m.previewVisible() {
		hints = append([]key.Binding{pickHint("tab/←→", "pane")}, hints...)
	}
	if m.previewFocused {
		hints = []key.Binding{pickHint("tab/←", "list"), pickHint("↑↓", "scroll"), pickHint("pgup/pgdn", "page"), pickHint("enter", "pick")}
	}
	hints = append(hints, pickHint("?", "help"), pickHint("q", "quit"))
	if m.notice != nil {
		summary = lipgloss.NewStyle().Foreground(pickYellow).Render("[!] " + m.notice.summary())
	}
	if m.err != nil {
		summary = lipgloss.NewStyle().Foreground(pickRed).Render("Could not load questions: " + m.err.Error() + " · r retry")
	}
	if m.filter != noFilter {
		summary, count = m.filter.title(), fmt.Sprintf("%d options", len(m.filterList.VisibleItems()))
		hints = []key.Binding{pickHint("↑↓", "move"), pickHint("enter", "apply"), pickHint("esc", "cancel")}
		if m.filter == tagsFilter {
			selected := 0
			for _, entry := range m.filterItems {
				if entry.(*filterItem).checked {
					selected++
				}
			}
			summary = fmt.Sprintf("%d selected", selected)
			hints = append([]key.Binding{pickHint("space", "toggle"), pickHint("/", "find")}, hints[1:]...)
			if m.tagsLoading {
				count = "Loading tags…"
			}
			if m.tagsErr != nil {
				summary = lipgloss.NewStyle().Foreground(pickRed).Render("Could not load tags: " + m.tagsErr.Error() + " · r retry")
			}
		}
	}
	if m.search.Focused() {
		hints = []key.Binding{pickHint("enter/esc", "browse results")}
	}
	if m.filterSearch.Focused() {
		hints = []key.Binding{pickHint("↑↓", "move"), pickHint("click", "toggle"), pickHint("enter", "apply"), pickHint("esc", "cancel")}
	}
	if m.panel != "" {
		summary, count = m.panel, fmt.Sprintf("%.0f%%", m.details.ScrollPercent()*100)
		hints = []key.Binding{pickHint("↑↓", "scroll"), pickHint("pgup/pgdn", "page"), pickHint("esc", "back")}
	}
	h := help.New()
	h.Width = width
	h.Styles.ShortKey = pickAccentStyle
	h.Styles.ShortDesc = pickMutedStyle
	return strings.Join([]string{
		pickMutedStyle.Render(strings.Repeat("─", width)),
		splitLine(summary, pickMutedStyle.Render(count), width),
		fitCell(h.ShortHelpView(hints), width),
	}, "\n")
}

func pickHint(keyName, description string) key.Binding {
	return key.NewBinding(key.WithKeys(keyName), key.WithHelp(keyName, description))
}

type pickLayout struct {
	body      image.Rectangle
	questions image.Rectangle
	preview   image.Rectangle
	search    image.Rectangle
}

// Component sizes and mouse hit areas must use the same terminal coordinates.
func (m *model) layout() pickLayout {
	width := max(1, m.width-2*pickPadding)
	controls := m.filterControls()
	toolbarHeight := controls[len(controls)-1].bounds.Max.Y
	headerHeight := toolbarHeight + 4
	height := max(1, m.height-headerHeight-pickFooterHeight)
	body := image.Rect(pickPadding, headerHeight, pickPadding+width, headerHeight+height)
	layout := pickLayout{
		body: body, questions: body,
		search: image.Rect(body.Min.X, toolbarHeight, body.Max.X, toolbarHeight+1),
	}
	if m.width >= pickPreviewMinWidth {
		dividerWidth := lipgloss.Width(pickDivider)
		layout.questions.Max.X = body.Min.X + (width-dividerWidth)*45/100
		layout.preview = body
		layout.preview.Min.X = layout.questions.Max.X + dividerWidth
	}
	return layout
}

func (m *model) resize() {
	layout := m.layout()
	width, height := layout.body.Dx(), layout.body.Dy()
	m.search.Width = max(1, width-3)
	m.list.SetSize(layout.questions.Dx(), height)
	if m.filter != noFilter {
		menu := m.dropdownLayout()
		m.filterList.SetSize(menu.options.Dx(), menu.options.Dy())
		m.filterSearch.Width = max(1, menu.search.Dx()-3)
	}
	m.details.Width, m.details.Height = width, height
	previewWidth := max(1, layout.preview.Dx())
	if m.preview.viewport.Width != previewWidth {
		m.preview.viewport.Width = previewWidth
		if m.width >= pickPreviewMinWidth {
			m.renderPreview()
		}
	}
	if m.width < pickPreviewMinWidth || m.height < pickMinHeight {
		m.previewFocused = false
	}
	m.preview.viewport.Height = height
	m.preview.viewport.SetYOffset(m.preview.viewport.YOffset)
	m.refreshPanel()
}

func (m *model) openPanel(title string) {
	m.panel = title
	m.refreshPanel()
	m.details.GotoTop()
}

func (m *model) refreshPanel() {
	content := pickHelp
	if m.panel == "Messages" {
		content = m.notice.details()
	}
	if m.panel != "" {
		m.details.SetContent(wrap.String(wordwrap.String(content, m.details.Width), m.details.Width))
	}
}

func (m *model) View() string {
	if m.width < pickMinWidth || m.height < pickMinHeight {
		return lipgloss.NewStyle().MaxWidth(max(1, m.width)).MaxHeight(max(1, m.height)).Render(
			"Resize to at least 36 × 12. Ctrl+C to quit.",
		)
	}
	body := m.list.View()
	var empty string
	if len(m.list.Items()) == 0 {
		switch {
		case m.loading:
			empty = "Loading questions…"
		case m.err != nil:
			empty = "Unable to load questions. Press r to retry."
		default:
			empty = "No matching questions. Press c to clear filters."
		}
	}
	bodyWidth := m.list.Width()
	if empty != "" {
		body = lipgloss.Place(bodyWidth, m.list.Height(), lipgloss.Center, lipgloss.Center,
			pickMutedStyle.Render(strings.TrimSpace(fitCell(empty, bodyWidth))))
	}
	if m.previewVisible() {
		body = m.joinPreview(body, m.preview.viewport.View())
	}
	if m.panel != "" {
		body = m.details.View()
	}
	view := lipgloss.NewStyle().Padding(0, pickPadding).Render(
		lipgloss.JoinVertical(lipgloss.Left, m.headerView(), body, m.footerView()),
	)
	if m.filter != noFilter && m.panel == "" {
		view = m.overlayFilter(view)
	}
	return view
}
