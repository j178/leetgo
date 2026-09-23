package tui

import (
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"

	"github.com/j178/leetgo/leetcode"
)

// Pick opens the interactive question picker and returns the selected question.
// Cancelling returns a nil question and no error.
func Pick(c leetcode.Client) (*leetcode.QuestionData, error) {
	m := newModel(c)
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Authentication and HTTP retries log from command goroutines. Route all of
	// their output through Update so only Bubble Tea writes to the terminal.
	previous := log.Default()
	logger := previous.With()
	logger.SetOutput(&logWriter{send: p.Send})
	logger.SetFormatter(log.JSONFormatter)
	log.SetDefault(logger)
	defer log.SetDefault(previous)

	if _, err := p.Run(); err != nil {
		return nil, err
	}
	return m.selected, nil
}

type filterOption struct {
	label string
	value string
}

var (
	difficulties = []filterOption{{"All", ""}, {"Easy", "EASY"}, {"Medium", "MEDIUM"}, {"Hard", "HARD"}}
	statuses     = []filterOption{{"All", ""}, {"Not Started", "NOT_STARTED"}, {"Tried", "TRIED"}, {"Accepted", "AC"}}
)

type questionsMsg struct {
	request int
	result  leetcode.QuestionList
	err     error
}

type tagsMsg struct {
	tags []leetcode.QuestionTag
	err  error
}

type item leetcode.QuestionData

func (i *item) FilterValue() string { return (*leetcode.QuestionData)(i).GetTitle() }

type tagItem struct {
	leetcode.QuestionTag
	checked bool
}

func (i *tagItem) FilterValue() string { return i.Name + " " + i.NameTranslated + " " + i.Slug }

type model struct {
	client     leetcode.Client
	difficulty int
	status     int
	query      string
	tags       []string
	list       list.Model
	search     textinput.Model
	selected   *leetcode.QuestionData
	width      int
	height     int
	request    int
	loading    bool
	err        error
	total      int
	hasMore    bool

	showTags    bool
	tagItems    []list.Item
	tagList     list.Model
	tagsLoaded  bool
	tagsLoading bool
	tagsErr     error

	notice  *logMsg
	panel   string
	details viewport.Model
}

func newModel(c leetcode.Client) *model {
	l := list.New(nil, rowDelegate{}, 80, 20)
	tags := list.New(nil, rowDelegate{}, 80, 20)
	for _, l := range []*list.Model{&l, &tags} {
		l.SetShowTitle(false)
		l.SetShowStatusBar(false)
		l.SetShowHelp(false)
		l.SetShowPagination(false)
		l.SetFilteringEnabled(false)
	}
	l.KeyMap.NextPage.SetKeys("pgdown", "f", "ctrl+f")
	l.KeyMap.PrevPage.SetKeys("pgup", "b", "ctrl+b")
	tags.DisableQuitKeybindings()

	search := textinput.New()
	search.Prompt = "/ "
	search.Placeholder = "Title or question ID"
	m := &model{
		client: c, list: l, tagList: tags, search: search,
		width: 80, height: 24, details: viewport.New(80, 16),
	}
	m.resize()
	return m
}

func (m *model) Init() tea.Cmd { return m.loadQuestions(true) }

func (m *model) loadQuestions(reset bool) tea.Cmd {
	if reset {
		m.list.SetItems(nil)
		m.list.ResetSelected()
		m.total = 0
		m.hasMore = false
	}
	m.request++
	m.loading = true
	m.err = nil
	request, skip, client := m.request, len(m.list.Items()), m.client
	filter := leetcode.QuestionFilter{
		Difficulty: difficulties[m.difficulty].value, Status: statuses[m.status].value,
		SearchKeywords: m.query, Tags: m.tags,
	}
	return func() tea.Msg {
		result, err := client.GetQuestionsByFilter(filter, 100, skip)
		return questionsMsg{request: request, result: result, err: err}
	}
}

func (m *model) loadTags() tea.Cmd {
	m.tagsLoading = true
	m.tagsErr = nil
	client := m.client
	return func() tea.Msg {
		tags, err := client.GetQuestionTags()
		return tagsMsg{tags: tags, err: err}
	}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case logMsg:
		if msg.Level == "warn" || msg.Level == "error" {
			m.notice = &msg
			if m.panel == "Messages" {
				m.refreshPanel()
			}
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.resize()
		return m, nil
	case questionsMsg:
		if msg.request != m.request {
			return m, nil
		}
		m.loading = false
		m.err = msg.err
		if msg.err != nil {
			return m, nil
		}
		items := m.list.Items()
		for _, q := range msg.result.Questions {
			items = append(items, (*item)(q))
		}
		m.total = msg.result.Total
		// The US endpoint provides total, but not hasMore.
		m.hasMore = len(msg.result.Questions) > 0 && (msg.result.HasMore || len(items) < m.total)
		return m, m.list.SetItems(items)
	case tagsMsg:
		m.tagsLoading = false
		m.tagsErr = msg.err
		if msg.err != nil {
			return m, nil
		}
		m.tagsLoaded = true
		m.tagItems = make([]list.Item, len(msg.tags))
		for i, tag := range msg.tags {
			m.tagItems[i] = &tagItem{QuestionTag: tag}
		}
		m.filterTags()
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	if m.panel != "" {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "esc", "q", "?", "!":
				m.panel = ""
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.details, cmd = m.details.Update(msg)
		return m, cmd
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok && !m.search.Focused() {
		switch keyMsg.String() {
		case "?":
			m.openPanel("Help")
			return m, nil
		case "!":
			if m.notice != nil {
				m.openPanel("Messages")
			}
			return m, nil
		}
	}
	if m.search.Focused() {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "enter":
				m.search.Blur()
				if m.showTags {
					return m, nil
				}
				m.query = strings.TrimSpace(m.search.Value())
				return m, m.loadQuestions(true)
			case "esc":
				m.search.Blur()
				if m.showTags {
					m.search.Reset()
					m.filterTags()
				}
				return m, nil
			}
		}
		var cmd tea.Cmd
		query := m.search.Value()
		m.search, cmd = m.search.Update(msg)
		if m.showTags && m.search.Value() != query {
			m.filterTags()
		}
		return m, cmd
	}
	if m.showTags {
		return m.updateTags(msg)
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "d", "tab", "right", "l", "shift+tab", "left", "h", "1", "2", "3", "4":
			switch keyMsg.String() {
			case "shift+tab", "left", "h":
				m.difficulty = (m.difficulty + len(difficulties) - 1) % len(difficulties)
			case "1", "2", "3", "4":
				m.difficulty = int(keyMsg.String()[0] - '1')
			default:
				m.difficulty = (m.difficulty + 1) % len(difficulties)
			}
			return m, m.loadQuestions(true)
		case "s":
			m.status = (m.status + 1) % len(statuses)
			return m, m.loadQuestions(true)
		case "t":
			m.showTags = true
			m.search.Reset()
			m.search.Placeholder = "Tag name"
			for _, entry := range m.tagItems {
				tag := entry.(*tagItem)
				tag.checked = slices.Contains(m.tags, tag.Slug)
			}
			m.filterTags()
			if !m.tagsLoaded && !m.tagsLoading {
				return m, m.loadTags()
			}
			return m, nil
		case "/":
			m.search.Placeholder = "Title or question ID"
			m.search.SetValue(m.query)
			m.search.CursorEnd()
			return m, m.search.Focus()
		case "c":
			m.difficulty, m.status = 0, 0
			m.query, m.tags = "", nil
			return m, m.loadQuestions(true)
		case "r", "m":
			if !m.loading && (m.err != nil || m.hasMore) {
				return m, m.loadQuestions(false)
			}
			return m, nil
		case "enter":
			if selected, ok := m.list.SelectedItem().(*item); ok {
				m.selected = (*leetcode.QuestionData)(selected)
				return m, tea.Quit
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *model) updateTags(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc", "q":
			m.showTags = false
			return m, nil
		case "r":
			if m.tagsErr != nil && !m.tagsLoading {
				return m, m.loadTags()
			}
		case "/":
			return m, m.search.Focus()
		case " ":
			if tag, ok := m.tagList.SelectedItem().(*tagItem); ok {
				tag.checked = !tag.checked
			}
			return m, nil
		case "enter":
			if m.tagsLoading || m.tagsErr != nil {
				return m, nil
			}
			var tags []string
			for _, entry := range m.tagItems {
				if tag := entry.(*tagItem); tag.checked {
					tags = append(tags, tag.Slug)
				}
			}
			if len(tags) == len(m.tagItems) {
				tags = nil
			}
			m.tags = tags
			m.showTags = false
			return m, m.loadQuestions(true)
		}
	}
	var cmd tea.Cmd
	m.tagList, cmd = m.tagList.Update(msg)
	return m, cmd
}

// Tags are a small local list; filtering here avoids asynchronous results
// replacing a newer search or a reopened selection draft.
func (m *model) filterTags() {
	items := m.tagItems
	if query := m.search.Value(); query != "" && m.showTags {
		targets := make([]string, len(m.tagItems))
		for i, entry := range m.tagItems {
			targets[i] = entry.FilterValue()
		}
		items = nil
		for _, rank := range list.DefaultFilter(query, targets) {
			items = append(items, m.tagItems[rank.Index])
		}
	}
	m.tagList.SetItems(items)
	m.tagList.ResetSelected()
}
