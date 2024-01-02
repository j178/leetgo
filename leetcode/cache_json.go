//go:build !sqlite

package leetcode

import (
	"errors"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/goccy/go-json"

	"github.com/j178/leetgo/utils"
)

var cacheExt = ".json"

type jsonCache struct {
	path     string
	client   Client
	once     sync.Once
	slugs    map[string]*QuestionData
	frontIds map[string]*QuestionData
}

func newCache(path string, c Client) QuestionsCache {
	return &jsonCache{path: path, client: c}
}

func (c *jsonCache) CacheFile() string {
	return c.path
}

func (c *jsonCache) doLoad() error {
	c.slugs = make(map[string]*QuestionData)
	c.frontIds = make(map[string]*QuestionData)

	if _, err := os.Stat(c.path); errors.Is(err, os.ErrNotExist) {
		return err
	}
	s, err := os.ReadFile(c.path)
	if err != nil {
		return err
	}

	var records []*QuestionData
	err = json.Unmarshal(s, &records)
	if err != nil {
		return err
	}
	for _, r := range records {
		r.partial = 1
		r.client = c.client
		c.slugs[r.TitleSlug] = r
		c.frontIds[r.QuestionFrontendId] = r
	}
	return nil
}

func (c *jsonCache) load() {
	c.once.Do(
		func() {
			defer func(start time.Time) {
				log.Debug("cache loaded", "path", c.path, "elapsed", time.Since(start))
			}(time.Now())
			err := c.doLoad()
			if err != nil {
				log.Error("failed to load cache, try updating with `leetgo cache update`", "err", err)
				return
			}
			if c.Outdated() {
				log.Warn("cache is too old, try updating with `leetgo cache update`")
			}
		},
	)
}

func (c *jsonCache) Outdated() bool {
	stat, err := os.Stat(c.path)
	if os.IsNotExist(err) {
		return true
	}
	return time.Since(stat.ModTime()) >= 14*24*time.Hour
}

func (c *jsonCache) Update() error {
	err := utils.CreateIfNotExists(c.path, false)
	if err != nil {
		return err
	}

	all, err := c.client.GetAllQuestions()
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
	log.Info("cache updated", "path", c.path)
	return nil
}

func (c *jsonCache) GetBySlug(slug string) *QuestionData {
	c.load()
	return c.slugs[slug]
}

func (c *jsonCache) GetById(id string) *QuestionData {
	defer func(start time.Time) {
		log.Debug("get by id", "elapsed", time.Since(start))
	}(time.Now())

	c.load()
	return c.frontIds[id]
}

func (c *jsonCache) GetAllQuestions() []*QuestionData {
	c.load()
	all := make([]*QuestionData, 0, len(c.slugs))
	for _, q := range c.slugs {
		all = append(all, q)
	}
	return all
}
