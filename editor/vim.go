package editor

import (
	"fmt"

	"github.com/hashicorp/go-hclog"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/lang"
)

type vim struct {
	command string
}

func (e *vim) cmd() string {
	if e.command == "" {
		e.command = "vim"
	}

	return e.command
}

func (e *vim) args(file lang.FileOutput) []string {
	return []string{"-p", fmt.Sprintf("+/%s", config.CodeBeginMarker)}
}

func (e *vim) Open(file lang.FileOutput) error {
	hclog.L().Info("opening files with", "editor", e.cmd())
	return runCmd(e.cmd(), e.args(file), file.Path)
}

func (e *vim) OpenMulti(files ...lang.FileOutput) error {
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = f.Path
	}
	hclog.L().Info("opening files with", "editor", e.cmd())
	return runCmd(e.cmd(), e.args(files[0]), paths...)
}
