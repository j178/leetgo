package lang

import (
	"fmt"

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

// TODO: implement
func removeComments(code string, q *leetcode.QuestionData) string {
	return code
}

func prepend(s string) Modifier {
	return func(code string, q *leetcode.QuestionData) string {
		return s + code
	}
}
