package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"

	"github.com/j178/leetgo/leetcode"
)

type questionPreview struct {
	viewport viewport.Model
	slug     string
	request  int
	loading  bool
	err      error
	cache    map[string]*leetcode.QuestionData
	pending  map[string]bool
}

type previewLoadMsg struct {
	slug    string
	request int
}

type previewMsg struct {
	slug     string
	question *leetcode.QuestionData
	err      error
}

func (m *model) previewVisible() bool {
	return m.width >= pickPreviewMinWidth && m.height >= pickMinHeight && m.panel == ""
}

func (m *model) syncPreview() tea.Cmd {
	m.preview.request++
	var slug string
	if selected, ok := m.list.SelectedItem().(*item); ok {
		slug = selected.TitleSlug
	}
	if slug != m.preview.slug {
		m.preview.slug = slug
		m.preview.err = nil
		m.preview.loading = slug != "" && m.preview.cache[slug] == nil
		m.preview.viewport.GotoTop()
		m.renderPreview()
	}
	if !m.preview.loading || m.preview.pending[slug] || !m.previewVisible() {
		return nil
	}
	// Invalidate older timers even when navigation returns to the same question.
	request := m.preview.request
	return tea.Tick(400*time.Millisecond, func(time.Time) tea.Msg {
		return previewLoadMsg{slug: slug, request: request}
	})
}

func (m *model) loadPreview() tea.Cmd {
	slug := m.preview.slug
	if slug == "" || m.preview.pending[slug] || m.preview.cache[slug] != nil {
		return nil
	}
	m.preview.loading, m.preview.err = true, nil
	m.preview.pending[slug] = true
	m.renderPreview()
	client := m.client
	return func() tea.Msg {
		q, err := client.GetQuestionData(slug)
		return previewMsg{slug: slug, question: q, err: err}
	}
}

func (m *model) renderPreview() {
	if m.width < pickPreviewMinWidth {
		return
	}
	width := m.preview.viewport.Width
	content := "Select a question to preview its description."
	switch {
	case m.preview.loading:
		content = "Loading description…"
	case m.preview.err != nil:
		content = "Could not load description: " + m.preview.err.Error() + "\n\nPress r to retry."
	case m.preview.cache[m.preview.slug] != nil:
		q := m.preview.cache[m.preview.slug]
		description := strings.TrimSpace(q.GetFormattedContent())
		if description == "" {
			description = "No description available."
		}
		content = fmt.Sprintf("# %s. %s\n\n%s", q.QuestionFrontendId, q.GetTitle(), description)
		style := "light"
		if lipgloss.HasDarkBackground() {
			style = "dark"
		}
		renderer, err := glamour.NewTermRenderer(
			glamour.WithStandardStyle(style),
			glamour.WithColorProfile(lipgloss.ColorProfile()),
			glamour.WithWordWrap(width),
		)
		if err == nil {
			if rendered, err := renderer.Render(content); err == nil {
				// Code blocks and long tokens can exceed Glamour's word wrap.
				m.preview.viewport.SetContent(wrap.String(strings.Trim(rendered, "\n"), width))
				return
			}
		}
	}
	m.preview.viewport.SetContent(wrap.String(wordwrap.String(content, width), width))
}

func (m *model) previewHeader() string {
	var status string
	switch {
	case m.preview.loading:
		status = "Loading…"
	case m.preview.err != nil:
		status = "r retry"
	case m.preview.slug != "":
		status = fmt.Sprintf("%.0f%%", m.preview.viewport.ScrollPercent()*100)
	}
	return splitLine(paneTitle("PREVIEW", m.previewFocused), status, m.preview.viewport.Width)
}

func (m *model) joinPreview(left, right string) string {
	divider := strings.TrimSuffix(strings.Repeat(pickDivider+"\n", lipgloss.Height(left)), "\n")
	return lipgloss.JoinHorizontal(lipgloss.Top, left, pickMutedStyle.Render(divider), right)
}
