package leetcode

import (
	"sync"

	"github.com/j178/leetgo/config"
)

type QuestionsCache interface {
	CacheFile() string
	GetBySlug(slug string) *QuestionData
	GetById(id string) *QuestionData
	GetAllQuestions() []*QuestionData
	Outdated() bool
	Update() error
}

func GetCache(c Client) QuestionsCache {
	once.Do(
		func() {
			cfg := config.Get()
			lazyCache = newCache(cfg.QuestionCacheFile(cacheExt), c)
		},
	)
	return lazyCache
}

var (
	lazyCache QuestionsCache
	once      sync.Once
)
