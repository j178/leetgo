package lang

import (
	"path/filepath"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
)

type golang struct {
	baseLang
}

func (g golang) Name() string {
	return g.baseLang.Name
}

func (g golang) Generate(q leetcode.QuestionData) []FileOutput {
	cfg := config.Get()
	return []FileOutput{
		{
			Filename: filepath.Join(cfg.Go.OutDir, q.TitleSlug, "solution.go"),
			Content:  "package main\n",
		},
		{
			Filename: filepath.Join(cfg.Go.OutDir, q.TitleSlug, "solution_test.go"),
			Content:  "package main\n",
		},
	}
}

func (golang) GenerateContest(leetcode.Contest) []FileOutput {
	return nil
}
