package common

import (
	"os"
	"strings"
	"time"

	"github.com/mitchellh/go-ps"
	C "golang.org/x/exp/constraints"
)

// IsDebugging will return true if the process was launched from Delve or the
// gopls language server debugger.
//
// It does not detect situations where a debugger attached after process start.
func IsDebugging() bool {
	pid := os.Getppid()

	// We loop in case there were intermediary processes like the gopls language server.
	for pid != 0 {
		switch p, err := ps.FindProcess(pid); {
		case p == nil || err != nil:
			return false
		case strings.HasPrefix(p.Executable(), "dlv"):
			return true
		default:
			pid = p.PPid()
		}
	}
	return false
}

func isTLE(f func()) bool {
	if DebugTLE == 0 || IsDebugging() {
		f()
		return false
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		f()
	}()
	select {
	case <-done:
		return false
	case <-time.After(DebugTLE):
		return true
	}
}

func trimSpaceAndNewLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	return strings.Join(lines, "")
}

// 移除多余空行和左右空格
func splitAndRemoveEmptyLine(s string) (res []string) {
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			res = append(res, line)
		}
	}
	return
}
