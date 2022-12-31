package leetcode

import (
	"github.com/j178/leetgo/config"
)

type questionRecord struct {
	FrontendId string   `json:"frontendId"`
	Slug       string   `json:"slug"`
	Title      string   `json:"title"`
	CnTitle    string   `json:"cnTitle"`
	Difficulty string   `json:"difficulty"`
	Tags       []string `json:"tags"`
	PaidOnly   bool     `json:"paidOnly"`
}

type QuestionsCache interface {
	GetBySlug(slug string) *questionRecord
	GetById(id string) *questionRecord
	Update(client Client) error
}

func GetCache() QuestionsCache {
	if lazyCache == nil {
		lazyCache = newCache(config.Get().LeetCodeCacheFile())
	}

	return lazyCache
}

var (
	lazyCache QuestionsCache
)
