package lang

import (
	"github.com/j178/leetgo/leetcode"
)

type golang struct {
	baseLang
}

func prepend(s string) Modifier {
	return func(code string, q *leetcode.QuestionData) string {
		return s + code
	}
}

func addNamedReturn(code string, q *leetcode.QuestionData) string {
	return code
}

func changeReceiverName(code string, q *leetcode.QuestionData) string {
	return code
}

func (g golang) CheckLibrary(projectRoot string) bool {
	return true
}

func (g golang) GenerateLibrary(projectRoot string) error {
	return nil
}

func (g golang) SupportTest() bool {
	return true
}

func (g golang) Generate(q *leetcode.QuestionData) ([]FileOutput, error) {
	comment := g.generateComments(q)
	code := g.generateCode(
		q,
		addCodeMark(g.lineComment),
		removeComments,
		addNamedReturn,
		changeReceiverName,
		prepend("package main\n\n"),
	)
	content := comment + "\n" + code + "\n"

	filenameTmpl := getFilenameTemplate(g)
	baseFilename, err := q.GetFormattedFilename(g.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}
	codeFile := filepath.Join(baseFilename, "solution.go")
	testFile := filepath.Join(baseFilename, "solution_test.go")

	files := []FileOutput{
		{
			Path:    codeFile,
			Content: content,
		},
		{
			Path:    testFile,
			Content: "",
		},
	}

	return files, nil
}
