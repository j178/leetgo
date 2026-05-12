package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/goccy/go-json"

	"github.com/j178/leetgo/utils"
)

type OfflineQuestion struct {
	FrontendID    string `json:"frontend_id"`
	Slug          string `json:"slug"`
	Lang          string `json:"lang"`
	OutDir        string `json:"out_dir"`
	SubDir        string `json:"sub_dir"`
	CodeFile      string `json:"code_file"`
	TestCasesFile string `json:"testcases_file"`
	SystemDesign  bool   `json:"system_design"`
}

type OfflineProjectState struct {
	Questions map[string]OfflineQuestion `json:"questions"`
}

type OfflineStates map[string]OfflineProjectState

func offlineKey(lang, qid string) string {
	return strings.ToLower(lang) + ":" + qid
}

func loadOfflineStates() OfflineStates {
	s := make(OfflineStates)

	file := Get().OfflineStateFile()
	f, err := os.Open(file)
	if err != nil {
		log.Debug("failed to open offline state file", "err", err)
		return s
	}
	defer func() { _ = f.Close() }()

	dec := json.NewDecoder(f)
	err = dec.Decode(&s)
	if err != nil {
		log.Debug("failed to load offline state", "err", err)
	}

	return s
}

func saveOfflineStates(states OfflineStates) error {
	file := Get().OfflineStateFile()
	err := utils.CreateIfNotExists(file, false)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(file, os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	return json.NewEncoder(f).Encode(states)
}

func SaveOfflineQuestion(q OfflineQuestion) error {
	if q.Lang == "" || q.Slug == "" {
		return fmt.Errorf("offline question requires lang and slug")
	}

	projectRoot := Get().ProjectRoot()
	states := loadOfflineStates()
	state := states[projectRoot]
	if state.Questions == nil {
		state.Questions = make(map[string]OfflineQuestion)
	}

	lang := strings.ToLower(q.Lang)
	state.Questions[offlineKey(lang, q.Slug)] = q
	if q.FrontendID != "" {
		state.Questions[offlineKey(lang, q.FrontendID)] = q
	}
	states[projectRoot] = state

	err := saveOfflineStates(states)
	if err != nil {
		log.Error("failed to save offline state", "err", err)
	}
	return err
}

func ResolveOfflineQuestion(qid string, langHint string) (OfflineQuestion, error) {
	projectRoot := Get().ProjectRoot()
	states := loadOfflineStates()
	state := states[projectRoot]
	if len(state.Questions) == 0 {
		return OfflineQuestion{}, fmt.Errorf("no offline questions found, run `leetgo pick` first")
	}

	if qid == "last" {
		last := LoadState().LastQuestion
		if last.Slug == "" {
			return OfflineQuestion{}, fmt.Errorf("offline question %q not found", qid)
		}
		if last.Gen != "" {
			if q, ok := state.Questions[offlineKey(last.Gen, last.Slug)]; ok {
				return q, nil
			}
			if last.FrontendID != "" {
				if q, ok := state.Questions[offlineKey(last.Gen, last.FrontendID)]; ok {
					return q, nil
				}
			}
		}
		if langHint != "" {
			if q, ok := state.Questions[offlineKey(langHint, last.Slug)]; ok {
				return q, nil
			}
			if last.FrontendID != "" {
				if q, ok := state.Questions[offlineKey(langHint, last.FrontendID)]; ok {
					return q, nil
				}
			}
		}
		return OfflineQuestion{}, fmt.Errorf("offline question %q not found", qid)
	}

	if langHint != "" {
		if q, ok := state.Questions[offlineKey(langHint, qid)]; ok {
			return q, nil
		}
	}
	for _, q := range state.Questions {
		if q.Slug == qid || q.FrontendID == qid {
			return q, nil
		}
	}

	return OfflineQuestion{}, fmt.Errorf("offline question %q not found", qid)
}
