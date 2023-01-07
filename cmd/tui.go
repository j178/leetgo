package cmd

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/j178/leetgo/leetcode"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = lipgloss.NewStyle().PaddingLeft(4).PaddingBottom(1)
	// textStyle         = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type rowDelegate struct{}

func (d rowDelegate) Height() int {
	return 1
}

func (d rowDelegate) Spacing() int {
	return 0
}

func (d rowDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d rowDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(*item)
	if !ok {
		return
	}
	q := (*leetcode.QuestionData)(i)

	str := q.GetTitle()
	if index == m.Index() {
		str = selectedItemStyle.Render("> " + str)
	} else {
		str = itemStyle.Render(str)
	}
	_, _ = fmt.Fprint(w, str)
}

type qsMsg []*leetcode.QuestionData

type item leetcode.QuestionData

func (i *item) FilterValue() string {
	return (*leetcode.QuestionData)(i).GetTitle()
}

type tui struct {
	cache    leetcode.QuestionsCache
	list     *list.Model
	selected *leetcode.QuestionData
}

func newTuiModel(cache leetcode.QuestionsCache) *tui {
	l := list.New(nil, rowDelegate{}, 60, 60)
	l.Title = "Select a question"
	l.SetShowStatusBar(true)
	l.SetShowTitle(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return &tui{
		cache: cache,
		list:  &l,
	}
}

func (m *tui) Selected() *leetcode.QuestionData {
	return m.selected
}

func (m *tui) Init() tea.Cmd {
	return func() tea.Msg {
		qs := m.cache.GetAllQuestions()
		return qsMsg(qs)
	}
}

func (m *tui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.list.SelectedItem() != nil {
				m.selected = (*leetcode.QuestionData)(m.list.SelectedItem().(*item))
				return m, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		if m.list == nil {
			return m, nil
		}
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil
	case qsMsg:
		items := make([]list.Item, len(msg))
		for i, q := range msg {
			items[i] = (*item)(q)
		}
		m.list.SetItems(items)
		return m, nil
	}
	lst, cmd := m.list.Update(msg)
	m.list = &lst
	return m, cmd
}

func (m *tui) View() string {
	return "\n" + m.list.View()
}
