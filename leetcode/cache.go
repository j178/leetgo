package leetcode

import (
	"sync"

	"github.com/j178/leetgo/config"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type QuestionsCache interface {
	GetBySlug(slug string) *QuestionData
	GetById(id string) *QuestionData
	GetCacheFile() string
	Update(client Client) error
}

func GetCache() QuestionsCache {
	once.Do(
		func() {
			cfg := config.Get()
			lazyCache = newCache(cfg.LeetCodeCacheBaseName())
		},
	)
	return lazyCache
}

var (
	lazyCache QuestionsCache
	once      sync.Once
)
