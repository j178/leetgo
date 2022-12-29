package lang

import "github.com/j178/leetgo/leetcode"

type python struct {
	baseLang
}

func (p python) Name() string {
	return p.baseLang.Name
}

func (p python) Generate(q leetcode.QuestionData) []any {
	// TODO implement me
	panic("implement me")
}

func (p python) GenerateContest() []any {
	// TODO implement me
	panic("implement me")
}
