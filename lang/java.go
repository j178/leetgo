package lang

import (
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/charmbracelet/log"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/constants"
	"github.com/j178/leetgo/leetcode"
	javaEmbed "github.com/j178/leetgo/testutils/java"
	"github.com/j178/leetgo/utils"
)

const (
	pomTemplate = `
<project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
  <modelVersion>4.0.0</modelVersion>

  <groupId>%s</groupId>
  <artifactId>%s</artifactId>
  <version>1.0</version>
  <packaging>jar</packaging>

  <name>leetcode-solutions</name>

  <properties>
    <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
  </properties>

  <dependencies>
    <dependency>
      <groupId>io.github.j178</groupId>
	    <artifactId>leetgo-java</artifactId>
	    <version>1.0</version>
    </dependency>
  </dependencies>

  <build>
    <plugins>
      <plugin>
        <groupId>org.codehaus.mojo</groupId>
        <artifactId>exec-maven-plugin</artifactId>
        <version>3.1.1</version>
      </plugin>
    </plugins>
  </build>
</project>
`
)

type java struct {
	baseLang
}

func mvnwCmd(dir string, args ...string) []string {
	cmdName := "mvnw"
	if runtime.GOOS == "windows" {
		cmdName = "mvnw.cmd"
	}
	return append([]string{filepath.Join(dir, cmdName), "-q"}, args...)
}

func (j java) HasInitialized(outDir string) (bool, error) {
	return utils.IsExist(filepath.Join(outDir, "pom.xml")), nil
}

var nameReplacer = regexp.MustCompile(`[^a-zA-Z0-9_]`)

func (j java) groupID() string {
	cfg := config.Get()
	if cfg.Code.Java.GroupID != "" {
		return cfg.Code.Java.GroupID
	}
	name := nameReplacer.ReplaceAllString(strings.ToLower(cfg.Author), "_")
	return "io.github." + name
}

func (j java) sourceDir() string {
	groupID := j.groupID()
	path := []string{"src", "main", "java"}
	path = append(path, strings.Split(groupID, ".")...)
	return filepath.Join(path...)
}

func (j java) Initialize(outDir string) error {
	log.Info("initializing java project", "outDir", utils.RelToCwd(outDir))
	// Copy mvn wrapper from embed
	err := utils.CopyFS(outDir, javaEmbed.MvnWrapper)
	if err != nil {
		return err
	}

	groupID := j.groupID()
	artifactID := "leetcode-solutions"

	// Write pom.xml
	pomXML := fmt.Sprintf(pomTemplate, groupID, artifactID)
	err = utils.WriteFile(filepath.Join(outDir, "pom.xml"), []byte(pomXML))
	if err != nil {
		return err
	}

	// Create layout
	sourceDir := filepath.Join(outDir, j.sourceDir())
	err = utils.MakeDir(sourceDir)
	if err != nil {
		return err
	}
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

	buildCmd := mvnwCmd(outDir, "compile")
	err = buildTest(q, genResult, buildCmd)
	if err != nil {
		return false, fmt.Errorf("build failed: %w", err)
	}

	filenameTmpl := getFilenameTemplate(q, j)
	packageName, err := q.GetFormattedFilename(j.slug, filenameTmpl)
	if err != nil {
		return false, err
	}
	className := fmt.Sprintf("%s.%s.Main", j.groupID(), packageName)
	execCmd := mvnwCmd(outDir, "exec:java", "-Dexec.mainClass="+className)
	return runTest(q, genResult, execCmd, targetCase)
}

func (j java) generateNormalTestCode(q *leetcode.QuestionData) (string, error) {
	code := `
public class Main {
    public static void main(String[] args) {
		Solution solution = new Solution();
	}
}
`
	return code, nil
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
	packageName string,
	filename string,
	blocks []config.Block,
	modifiers []ModifierFunc,
	separateDescriptionFile bool,
) (
	FileOutput,
	error,
) {
	codeHeader := fmt.Sprintf(
		`package %s;

import %s.*;
`, packageName, constants.JavaTestUtilsGroupId,
	)
	blocks = append(
		[]config.Block{
			{
				Name:     beforeBeforeMarker,
				Template: codeHeader,
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
		Type:     CodeFile,
	}, nil
}

func (j java) generateTestFile(
	q *leetcode.QuestionData,
	packageName string,
	filename string,
) (
	FileOutput,
	error,
) {
	content, err := j.generateTestContent(q)
	if err != nil {
		return FileOutput{}, err
	}
	content = fmt.Sprintf(
		`package %s;
%s`, packageName, content)
	return FileOutput{
		Filename: filename,
		Content:  content,
		Type:     TestFile,
	}, nil
}

func (j java) GeneratePaths(q *leetcode.QuestionData) (*GenerateResult, error) {
	filenameTmpl := getFilenameTemplate(q, j)
	packageName, err := q.GetFormattedFilename(j.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}
	genResult := &GenerateResult{
		Question: q,
		Lang:     j,
		SubDir:   filepath.Join(j.sourceDir(), packageName),
	}
	genResult.AddFile(
		FileOutput{
			Filename: "Solution.java",
			Type:     CodeFile,
		},
	)
	genResult.AddFile(
		FileOutput{
			Filename: "Main.java",
			Type:     TestFile,
		})
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
	packageName, err := q.GetFormattedFilename(j.slug, filenameTmpl)
	if err != nil {
		return nil, err
	}
	genResult := &GenerateResult{
		Question: q,
		Lang:     j,
		SubDir:   filepath.Join(j.sourceDir(), packageName),
	}

	separateDescriptionFile := separateDescriptionFile(j)
	blocks := getBlocks(j)
	modifiers, err := getModifiers(j, builtinModifiers)
	if err != nil {
		return nil, err
	}

	fqPackageName := fmt.Sprintf("%s.%s", j.groupID(), packageName)
	codeFile, err := j.generateCodeFile(q, fqPackageName, "Solution.java", blocks, modifiers, separateDescriptionFile)
	if err != nil {
		return nil, err
	}

	testFile, err := j.generateTestFile(q, fqPackageName, "Main.java")
	if err != nil {
		return nil, err
	}

	testcaseFile, err := j.generateTestCasesFile(q, "testcases.txt")
	if err != nil {
		return nil, err
	}

	genResult.AddFile(codeFile)
	genResult.AddFile(testFile)
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
