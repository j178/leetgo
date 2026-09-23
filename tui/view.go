package tui

import (
	"fmt"
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

const pickHelp = `BROWSE QUESTIONS

↑ / ↓ or j / k       Move between questions
PgUp / PgDn          Move one page
Home / End           First / last loaded question
Enter                Pick the selected question
/                    Search by title or question ID
m                    Load more results
r                    Retry a failed request

FILTERS

Tab / →              Next difficulty
Shift+Tab / ←        Previous difficulty
1 / 2 / 3 / 4        All / Easy / Medium / Hard
d                    Cycle difficulty
s                    Cycle completion status
t                    Open the tag picker
c                    Clear all filters and search

TAG PICKER

/                    Find a tag
Space                Toggle the highlighted tag
Enter                Apply selected tags
Esc                  Cancel tag changes

MESSAGES

!                    Show the latest warning
?                    Show this help
Esc / q              Close help or messages
q / Ctrl+C           Quit the picker`

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
	case *tagItem:
		check := "[ ] "
		if entry.checked {
			check = "[✓] "
		}
		row = style.Render(cursor + fitCell(check+entry.Name, m.Width()-2))
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

func (m *model) headerView() string {
	width := m.list.Width()
	var tabs []string
	for i, difficulty := range difficulties {
		label := fmt.Sprintf(" %d %s ", i+1, difficulty.label)
		if width < 50 {
			label = " " + difficulty.label + " "
		}
		if i == m.difficulty {
			label = pickActiveStyle.Render("[" + strings.TrimSpace(label) + "]")
		} else {
			label = pickMutedStyle.Render(label)
		}
		tabs = append(tabs, label)
	}
	title := strings.Join(tabs, " ")
	search := pickMutedStyle.Render("/ Search by title or question ID")
	if m.query != "" {
		search = "/ " + m.query
	}
	if m.search.Focused() {
		search = m.search.View()
	}
	columns := "  " + questionCells("ID", "TITLE", "LEVEL", "STATUS", width-2, pickAccentStyle.Bold(true))
	if m.showTags {
		title = pickMutedStyle.Render("Questions / ") + pickActiveStyle.Render(" Tags ")
		search = pickMutedStyle.Render("/ Find a tag · space to toggle")
		if m.search.Focused() || m.search.Value() != "" {
			search = m.search.View()
		}
		columns = "  SELECT TAGS"
	}
	if m.panel != "" {
		title = pickMutedStyle.Render("Questions / ") + pickActiveStyle.Render(" "+m.panel+" ")
		search = pickMutedStyle.Render("↑↓ scroll · esc to return")
		columns = ""
	}
	return strings.Join([]string{
		splitLine(title, pickAccentStyle.Bold(true).Render("leetgo / pick"), width),
		fitCell(search, width),
		pickMutedStyle.Render(strings.Repeat("─", width)),
		pickAccentStyle.Bold(true).Render(fitCell(columns, width)),
	}, "\n")
}

func (m *model) footerView() string {
	width := m.list.Width()
	tags := "All"
	if len(m.tags) == 1 {
		tags = m.tags[0]
	} else if len(m.tags) > 1 {
		tags = fmt.Sprintf("%d selected", len(m.tags))
	}
	filters := pickAccentStyle.Render("[s]") + " " + statuses[m.status].label + "   " +
		pickAccentStyle.Render("[t]") + " " + tags + "   " + pickMutedStyle.Render("[c] clear")
	count := fmt.Sprintf("%d loaded / %d total", len(m.list.Items()), m.total)
	if len(m.list.Items()) > 0 {
		start, end := m.list.Paginator.GetSliceBounds(len(m.list.Items()))
		count = fmt.Sprintf("%d–%d / %d", start+1, end, m.total)
		if m.hasMore {
			count += "  [m] more"
		}
	}
	if width < 60 {
		filters = pickAccentStyle.Render("[s]") + " " + statuses[m.status].label + "  " +
			pickAccentStyle.Render("[t]") + " " + tags
		count = fmt.Sprintf("%d / %d", len(m.list.Items()), m.total)
	}
	notice := pickMutedStyle.Render("Choose a question to generate a solution")
	if m.notice != nil {
		notice = lipgloss.NewStyle().Foreground(pickYellow).Render("[!] " + m.notice.summary())
	}
	if m.loading {
		count = "Loading questions…"
	}
	if m.err != nil {
		notice = lipgloss.NewStyle().Foreground(pickRed).Render("Could not load questions: " + m.err.Error() + " · r retry")
	}
	hints := []key.Binding{
		pickHint("↑↓", "move"), pickHint("enter", "pick"), pickHint("/", "search"),
		pickHint("?", "help"), pickHint("q", "quit"),
	}
	if m.search.Focused() {
		hints = []key.Binding{pickHint("enter", "search"), pickHint("esc", "cancel")}
	}
	if m.showTags {
		selected := 0
		for _, entry := range m.tagItems {
			if entry.(*tagItem).checked {
				selected++
			}
		}
		filters = fmt.Sprintf("%d selected", selected)
		count = fmt.Sprintf("%d tags", len(m.tagList.VisibleItems()))
		if m.tagsLoading {
			count = "Loading tags…"
		}
		if m.tagsErr != nil {
			notice = lipgloss.NewStyle().Foreground(pickRed).Render("Could not load tags: " + m.tagsErr.Error() + " · r retry")
		}
		hints = []key.Binding{
			pickHint("space", "toggle"), pickHint("enter", "apply"), pickHint("/", "find"), pickHint("esc", "cancel"),
		}
		if m.search.Focused() {
			hints = []key.Binding{pickHint("enter", "find"), pickHint("esc", "cancel search")}
		}
	}
	if m.panel != "" {
		filters, count = m.panel, fmt.Sprintf("%.0f%%", m.details.ScrollPercent()*100)
		hints = []key.Binding{pickHint("↑↓", "scroll"), pickHint("pgup/pgdn", "page"), pickHint("esc", "back")}
	}
	h := help.New()
	h.Width = width
	h.Styles.ShortKey = pickAccentStyle
	h.Styles.ShortDesc = pickMutedStyle
	return strings.Join([]string{
		pickMutedStyle.Render(strings.Repeat("─", width)),
		splitLine(filters, pickMutedStyle.Render(count), width),
		fitCell(h.ShortHelpView(hints), width),
		fitCell(notice, width),
	}, "\n")
}

func pickHint(keyName, description string) key.Binding {
	return key.NewBinding(key.WithKeys(keyName), key.WithHelp(keyName, description))
}

func (m *model) resize() {
	width, height := max(1, m.width-2), max(1, m.height-8)
	m.search.Width = max(1, width-3)
	m.list.SetSize(width, height)
	m.tagList.SetSize(width, height)
	m.details.Width, m.details.Height = width, height
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
	if m.width < 36 || m.height < 12 {
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
	if m.showTags {
		body, empty = m.tagList.View(), ""
		switch {
		case m.tagsLoading:
			empty = "Loading tags…"
		case m.tagsErr != nil:
			empty = "Unable to load tags. Press r to retry."
		case len(m.tagList.VisibleItems()) == 0:
			empty = "No matching tags."
		}
	}
	if empty != "" {
		body = lipgloss.Place(m.list.Width(), m.list.Height(), lipgloss.Center, lipgloss.Center,
			pickMutedStyle.Render(strings.TrimSpace(fitCell(empty, m.list.Width()))))
	}
	if m.panel != "" {
		body = m.details.View()
	}
	return lipgloss.NewStyle().Padding(0, 1).Render(
		lipgloss.JoinVertical(lipgloss.Left, m.headerView(), body, m.footerView()),
	)
}
