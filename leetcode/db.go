package leetcode

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type QuestionRecord struct {
	FrontendId string   `json:"frontendId,omitempty"`
	Slug       string   `json:"slug,omitempty"`
	Title      string   `json:"title,omitempty"`
	CnTitle    string   `json:"cnTitle,omitempty"`
	Difficulty string   `json:"difficulty,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	PaidOnly   bool     `json:"paidOnly,omitempty"`
}

type QuestionsDB interface {
	GetBySlug(slug string) *QuestionRecord
	GetById(id string) *QuestionRecord
	Update() error
}

type cache struct {
	path     string
	client   *Client
	slugs    map[string]*QuestionRecord
	frontIds map[string]*QuestionRecord
}

func (c *cache) load() error {
	c.slugs = make(map[string]*QuestionRecord)
	c.frontIds = make(map[string]*QuestionRecord)

	var records []QuestionRecord
	if _, err := os.Stat(c.path); errors.Is(err, os.ErrNotExist) {
		return err
	}
	s, err := os.ReadFile(c.path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(s, &records)
	if err != nil {
		return err
	}
	for _, r := range records {
		r := r
		c.slugs[r.Slug] = &r
		c.frontIds[r.FrontendId] = &r
	}
	return nil
}

func (c *cache) Update() error {
	dir := filepath.Dir(c.path)
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	all, err := c.client.GetAllQuestions()
	if err != nil {
		return err
	}
	f, err := os.Create(c.path)
	if err != nil {
		return err
	}
	questions := make([]QuestionRecord, 0, len(all))
	for _, q := range all {
		tags := make([]string, 0, len(q["topicTags"].([]any)))
		for _, t := range q["topicTags"].([]any) {
			tags = append(tags, t.(map[string]any)["slug"].(string))
		}
		cnTitle := ""
		if q["translatedTitle"] != nil {
			cnTitle = q["translatedTitle"].(string)
		}
		questions = append(
			questions, QuestionRecord{
				FrontendId: q["questionFrontendId"].(string),
				Slug:       q["titleSlug"].(string),
				Title:      q["title"].(string),
				CnTitle:    cnTitle,
				Difficulty: q["difficulty"].(string),
				Tags:       tags,
				PaidOnly:   q["isPaidOnly"].(bool),
			},
		)
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")
	err = enc.Encode(questions)
	if err != nil {
		return err
	}
	return c.load()
}

func (c *cache) GetBySlug(slug string) *QuestionRecord {
	return c.slugs[slug]
}

func (c *cache) GetById(id string) *QuestionRecord {
	return c.frontIds[id]
}

func NewDB(path string, client *Client) QuestionsDB {
	c := &cache{path: path, client: client}
	_ = c.load()
	return c
}
