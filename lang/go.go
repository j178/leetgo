package lang

import (
	"path/filepath"

	"github.com/j178/leetgo/leetcode"
)

var (
	testutilsModPath = "github.com/j178/leetgo/testutils/go"
)

type golang struct {
	baseLang
}

func addNamedReturn(code string, q *leetcode.QuestionData) string {
	return code
}

func changeReceiverName(code string, q *leetcode.QuestionData) string {
	return code
}

func (g golang) CheckLibrary() bool {
	// 执行 go list -m json 查看是否有 github.com/j178/leetgo/testutils/go 的依赖
	// go list -m -json github.com/j178/leetgo/testutils/go => not a known dependency
	return true
}

func (g golang) GenerateLibrary() error {
	// 执行 go mod init & go get
	return nil
}

func (g golang) RunTest(q *leetcode.QuestionData) error {
	return nil
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
