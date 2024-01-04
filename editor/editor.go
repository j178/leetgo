package editor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/charmbracelet/log"
	"github.com/google/shlex"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/constants"
	"github.com/j178/leetgo/lang"
)

type Opener interface {
	Open(result *lang.GenerateResult) error
}

const specialAllFiles = "{{.Files}}"

var knownEditors = map[string]Opener{
	"none": &noneEditor{},
	"vim": &editor{
		command: "vim",
		args:    []string{"-p", fmt.Sprintf("+/%s", constants.CodeBeginMarker), specialAllFiles},
	},
	"neovim": &editor{
		command: "nvim",
		args:    []string{"-p", fmt.Sprintf("+/%s", constants.CodeBeginMarker), specialAllFiles},
	},
	"vscode": &editor{command: "code", args: []string{specialAllFiles}},
}

type noneEditor struct{}

func (e *noneEditor) Open(result *lang.GenerateResult) error {
	log.Info("none editor is used, skip opening files")
	return nil
}

type editor struct {
	command string
	args    []string
}

// substituteArgs substitutes the special arguments with the actual values.
func (ed *editor) substituteArgs(result *lang.GenerateResult) ([]string, error) {
	getPath := func(fileType lang.FileType) string {
		f := result.GetFile(fileType)
		if f == nil {
			return ""
		}
		return f.GetPath()
	}

	data := struct {
		Folder          string
		Files           string
		CodeFile        string
		TestFile        string
		DescriptionFile string
		TestCasesFile   string
	}{
		Folder:          result.TargetDir(),
		Files:           specialAllFiles,
		CodeFile:        getPath(lang.CodeFile),
		TestFile:        getPath(lang.TestFile),
		DescriptionFile: getPath(lang.DocFile),
		TestCasesFile:   getPath(lang.TestCasesFile),
	}

	args := make([]string, len(ed.args))
	copy(args, ed.args)
	for i, arg := range args {
		if !strings.Contains(arg, "{{") {
			continue
		}

		tmpl := template.New("")
		_, err := tmpl.Parse(arg)
		if err != nil {
			return nil, err
		}
		var s strings.Builder
		err = tmpl.Execute(&s, data)
		if err != nil {
			return nil, err
		}
		args[i] = s.String()
	}

	// replace the special marker with all files
	for i, arg := range args {
		if arg == specialAllFiles {
			allFiles := make([]string, len(result.Files))
			for j, f := range result.Files {
				allFiles[j] = f.GetPath()
			}
			args = append(args[:i], append(allFiles, args[i+1:]...)...)
			break
		}
	}

	return args, nil
}

func (ed *editor) Open(result *lang.GenerateResult) error {
	args, err := ed.substituteArgs(result)
	if err != nil {
		return fmt.Errorf("invalid editor command: %w", err)
	}
	return runCmd(ed.command, args, result.OutDir)
}

// Get returns the editor with the given name.
func Get(ed config.Editor) Opener {
	if ed.Use == "custom" {
		args, _ := shlex.Split(ed.Args)
		return &editor{
			command: ed.Command,
			args:    args,
		}
	}
	return knownEditors[ed.Use]
}

// Open opens the files in the given result with the configured editor.
func Open(result *lang.GenerateResult) error {
	cfg := config.Get()
	ed := Get(cfg.Editor)
	if ed == nil {
		return fmt.Errorf(
			"editor not supported: %s, you can use `editor.command` to customize the command",
			cfg.Editor.Use,
		)
	}
	return ed.Open(result)
}

func runCmd(command string, args []string, dir string) error {
	cmd := exec.Command(command, args...)
	if log.GetLevel() <= log.DebugLevel {
		log.Info("opening files", "command", cmd.String())
	} else {
		log.Info("opening files", "command", cmd.Path)
	}
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
