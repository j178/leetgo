//go:build !cgo && gob

package leetcode

import (
	"encoding/gob"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/log"

	"github.com/j178/leetgo/utils"
)

var cacheExt = ".gob"

type gobCache struct {
	path     string
	client   Client
	once     sync.Once
	slugs    map[string]*QuestionData
	frontIds map[string]*QuestionData
}

func newCache(path string, c Client) QuestionsCache {
	return &gobCache{path: path, client: c}
}

func (c *gobCache) CacheFile() string {
	return c.path
}

func (c *gobCache) doLoad() error {
	c.slugs = make(map[string]*QuestionData)
	c.frontIds = make(map[string]*QuestionData)

	_, err := os.Stat(c.path)
	if err != nil {
		return err
	}
	s, err := os.Open(c.path)
	if err != nil {
		return err
	}

	var records []*QuestionData
	err = gob.NewDecoder(s).Decode(&records)
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

func (c *gobCache) load() {
	c.once.Do(
		func() {
			defer func(now time.Time) {
				log.Debug("cache loaded", "path", c.path, "elapsed", time.Since(now))
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

func (c *gobCache) Outdated() bool {
	stat, err := os.Stat(c.path)
	if os.IsNotExist(err) {
		return true
	}
	return time.Since(stat.ModTime()) >= 14*24*time.Hour
}

func (c *gobCache) Update() error {
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
	enc := gob.NewEncoder(f)
	err = enc.Encode(all)
	if err != nil {
		return err
	}
	log.Info("cache updated", "path", c.path)
	return nil
}

func (c *gobCache) GetBySlug(slug string) *QuestionData {
	c.load()
	return c.slugs[slug]
}

func (c *gobCache) GetById(id string) *QuestionData {
	c.load()
	return c.frontIds[id]
}

func (c *gobCache) GetAllQuestions() []*QuestionData {
	c.load()
	all := make([]*QuestionData, 0, len(c.slugs))
	for _, q := range c.slugs {
		all = append(all, q)
	}
	return all
}
