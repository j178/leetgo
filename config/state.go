package config

import (
	"encoding/json"
	"os"

	"github.com/hashicorp/go-hclog"
)

// Small project state management.

type LastGeneratedQuestion struct {
	FrontendID string `json:"frontend_id"`
	Slug       string `json:"slug"`
	Gen        string `json:"gen"`
}

type State struct {
	LastGenerated LastGeneratedQuestion `json:"last_generated"`
}

func LoadState() State {
	var s State

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

func SaveState(s State) {
	file := Get().StateFile()
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
