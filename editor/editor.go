package editor

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/lang"
)

type Opener interface {
	Open(file lang.FileOutput) error
}

type MultiOpener interface {
	Opener
	OpenMulti(files ...lang.FileOutput) error
}

var editors = map[string]Opener{
	"none":   &noneEditor{},
	"custom": &customEditor{},
	"vim":    &vim{command: "vim"},
	"neovim": &vim{command: "nvim"},
	"vscode": &commonMultiEditor{commonEditor{command: "code"}},
	"goland": &commonEditor{command: "goland"},
}

type noneEditor struct{}

func (e *noneEditor) Open(file lang.FileOutput) error {
	hclog.L().Info("none editor is used, skip opening files")
	return nil
}

type commonEditor struct {
	command string
	args    []string
}

type commonMultiEditor struct {
	commonEditor
}

func (e *commonEditor) Open(file lang.FileOutput) error {
	hclog.L().Info("opening files with command", "command", e.command)
	return runCmd(e.command, e.args, file.Path)
}

type customEditor struct{}

func (e *customEditor) Open(file lang.FileOutput) error {
	cfg := config.Get()
	if cfg.Editor.Command == "" {
		hclog.L().Warn("editor.command is empty, skip opening files")
		return nil
	}
	hclog.L().Info("opening files with custom command", "command", cfg.Editor.Command)
	return runCmd(cfg.Editor.Command, cfg.Editor.Args, file.Path)
}

func (e *commonMultiEditor) OpenMulti(files ...lang.FileOutput) error {
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = f.Path
	}
	return runCmd(e.command, e.args, paths...)
}

func Get(s string) Opener {
	return editors[s]
}

func Open(paths []lang.FileOutput) error {
	if len(paths) == 0 {
		return nil
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
		return ed.OpenMulti(paths...)
	}
	return ed.Open(paths[0])
}

func runCmd(command string, args []string, files ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Args = append(cmd.Args, files...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
