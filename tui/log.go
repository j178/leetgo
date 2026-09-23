package tui

import (
	"encoding/json"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type logMsg struct {
	Level   string `json:"level"`
	Message string `json:"msg"`
	Err     string `json:"err"`
	Error   string `json:"error"`
}

func (m logMsg) summary() string {
	if m.Message == "add credentials failed, continue requesting without credentials" {
		return "Credentials unavailable · browsing as guest"
	}
	return strings.Join(strings.Fields(m.Message), " ")
}

func (m logMsg) details() string {
	return strings.TrimSpace(m.Message + "\n\n" + m.Err + "\n" + m.Error)
}

type logWriter struct {
	send func(tea.Msg)
}

func (w *logWriter) Write(data []byte) (int, error) {
	var msg logMsg
	if err := json.Unmarshal(data, &msg); err != nil {
		return 0, err
	}
	w.send(msg)
	return len(data), nil
}
