package lang

import "github.com/j178/leetgo/leetcode"

var pythonGen = python{
	baseLang{
		Name:              "Python",
		Suffix:            ".py",
		LineComment:       "#",
		BlockCommentStart: `"""`,
		BlockCommentEnd:   `"""`,
	},
}

type python struct {
	baseLang
}

func (p python) Name() string {
	return p.baseLang.Name
}

func (p python) Generate(q leetcode.QuestionData) []FileOutput {
	// TODO implement me
	panic("implement me")
}

func (p python) GenerateContest(leetcode.Contest) []FileOutput {
	// TODO implement me
	panic("implement me")
}
