package lang

import (
	"fmt"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/utils"
)

type java struct {
	baseLang
}

func (j java) HasInitialized(outDir string) (bool, error) {
	return false, nil
}

func (j java) Initialize(outDir string) error {
	return nil
}

func (j java) RunLocalTest(q *leetcode.QuestionData, outDir string, targetCase string) (bool, error) {
	genResult, err := j.GeneratePaths(q)
	if err != nil {
		return false, fmt.Errorf("generate paths failed: %w", err)
	}
	genResult.SetOutDir(outDir)

	testFile := genResult.GetFile(TestFile).GetPath()
	if !utils.IsExist(testFile) {
		return false, fmt.Errorf("file %s not found", utils.RelToCwd(testFile))
	}

	execDir, err := getTempBinDir(q, j)
	if err != nil {
		return false, fmt.Errorf("get temp bin dir failed: %w", err)
	}

	args := []string{"javac", "-d", execDir, testFile}
	err = buildTest(q, genResult, args)
	if err != nil {
		return false, fmt.Errorf("build failed: %w", err)
	}

	args = []string{"java", "--class-path", execDir, "Main"}
	return runTest(q, genResult, args, targetCase)
}

func (j java) generateNormalTestCode(q *leetcode.QuestionData) (string, error) {
	return "", nil
}

func (j java) generateSystemDesignTestCode(q *leetcode.QuestionData) (string, error) {
	return "", nil
}

func (j java) generateTestContent(q *leetcode.QuestionData) (string, error) {
	if q.MetaData.SystemDesign {
		return j.generateSystemDesignTestCode(q)
	}
	return j.generateNormalTestCode(q)
}

func (j java) generateCodeFile(
	q *leetcode.QuestionData,
	filename string,
	blocks []config.Block,
	modifiers []ModifierFunc,
	separateDescriptionFile bool,
) (
	FileOutput,
	error,
) {
	codeHeader := ""
	testContent, err := j.generateTestContent(q)
	if err != nil {
		return FileOutput{}, err
	}
	blocks = append(
		[]config.Block{
			{
				Name:     beforeBeforeMarker,
				Template: codeHeader,
			},
			{
				Name:     afterAfterMarker,
				Template: testContent,
			},
		},
		blocks...,
	)
	content, err := j.generateCodeContent(
		q,
		blocks,
		modifiers,
		separateDescriptionFile,
	)
	if err != nil {
		return FileOutput{}, err
	}
	return FileOutput{
		Filename: filename,
		Content:  content,
		Type:     CodeFile | TestFile,
	}, nil
}

func (j java) GeneratePaths(q *leetcode.QuestionData) (*GenerateResult, error) {
	filenameTmpl := getFilenameTemplate(q, j)
	baseFilename, err := q.GetFormattedFilename(j.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}
	genResult := &GenerateResult{
		SubDir:   baseFilename,
		Question: q,
		Lang:     j,
	}
	genResult.AddFile(
		FileOutput{
			Filename: "solution.java",
			Type:     CodeFile | TestFile,
		},
	)
	genResult.AddFile(
		FileOutput{
			Filename: "testcases.txt",
			Type:     TestCasesFile,
		},
	)
	if separateDescriptionFile(j) {
		genResult.AddFile(
			FileOutput{
				Filename: "question.md",
				Type:     DocFile,
			},
		)
	}
	return genResult, nil
}

func (j java) Generate(q *leetcode.QuestionData) (*GenerateResult, error) {
	filenameTmpl := getFilenameTemplate(q, j)
	baseFilename, err := q.GetFormattedFilename(j.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}
	genResult := &GenerateResult{
		Question: q,
		Lang:     j,
		SubDir:   baseFilename,
	}

	separateDescriptionFile := separateDescriptionFile(j)
	blocks := getBlocks(j)
	modifiers, err := getModifiers(j, builtinModifiers)
	if err != nil {
		return nil, err
	}
	codeFile, err := j.generateCodeFile(q, "solution.java", blocks, modifiers, separateDescriptionFile)
	if err != nil {
		return nil, err
	}
	testcaseFile, err := j.generateTestCasesFile(q, "testcases.txt")
	if err != nil {
		return nil, err
	}
	genResult.AddFile(codeFile)
	genResult.AddFile(testcaseFile)

	if separateDescriptionFile {
		docFile, err := j.generateDescriptionFile(q, "question.md")
		if err != nil {
			return nil, err
		}
		genResult.AddFile(docFile)
	}

	return genResult, nil
}
