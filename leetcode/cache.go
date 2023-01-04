package leetcode

import (
	"github.com/j178/leetgo/config"
)

type QuestionsCache interface {
	GetBySlug(slug string) *QuestionData
	GetById(id string) *QuestionData
	GetCacheFile() string
	Update(client Client) error
}

func GetCache() QuestionsCache {
	if lazyCache == nil {
		cfg := config.Get()
		lazyCache = newCache(cfg.LeetCodeCacheBaseName())
	}

	return lazyCache
}

var (
	lazyCache QuestionsCache
)
