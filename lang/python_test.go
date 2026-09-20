package lang

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/j178/leetgo/leetcode"
)

func TestPythonGenerateNormalTestCode_VoidReturn(t *testing.T) {
	question := &leetcode.QuestionData{
		TitleSlug: "rotate-image",
		MetaData: leetcode.MetaData{
			Name: "rotate",
			Params: []leetcode.MetaDataParam{
				{Name: "matrix", Type: "integer[][]"},
			},
			Return: &leetcode.MetaDataReturn{Type: "void"},
			Output: &leetcode.MetaDataOutput{ParamIndex: 0},
		},
	}

	code, err := (python{}).generateNormalTestCode(question)
	if err != nil {
		t.Fatal(err)
	}

	want := `if __name__ == "__main__":
	matrix: List[List[int]] = deserialize("List[List[int]]", read_line())
	Solution().rotate(matrix)
	ans = matrix
	print("\noutput:", serialize(ans, "List[List[int]]"))
`
	if code != want {
		t.Fatalf("generated Python test code mismatch:\n--- want\n%s\n--- got\n%s", want, code)
	}

	pythonExe, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 is not installed")
	}
	program := `import json
import sys
from typing import *

def read_line():
    return sys.stdin.readline().strip()

def deserialize(_, value):
    return json.loads(value)

def serialize(value, _):
    return json.dumps(value, separators=(",", ":"))

class Solution:
    def rotate(self, matrix):
        matrix[:] = [list(row) for row in zip(*matrix[::-1])]

` + code
	programPath := filepath.Join(t.TempDir(), "solution.py")
	if err := os.WriteFile(programPath, []byte(program), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, pythonExe, programPath)
	command.Stdin = strings.NewReader("[[1,2,3],[4,5,6],[7,8,9]]\n")
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("generated Python program timed out: %v", ctx.Err())
	}
	if err != nil {
		t.Fatalf("generated Python program failed: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "output: [[7,4,1],[8,5,2],[9,6,3]]") {
		t.Fatalf("unexpected generated Python output: %q", output)
	}
}
