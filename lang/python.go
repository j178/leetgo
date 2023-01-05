package lang

import "github.com/j178/leetgo/leetcode"

type python struct {
	baseLang
}

func (python) Generate(*leetcode.QuestionData) (*GenerateResult, error) {
	return nil, NotImplemented
}
