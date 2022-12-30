package lang

import "github.com/j178/leetgo/leetcode"

type baseLang struct {
	Name              string
	Suffix            string
	LineComment       string
	BlockCommentStart string
	BlockCommentEnd   string
}

type FileOutput struct {
	Filename string
	Content  string
}

type Generator interface {
	Name() string
	Generate(q leetcode.QuestionData) []FileOutput
	GenerateContest(c leetcode.Contest) []FileOutput
}

// defaultGenerator generates anything that other generators can't process.
type defaultGenerator struct {
}

func (d defaultGenerator) Name() string {
	return "default"
}

func (d defaultGenerator) Generate(q leetcode.QuestionData) []FileOutput {
	// TODO implement me
	panic("implement me")
}

func (d defaultGenerator) GenerateContest(c leetcode.Contest) []FileOutput {
	// TODO implement me
	panic("implement me")
}

var SupportedLanguages = []Generator{
	golangGen,
	pythonGen,
}
