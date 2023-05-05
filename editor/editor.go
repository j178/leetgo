package editor

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/charmbracelet/log"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/constants"
	"github.com/j178/leetgo/lang"
)

type Opener interface {
	Open(result *lang.GenerateResult) error
}

const specialAllFiles = "{{.AllFiles}}"

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

func (ed *editor) substituteArgs(result *lang.GenerateResult) error {
	getPath := func(fileType lang.FileType) string {
		f := result.GetFile(fileType)
		if f == nil {
			return ""
		}
		return f.GetPath()
	}

	const replaceWithAllFiles = "__all_files__"
	data := struct {
		AllFiles        string
		CodeFile        string
		TestFile        string
		DescriptionFile string
		TestCasesFile   string
	}{
		AllFiles:        replaceWithAllFiles,
		CodeFile:        getPath(lang.CodeFile),
		TestFile:        getPath(lang.TestFile),
		DescriptionFile: getPath(lang.DocFile),
		TestCasesFile:   getPath(lang.TestCasesFile),
	}

	for i, arg := range ed.args {
		if !strings.Contains(arg, "{{") {
			continue
		}

		tmpl := template.New("")
		_, err := tmpl.Parse(arg)
		if err != nil {
			return err
		}
		var s bytes.Buffer
		err = tmpl.Execute(&s, data)
		if err != nil {
			return err
		}
		ed.args[i] = s.String()
	}

	// replace the special marker with all files
	for i, arg := range ed.args {
		if arg == replaceWithAllFiles {
			allFiles := make([]string, 0, len(result.Files))
			for _, f := range result.Files {
				allFiles = append(allFiles, f.GetPath())
			}
			ed.args = append(ed.args[:i], append(allFiles, ed.args[i+1:]...)...)
			break
		}
	}

	return nil
}

func (ed *editor) Open(result *lang.GenerateResult) error {
	err := ed.substituteArgs(result)
	if err != nil {
		return fmt.Errorf("invalid editor command: %w", err)
	}
	return runCmd(ed.command, ed.args)
}

func Get(s string) Opener {
	if s == "custom" {
		cfg := config.Get()
		return &editor{
			command: cfg.Editor.Command,
			args:    cfg.Editor.Args,
		}
	}
	return knownEditors[s]
}

func Open(result *lang.GenerateResult) error {
	if result.GetFile(lang.CodeFile) == nil {
		return errors.New("no code file found, skip opening")
	}

	cfg := config.Get()
	ed := Get(cfg.Editor.Use)
	if ed == nil {
		return fmt.Errorf(
			"editor not supported: %s, you can use `editor.command` to customize the command",
			cfg.Editor.Use,
		)
	}
	return ed.Open(result)
}

func runCmd(command string, args []string, files ...string) error {
	cmd := exec.Command(command, args...)
	if log.GetLevel() <= log.DebugLevel {
		log.Info("opening files", "command", cmd.String())
	} else {
		log.Info("opening files", "command", cmd.Path)
	}
	cmd.Args = append(cmd.Args, files...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
