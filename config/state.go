package config

import (
	"os"

	"github.com/goccy/go-json"
	"github.com/hashicorp/go-hclog"

	"github.com/j178/leetgo/utils"
)

// Small project state management.

type LastQuestion struct {
	FrontendID string `json:"frontend_id"`
	Slug       string `json:"slug"`
	Gen        string `json:"gen"`
}

type State struct {
	LastQuestion LastQuestion `json:"last_question"`
	LastContest  string       `json:"last_contest"`
}

type States map[string]State

func loadStates() States {
	s := make(States)

	file := Get().StateFile()
	f, err := os.Open(file)
	if err != nil {
		hclog.L().Debug("failed to open state file", "err", err)
		return s
	}
	defer func() { _ = f.Close() }()

	dec := json.NewDecoder(f)
	err = dec.Decode(&s)
	if err != nil {
		hclog.L().Debug("failed to load state", "err", err)
	}

	return s
}

func LoadState() State {
	s := loadStates()
	projectRoot := Get().ProjectRoot()
	return s[projectRoot]
}

func SaveState(s State) {
	projectRoot := Get().ProjectRoot()
	file := Get().StateFile()
	states := loadStates()
	states[projectRoot] = s

	err := utils.CreateIfNotExists(file, false)
	if err != nil {
		hclog.L().Error("failed to create state file", "err", err)
		return
	}
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		hclog.L().Error("failed to open state file", "err", err)
		return
	}
	enc := json.NewEncoder(f)
	err = enc.Encode(states)
	if err != nil {
		hclog.L().Error("failed to save state", "err", err)
	}
}
