package editor

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/charmbracelet/log"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/constants"
	"github.com/j178/leetgo/lang"
)

type Opener interface {
	Open(result *lang.GenerateResult) error
}

type MultiOpener interface {
	Opener
	OpenMulti(result *lang.GenerateResult) error
}

var editors = map[string]Opener{
	"none":   &noneEditor{},
	"custom": &customEditor{},
	"vim": &commonMultiEditor{
		commonEditor{
			command: "vim",
			args:    []string{"-p", fmt.Sprintf("+/%s", constants.CodeBeginMarker)},
		},
	},
	"neovim": &commonMultiEditor{
		commonEditor{
			command: "nvim",
			args:    []string{"-p", fmt.Sprintf("+/%s", constants.CodeBeginMarker)},
		},
	},
	"vscode": &commonMultiEditor{commonEditor{command: "code"}},
	"goland": &commonEditor{command: "goland"},
}

type noneEditor struct{}

func (e *noneEditor) Open(result *lang.GenerateResult) error {
	log.Info("none editor is used, skip opening files")
	return nil
}

type commonEditor struct {
	command string
	args    []string
}

func (e *commonEditor) Open(result *lang.GenerateResult) error {
	log.Info("opening file", "command", e.command)
	return runCmd(e.command, e.args, result.GetFile(lang.CodeFile).GetPath())
}

type commonMultiEditor struct {
	commonEditor
}

func (e *commonMultiEditor) OpenMulti(result *lang.GenerateResult) error {
	paths := make([]string, 0, len(result.Files))
	for _, f := range result.Files {
		paths = append(paths, f.GetPath())
	}
	log.Info("opening files", "command", e.command)
	return runCmd(e.command, e.args, paths...)
}

type customEditor struct{}

func (e *customEditor) Open(result *lang.GenerateResult) error {
	cfg := config.Get()
	if cfg.Editor.Command == "" {
		log.Warn("editor.command is empty, skip opening files")
		return nil
	}
	log.Info("opening files", "command", cfg.Editor.Command)
	return runCmd(cfg.Editor.Command, cfg.Editor.Args, result.GetFile(lang.CodeFile).GetPath())
}

func Get(s string) Opener {
	return editors[s]
}

func Open(result *lang.GenerateResult) error {
	if len(result.Files) == 0 {
		return nil
	}
	if result.GetFile(lang.CodeFile) == nil {
		return fmt.Errorf("no code file found")
	}

	cfg := config.Get()
	ed := Get(cfg.Editor.Use)
	if ed == nil {
		return fmt.Errorf(
			"editor not supported: %s, you can use `editor.command` to customize the command",
			cfg.Editor.Use,
		)
	}
	if ed, ok := ed.(MultiOpener); ok {
		return ed.OpenMulti(result)
	}
	return ed.Open(result)
}

func runCmd(command string, args []string, files ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Args = append(cmd.Args, files...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
