package tui

import (
	"image"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) updateMouse(msg tea.MouseMsg) tea.Cmd {
	if msg.Action != tea.MouseActionPress || m.width < pickMinWidth || m.height < pickMinHeight {
		return nil
	}
	if m.filter != noFilter {
		return m.updateFilterMouse(msg)
	}
	layout := m.layout()
	position := image.Pt(msg.X, msg.Y)
	if msg.Button == tea.MouseButtonLeft && m.panel == "" {
		for _, control := range m.filterControls() {
			if position.In(control.bounds) {
				if control.kind == noFilter {
					return m.clearFilters()
				}
				return m.openFilter(control.kind)
			}
		}
		if position.In(layout.search) {
			m.search.SetValue(m.query)
			m.search.CursorEnd()
			return m.search.Focus()
		}
		if msg.Y == layout.body.Min.Y-2 {
			m.search.Blur()
			if msg.X >= layout.questions.Min.X && msg.X < layout.questions.Max.X {
				m.previewFocused = false
			} else if m.previewVisible() && msg.X >= layout.preview.Min.X && msg.X < layout.preview.Max.X {
				m.previewFocused = true
			}
			return nil
		}
	}
	if !position.In(layout.body) {
		return nil
	}
	if m.panel != "" {
		var cmd tea.Cmd
		m.details, cmd = m.details.Update(msg)
		return cmd
	}
	if m.previewVisible() && position.In(layout.preview) {
		if msg.Button == tea.MouseButtonLeft {
			m.previewFocused = true
			m.search.Blur()
		}
		var cmd tea.Cmd
		m.preview.viewport, cmd = m.preview.viewport.Update(msg)
		return cmd
	}

	if !updateListMouse(&m.list, msg, layout.questions) {
		return nil
	}
	if msg.Button == tea.MouseButtonLeft {
		m.previewFocused = false
		m.search.Blur()
	}
	return m.syncPreview()
}

func (m *model) updateFilterMouse(msg tea.MouseMsg) tea.Cmd {
	layout := m.dropdownLayout()
	position := image.Pt(msg.X, msg.Y)
	if !position.In(layout.bounds) {
		if msg.Button == tea.MouseButtonLeft {
			for _, control := range m.filterControls() {
				if position.In(control.bounds) {
					if control.kind == noFilter {
						return m.clearFilters()
					}
					return m.openFilter(control.kind)
				}
			}
			m.closeFilter()
			return m.syncPreview()
		}
		return nil
	}
	if msg.Button == tea.MouseButtonLeft {
		switch {
		case position.In(layout.search):
			m.filterSearch.CursorEnd()
			return m.filterSearch.Focus()
		case position.In(layout.apply):
			return m.applyFilter()
		case position.In(layout.cancel):
			m.closeFilter()
			return m.syncPreview()
		}
	}
	if m.filter == tagsFilter && (m.tagsLoading || m.tagsErr != nil) {
		return nil
	}
	if updateListMouse(&m.filterList, msg, layout.options) && msg.Button == tea.MouseButtonLeft {
		m.filterSearch.Blur()
		if m.filter != tagsFilter {
			return m.applyFilter()
		}
		choice := m.filterList.SelectedItem().(*filterItem)
		choice.checked = !choice.checked
	}
	return nil
}

func updateListMouse(l *list.Model, msg tea.MouseMsg, bounds image.Rectangle) bool {
	if !image.Pt(msg.X, msg.Y).In(bounds) || len(l.VisibleItems()) == 0 {
		return false
	}
	switch msg.Button {
	case tea.MouseButtonLeft:
		start, end := l.Paginator.GetSliceBounds(len(l.VisibleItems()))
		index := start + msg.Y - bounds.Min.Y
		if index >= end {
			return false
		}
		l.Select(index)
	case tea.MouseButtonWheelUp:
		for range 3 {
			l.CursorUp()
		}
	case tea.MouseButtonWheelDown:
		for range 3 {
			l.CursorDown()
		}
	default:
		return false
	}
	return true
}
