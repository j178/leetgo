package lang

import (
	"fmt"
	"strings"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
)

type Modifier func(string, *leetcode.QuestionData) string

func addCodeMark(commentMark string) Modifier {
	return func(s string, q *leetcode.QuestionData) string {
		cfg := config.Get()
		return fmt.Sprintf(
			"%s %s\n\n%s\n\n%s %s",
			commentMark,
			cfg.Code.CodeBeginMark,
			s,
			commentMark,
			cfg.Code.CodeEndMark,
		)
	}
}

func removeComments(code string, q *leetcode.QuestionData) string {
	lines := strings.Split(code, "\n")
	var newLines []string
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "/**") && (strings.Contains(
			lines[i+1],
			"object will be instantiated and called",
		) || strings.Contains(lines[i+1], "Definition for")) {
			for {
				i++
				if strings.HasSuffix(lines[i], "*/") {
					break
				}
			}
			continue
		}
		newLines = append(newLines, line)
	}
	return strings.Join(newLines, "\n")
}

func prepend(s string) Modifier {
	return func(code string, q *leetcode.QuestionData) string {
		return s + code
	}
}
