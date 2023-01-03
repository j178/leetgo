package config

import (
	"encoding/json"
	"os"

	"github.com/hashicorp/go-hclog"
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

	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		hclog.L().Error("failed to create state file", "err", err)
		return
	}
	enc := json.NewEncoder(f)
	err = enc.Encode(s)
	if err != nil {
		hclog.L().Error("failed to save state", "err", err)
	}
}
