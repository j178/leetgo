package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/j178/leetgo/leetcode"
)

type model struct {
	cache    leetcode.QuestionsCache
	choices  []string
	cursor   int
	selected map[int]struct{}
}

func initialModel(cache leetcode.QuestionsCache) model {
	return model{
		cache:    cache,
		choices:  []string{"加班", "回家", "辞职", "上班"},
		cursor:   0,
		selected: map[int]struct{}{},
	}
}

func (m model) Selected() *leetcode.QuestionData {
	return nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "j", "down":
			m.cursor++
			m.cursor = m.cursor % len(m.choices)
		case "k", "up":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.choices) - 1
			}
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	s := "What do you want to do?\n\n"
	for i, choice := range m.choices {
		if i == m.cursor {
			s += "▸ "
		} else {
			s += "  "
		}
		_, ok := m.selected[i]
		if ok {
			s += "[x] "
		} else {
			s += "[ ] "
		}
		s += choice + "\n"
	}
	s += "press q to quit"
	return s
}
