package editor

import (
	"fmt"

	"github.com/hashicorp/go-hclog"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/lang"
)

type vim struct{}

func (e *vim) args(file lang.FileOutput) []string {
	return []string{"-p", fmt.Sprintf("+/%s", config.CodeBeginMarker)}
}

func (e *vim) Open(file lang.FileOutput) error {
	hclog.L().Info("opening files with vim")
	return runCmd("vim", e.args(file), file.Path)
}

func (e *vim) OpenMulti(files ...lang.FileOutput) error {
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = f.Path
	}
	hclog.L().Info("opening files with vim")
	return runCmd("vim", e.args(files[0]), paths...)
}
