package lang

import "github.com/j178/leetgo/leetcode"

type python struct {
	baseLang
}

func (python) Generate(leetcode.QuestionData) ([]FileOutput, error) {
	return nil, NotImplemented
}

func (p python) GenerateTest(leetcode.QuestionData) ([]FileOutput, error) {
	return nil, NotImplemented
}

func (python) GenerateContest(leetcode.Contest) ([]FileOutput, error) {
	return nil, NotImplemented
}

func (python) GenerateContestTest(leetcode.Contest) ([]FileOutput, error) {
	return nil, NotImplemented
}
