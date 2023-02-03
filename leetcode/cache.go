package leetcode

import (
	"sync"

	"github.com/j178/leetgo/config"
)

type QuestionsCache interface {
	GetBySlug(slug string) *QuestionData
	GetById(id string) *QuestionData
	GetAllQuestions() []*QuestionData
	Update() error
	CacheFile() string
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
