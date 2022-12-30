package lang

import "github.com/j178/leetgo/leetcode"

var pythonGen = python{
	baseLang{
		Name:              "Python",
		ShortName:         "py",
		Suffix:            ".py",
		LineComment:       "#",
		BlockCommentStart: `"""`,
		BlockCommentEnd:   `"""`,
	},
}

type python struct {
	baseLang
}

func (p python) ShortName() string {
	return p.baseLang.ShortName
}

func (p python) Name() string {
	return p.baseLang.Name
}

func (python) Generate(leetcode.QuestionData) ([]FileOutput, error) {
	return nil, NotSupported
}

func (p python) GenerateTest(leetcode.QuestionData) ([]FileOutput, error) {
	return nil, NotSupported
}

func (python) GenerateContest(leetcode.Contest) ([]FileOutput, error) {
	return nil, NotSupported
}

func (python) GenerateContestTest(leetcode.Contest) ([]FileOutput, error) {
	return nil, NotSupported
}
