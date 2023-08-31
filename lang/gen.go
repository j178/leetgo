package lang

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/charmbracelet/log"
	"github.com/spf13/viper"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/constants"
	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/utils"
)

func GetGenerator(lang string) (Lang, error) {
	lang = strings.ToLower(lang)
	for _, l := range SupportedLangs {
		if l.Slug() == lang {
			return l, nil
		}
	}
	for _, l := range SupportedLangs {
		if l.ShortName() == lang {
			return l, nil
		}
	}
	for _, l := range SupportedLangs {
		if strings.HasPrefix(strings.ToLower(l.Name()), lang) {
			return l, nil
		}
	}
	return nil, fmt.Errorf("language %s is not supported yet, welcome to send a PR", lang)
}

func generate(q *leetcode.QuestionData) (Lang, *GenerateResult, error) {
	cfg := config.Get()
	gen, err := GetGenerator(cfg.Code.Lang)
	if err != nil {
		return nil, nil, err
	}

	err = q.Fulfill()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get question data: %w", err)
	}

	codeSnippet := q.GetCodeSnippet(gen.Slug())
	if codeSnippet == "" {
		if len(q.CodeSnippets) <= 3 {
			langs := make([]string, 0, len(q.CodeSnippets))
			for _, snippet := range q.CodeSnippets {
				langs = append(langs, snippet.Lang)
			}
			return nil, nil, fmt.Errorf(
				`question %q doesn't support language %s, it only supports %s`,
				q.TitleSlug,
				gen.Slug(),
				strings.Join(langs, ","),
			)
		}
		return nil, nil, fmt.Errorf(`question %q doesn't support language %q`, q.TitleSlug, cfg.Code.Lang)
	}

	outDir := getOutDir(q, gen)
	err = utils.CreateIfNotExists(outDir, true)
	if err != nil {
		return nil, nil, err
	}

	// Check and generate necessary library files.
	if t, ok := gen.(NeedInitialization); ok {
		ok, err := t.HasInitialized(outDir)
		if err == nil && !ok {
			err = t.Initialize(outDir)
			if err != nil {
				return nil, nil, err
			}
		}
		if err != nil {
			log.Error(
				"check initialization failed, skip initialization",
				"lang", gen.Slug(),
				"err", err,
			)
		}
	}

	// Generate files
	result, err := gen.Generate(q)
	if err != nil {
		return nil, nil, err
	}
	result.SetOutDir(outDir)

	for _, hook := range result.ResultHooks {
		err := hook(result)
		if err != nil {
			return nil, nil, err
		}
	}

	// Write files
	for i, file := range result.Files {
		written, err := tryWrite(file.GetPath(), file.Content)
		if err != nil {
			log.Error("failed to write file", "path", utils.RelToCwd(file.GetPath()), "err", err)
			continue
		}
		result.Files[i].Written = written
	}
	return gen, result, nil
}

func Generate(q *leetcode.QuestionData) (*GenerateResult, error) {
	gen, result, err := generate(q)
	if err != nil {
		return nil, err
	}

	state := config.LoadState()
	state.LastQuestion = config.LastQuestion{
		Slug:       q.TitleSlug,
		FrontendID: q.QuestionFrontendId,
		Gen:        gen.Slug(),
	}
	config.SaveState(state)

	return result, nil
}

func GenerateContest(ct *leetcode.Contest) ([]*GenerateResult, error) {
	qs, err := ct.GetAllQuestions()
	if err != nil {
		return nil, err
	}

	var results []*GenerateResult
	for _, q := range qs {
		_, result, err := generate(q)
		if err != nil {
			log.Error("failed to generate", "question", q.TitleSlug, "err", err)
			continue
		}
		results = append(results, result)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no question generated")
	}

	state := config.LoadState()
	state.LastContest = ct.TitleSlug
	config.SaveState(state)

	return results, nil
}

func tryWrite(file string, content string) (bool, error) {
	write := true
	relPath := utils.RelToCwd(file)
	if utils.IsExist(file) {
		if !viper.GetBool("yes") {
			prompt := &survey.Confirm{Message: fmt.Sprintf("File \"%s\" already exists, overwrite?", relPath)}
			err := survey.AskOne(prompt, &write)
			if err != nil {
				return false, err
			}
		}
	}
	if !write {
		return false, nil
	}

	err := utils.WriteFile(file, []byte(content))
	if err != nil {
		return false, err
	}

	log.Info("generated", "file", relPath)
	return true, nil
}

// GeneratePathsOnly runs generate process but does not generate real content.
func GeneratePathsOnly(q *leetcode.QuestionData) (*GenerateResult, error) {
	cfg := config.Get()
	gen, err := GetGenerator(cfg.Code.Lang)
	if err != nil {
		return nil, err
	}

	result, err := gen.GeneratePaths(q)
	if err != nil {
		return nil, err
	}

	outDir := getOutDir(q, gen)
	result.SetOutDir(outDir)
	return result, nil
}

func GetSolutionCode(q *leetcode.QuestionData) (string, error) {
	result, err := GeneratePathsOnly(q)
	if err != nil {
		return "", err
	}
	codeFile := result.GetFile(CodeFile)
	if codeFile == nil {
		return "", fmt.Errorf("no code file generated")
	}
	code, err := codeFile.GetContent()
	if err != nil {
		return "", err
	}
	codeLines := strings.Split(code, "\n")
	var codeLinesToKeep []string
	inCode := false
	for _, line := range codeLines {
		if !inCode && strings.Contains(line, constants.CodeBeginMarker) {
			inCode = true
			continue
		}
		if inCode && strings.Contains(line, constants.CodeEndMarker) {
			break
		}
		if inCode {
			codeLinesToKeep = append(codeLinesToKeep, line)
		}
	}

	nonEmptyLines := 0
	for _, line := range codeLinesToKeep {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
		}
	}
	if nonEmptyLines == 0 {
		return "", fmt.Errorf("no code found in %s", codeFile.GetPath())
	}

	return strings.Join(codeLinesToKeep, "\n"), nil
}

func UpdateSolutionCode(q *leetcode.QuestionData, newCode string) error {
	result, err := GeneratePathsOnly(q)
	if err != nil {
		return err
	}
	codeFile := result.GetFile(CodeFile)
	if codeFile == nil {
		return fmt.Errorf("no code file generated")
	}
	code, err := codeFile.GetContent()
	if err != nil {
		return err
	}
	lines := strings.Split(code, "\n")
	var newLines []string
	skip := false
	for _, line := range lines {
		if strings.Contains(line, constants.CodeBeginMarker) {
			newLines = append(newLines, line+"\n")
			newLines = append(newLines, newCode)
			skip = true
		} else if strings.Contains(line, constants.CodeEndMarker) {
			newLines = append(newLines, line)
			skip = false
		} else if !skip {
			newLines = append(newLines, line)
		}
	}

	newContent := strings.Join(newLines, "\n")
	err = os.WriteFile(codeFile.GetPath(), []byte(newContent), 0o644)
	if err != nil {
		return err
	}
	log.Info("updated", "file", utils.RelToCwd(codeFile.GetPath()))
	return nil
}

func GetTestCases(q *leetcode.QuestionData) (TestCases, error) {
	result, err := GeneratePathsOnly(q)
	if err != nil {
		return TestCases{}, err
	}
	testCasesFile := result.GetFile(TestCasesFile)
	if testCasesFile == nil {
		return TestCases{}, fmt.Errorf("no test cases file generated")
	}
	return ParseTestCases(q, testCasesFile)
}
