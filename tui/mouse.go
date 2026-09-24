package tui

import (
	"image"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

func (m *model) updateMouse(msg tea.MouseMsg) tea.Cmd {
	switch msg.(type) {
	case tea.MouseClickMsg, tea.MouseWheelMsg:
	default:
		return nil
	}
	if m.width < pickMinWidth || m.height < pickMinHeight {
		return nil
	}
	if m.filter != noFilter {
		return m.updateFilterMouse(msg)
	}
	mouse := msg.Mouse()
	layout := m.layout()
	position := image.Pt(mouse.X, mouse.Y)
	if mouse.Button == tea.MouseLeft && m.panel == "" {
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
		if mouse.Y == layout.body.Min.Y-2 {
			m.search.Blur()
			if mouse.X >= layout.questions.Min.X && mouse.X < layout.questions.Max.X {
				m.previewFocused = false
			} else if m.previewVisible() && mouse.X >= layout.preview.Min.X && mouse.X < layout.preview.Max.X {
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
		if mouse.Button == tea.MouseLeft {
			m.previewFocused = true
			m.search.Blur()
		}
		var cmd tea.Cmd
		m.preview.viewport, cmd = m.preview.viewport.Update(msg)
		return cmd
	}

	if !updateListMouse(&m.list, mouse, layout.questions) {
		return nil
	}
	if mouse.Button == tea.MouseLeft {
		m.previewFocused = false
		m.search.Blur()
	}
	return m.syncPreview()
}

func (m *model) updateFilterMouse(msg tea.MouseMsg) tea.Cmd {
	mouse := msg.Mouse()
	layout := m.dropdownLayout()
	position := image.Pt(mouse.X, mouse.Y)
	if !position.In(layout.bounds) {
		if mouse.Button == tea.MouseLeft {
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
	if mouse.Button == tea.MouseLeft {
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
	if updateListMouse(&m.filterList, mouse, layout.options) && mouse.Button == tea.MouseLeft {
		m.filterSearch.Blur()
		if m.filter != tagsFilter {
			return m.applyFilter()
		}
		choice := m.filterList.SelectedItem().(*filterItem)
		choice.checked = !choice.checked
	}
	return nil
}

func updateListMouse(l *list.Model, mouse tea.Mouse, bounds image.Rectangle) bool {
	if !image.Pt(mouse.X, mouse.Y).In(bounds) || len(l.VisibleItems()) == 0 {
		return false
	}
	switch mouse.Button {
	case tea.MouseLeft:
		start, end := l.Paginator.GetSliceBounds(len(l.VisibleItems()))
		index := start + mouse.Y - bounds.Min.Y
		if index >= end {
			return false
		}
		l.Select(index)
	case tea.MouseWheelUp:
		l.CursorUp()
	case tea.MouseWheelDown:
		l.CursorDown()
	default:
		return false
	}
	return true
}
