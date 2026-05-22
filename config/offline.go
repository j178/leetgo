package config

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/goccy/go-json"

	"github.com/j178/leetgo/utils"
)

type OfflineQuestion struct {
	FrontendID        string          `json:"frontend_id"`
	Slug              string          `json:"slug"`
	Lang              string          `json:"lang"`
	OutDir            string          `json:"out_dir"`
	SubDir            string          `json:"sub_dir"`
	CodeFile          string          `json:"code_file"`
	TestCasesFile     string          `json:"testcases_file"`
	SystemDesign      bool            `json:"system_design"`
	Content           string          `json:"content,omitempty"`
	TranslatedContent string          `json:"translated_content,omitempty"`
	MetaData          json.RawMessage `json:"meta_data,omitempty"`
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
		q, ok := resolveOfflineQuestion(state, []string{last.Slug, last.FrontendID}, []string{last.Gen, langHint})
		if ok {
			return q, nil
		}
		return OfflineQuestion{}, fmt.Errorf("offline question %q not found", qid)
	}

	q, ok := resolveOfflineQuestion(state, []string{qid}, []string{langHint})
	if ok {
		return q, nil
	}

	return OfflineQuestion{}, fmt.Errorf("offline question %q not found", qid)
}

func resolveOfflineQuestion(state OfflineProjectState, ids []string, preferredLangs []string) (OfflineQuestion, bool) {
	seenIDs := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		if _, ok := seenIDs[id]; ok {
			continue
		}
		seenIDs[id] = struct{}{}

		for _, lang := range preferredLangs {
			if lang == "" {
				continue
			}
			if q, ok := state.Questions[offlineKey(lang, id)]; ok {
				return q, true
			}
		}
	}

	keys := make([]string, 0, len(state.Questions))
	for key := range state.Questions {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		q := state.Questions[key]
		for _, id := range ids {
			if id == "" {
				continue
			}
			if q.Slug == id || q.FrontendID == id {
				return q, true
			}
		}
	}

	return OfflineQuestion{}, false
}
