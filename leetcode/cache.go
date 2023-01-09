package leetcode

import (
	"sync"

	"github.com/j178/leetgo/config"
)

type QuestionsCache interface {
	GetBySlug(slug string) *QuestionData
	GetById(id string) *QuestionData
	GetAllQuestions() []*QuestionData
	GetCacheFile() string
	Update() error
}

func GetCache(c Client) QuestionsCache {
	once.Do(
		func() {
			cfg := config.Get()
			lazyCache = newCache(cfg.LeetCodeCacheBaseName(), c)
		},
	)
	return lazyCache
}

var (
	lazyCache QuestionsCache
	once      sync.Once
)
