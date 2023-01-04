//go:build !cgo

package leetcode

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/utils"
)

type jsonCache struct {
	path     string
	once     sync.Once
	slugs    map[string]*QuestionData
	frontIds map[string]*QuestionData
}

func newCache(path string) QuestionsCache {
	return &jsonCache{path: path}
}

func (c *jsonCache) doLoad() error {
	c.slugs = make(map[string]*QuestionData)
	c.frontIds = make(map[string]*QuestionData)

	var records []QuestionData
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
		c.slugs[r.TitleSlug] = &r
		c.frontIds[r.QuestionFrontendId] = &r
	}
	return nil
}

func (c *jsonCache) load() {
	c.once.Do(
		func() {
			err := c.doLoad()
			if err != nil {
				hclog.L().Warn("failed to load cache, try updating with `leetgo cache update`")
				return
			}
			c.checkUpdateTime()
		},
	)
}

func (c *jsonCache) checkUpdateTime() {
	stat, err := os.Stat(c.path)
	if os.IsNotExist(err) {
		return
	}
	if time.Since(stat.ModTime()) >= 14*24*time.Hour {
		hclog.L().Warn("cache is too old, try updating with `leetgo cache update`")
	}
}

func (c *jsonCache) Update(client Client) error {
	err := utils.CreateIfNotExists(c.path, false)
	if err != nil {
		return err
	}

	all, err := client.GetAllQuestions()
	if err != nil {
		return err
	}
	f, err := os.Create(c.path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	enc := json.NewEncoder(f)
	err = enc.Encode(all)
	if err != nil {
		return err
	}
	hclog.L().Info("cache updated", "path", c.path)
	return nil
}

func (c *jsonCache) GetBySlug(slug string) *QuestionData {
	c.load()
	return c.slugs[slug]
}

func (c *jsonCache) GetById(id string) *QuestionData {
	c.load()
	return c.frontIds[id]
}
