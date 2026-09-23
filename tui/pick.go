package tui

import (
	"strings"
	"time"

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
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

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

type searchMsg struct{ request int }

type tagsMsg struct {
	tags []leetcode.QuestionTag
	err  error
}

type item leetcode.QuestionData

func (i *item) FilterValue() string { return (*leetcode.QuestionData)(i).GetTitle() }

type model struct {
	client        leetcode.Client
	difficulty    int
	status        int
	query         string
	tags          []string
	list          list.Model
	search        textinput.Model
	selected      *leetcode.QuestionData
	width         int
	height        int
	request       int
	loading       bool
	err           error
	total         int
	hasMore       bool
	pendingSearch tea.Cmd

	filter        filterKind
	filterItems   []list.Item
	filterList    list.Model
	filterSearch  textinput.Model
	availableTags []leetcode.QuestionTag
	tagsLoaded    bool
	tagsLoading   bool
	tagsErr       error

	notice         *logMsg
	panel          string
	details        viewport.Model
	preview        questionPreview
	previewFocused bool
}

func newModel(c leetcode.Client) *model {
	l := list.New(nil, rowDelegate{}, 80, 20)
	filters := list.New(nil, rowDelegate{}, 80, 20)
	for _, l := range []*list.Model{&l, &filters} {
		l.SetShowTitle(false)
		l.SetShowStatusBar(false)
		l.SetShowHelp(false)
		l.SetShowPagination(false)
		l.SetFilteringEnabled(false)
		l.KeyMap.NextPage.SetKeys("pgdown", "f", "ctrl+f")
		l.KeyMap.PrevPage.SetKeys("pgup", "b", "ctrl+b")
	}
	filters.DisableQuitKeybindings()

	search := textinput.New()
	search.Prompt = "/ "
	search.Placeholder = "Title or question ID"
	filterSearch := textinput.New()
	filterSearch.Prompt = "/ "
	filterSearch.Placeholder = "Find tags"
	m := &model{
		client: c, list: l, filterList: filters, search: search, filterSearch: filterSearch,
		width: 80, height: 24, details: viewport.New(80, 16),
		preview: questionPreview{
			viewport: viewport.New(1, 1),
			cache:    make(map[string]*leetcode.QuestionData),
			pending:  make(map[string]bool),
		},
	}
	m.resize()
	return m
}

func (m *model) Init() tea.Cmd { return m.loadQuestions(true) }

func (m *model) loadQuestions(reset bool) tea.Cmd {
	m.pendingSearch = nil
	if reset {
		m.list.SetItems(nil)
		m.list.ResetSelected()
		m.total = 0
		m.hasMore = false
		m.syncPreview()
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

func (m *model) searchQuestions() tea.Cmd {
	query := strings.TrimSpace(m.search.Value())
	if query == m.query {
		return nil
	}
	m.query = query
	// Invalidate old responses immediately, before the debounce timer fires.
	m.pendingSearch = m.loadQuestions(true)
	request := m.request
	return tea.Tick(150*time.Millisecond, func(time.Time) tea.Msg {
		return searchMsg{request: request}
	})
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
		return m, m.syncPreview()
	case tea.MouseMsg:
		return m, m.updateMouse(msg)
	case searchMsg:
		if msg.request != m.request {
			return m, nil
		}
		cmd := m.pendingSearch
		m.pendingSearch = nil
		return m, cmd
	case previewLoadMsg:
		if msg.slug != m.preview.slug || !m.preview.loading || !m.previewVisible() {
			return m, nil
		}
		return m, m.loadPreview()
	case previewMsg:
		delete(m.preview.pending, msg.slug)
		if msg.err == nil {
			m.preview.cache[msg.slug] = msg.question
		}
		if msg.slug == m.preview.slug {
			m.preview.loading, m.preview.err = false, msg.err
			m.renderPreview()
		}
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
		cmd := m.list.SetItems(items)
		return m, tea.Batch(cmd, m.syncPreview())
	case tagsMsg:
		m.tagsLoading = false
		m.tagsErr = msg.err
		if msg.err != nil {
			return m, nil
		}
		m.tagsLoaded = true
		m.availableTags = msg.tags
		if m.filter == tagsFilter {
			m.populateFilter()
		}
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
				return m, m.syncPreview()
			}
		}
		var cmd tea.Cmd
		m.details, cmd = m.details.Update(msg)
		return m, cmd
	}
	if m.filter != noFilter {
		return m, m.updateFilter(msg)
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
				m.previewFocused = false
				return m, nil
			case "esc":
				m.search.Blur()
				m.previewFocused = false
				return m, nil
			}
		}
		var cmd tea.Cmd
		query := m.search.Value()
		m.search, cmd = m.search.Update(msg)
		if m.search.Value() != query {
			return m, tea.Batch(cmd, m.searchQuestions())
		}
		return m, cmd
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "tab", "shift+tab":
			m.previewFocused = m.previewVisible() && !m.previewFocused
			return m, nil
		case "left", "h":
			m.previewFocused = false
			return m, nil
		case "right", "l":
			m.previewFocused = m.previewVisible()
			return m, nil
		case "d":
			return m, m.openFilter(difficultyFilter)
		case "s":
			return m, m.openFilter(statusFilter)
		case "t":
			return m, m.openFilter(tagsFilter)
		case "/":
			m.search.Placeholder = "Title or question ID"
			m.search.SetValue(m.query)
			m.search.CursorEnd()
			return m, m.search.Focus()
		case "c":
			return m, m.clearFilters()
		case "r", "m":
			if keyMsg.String() == "r" && m.err == nil && m.previewVisible() && m.preview.err != nil {
				return m, m.loadPreview()
			}
			if !m.loading && (m.err != nil || m.hasMore) {
				return m, m.loadQuestions(false)
			}
			return m, nil
		case "q", "esc":
			return m, tea.Quit
		case "enter":
			if selected, ok := m.list.SelectedItem().(*item); ok {
				m.selected = (*leetcode.QuestionData)(selected)
				if q := m.preview.cache[m.selected.TitleSlug]; q != nil {
					m.selected = q
				}
				return m, tea.Quit
			}
		case "J", "K", "ctrl+d", "ctrl+u":
			if m.previewVisible() {
				switch keyMsg.String() {
				case "J":
					m.preview.viewport.ScrollDown(1)
				case "K":
					m.preview.viewport.ScrollUp(1)
				case "ctrl+d":
					m.preview.viewport.HalfPageDown()
				case "ctrl+u":
					m.preview.viewport.HalfPageUp()
				}
			}
			return m, nil
		}
	}
	if m.previewFocused && m.previewVisible() {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "home", "g":
				m.preview.viewport.GotoTop()
				return m, nil
			case "end", "G":
				m.preview.viewport.GotoBottom()
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.preview.viewport, cmd = m.preview.viewport.Update(msg)
		return m, cmd
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, tea.Batch(cmd, m.syncPreview())
}
