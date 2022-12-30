package lang

import (
	"github.com/j178/leetgo/leetcode"
)

type golang struct {
	baseLang
}

func (g golang) Generate(q *leetcode.QuestionData) ([]FileOutput, error) {
	return g.baseLang.Generate(q)
}

func (g golang) GenerateTest(q *leetcode.QuestionData) ([]FileOutput, error) {
	return g.baseLang.GenerateTest(q)
}
