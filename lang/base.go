package lang

import "github.com/j178/leetgo/leetcode"

type baseLang struct {
	Name              string
	Suffix            string
	LineComment       string
	BlockCommentStart string
	BlockCommentEnd   string
}

type Generator interface {
	Name() string
	Generate(q leetcode.QuestionData) []any
	GenerateContest() []any
}

var SupportedLanguages = []Generator{
	golang{
		baseLang{
			Name:              "Go",
			Suffix:            ".go",
			LineComment:       "//",
			BlockCommentStart: "/*",
			BlockCommentEnd:   "*/",
		},
	},
	python{
		baseLang{
			Name:              "Python",
			Suffix:            ".py",
			LineComment:       "#",
			BlockCommentStart: `"""`,
			BlockCommentEnd:   `"""`,
		},
	},
}
