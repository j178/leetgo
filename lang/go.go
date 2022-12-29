package lang

import "github.com/j178/leetgo/leetcode"

type golang struct {
	baseLang
}

func (g golang) Name() string {
	return g.baseLang.Name
}

func (golang) Generate(q leetcode.QuestionData) []any {
	return nil
}

func (golang) GenerateContest() []any {
	return nil
}
