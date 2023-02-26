package lang

import (
	"github.com/j178/leetgo/leetcode"
)

func judgeResult(q *leetcode.QuestionData, actual, expected string) bool {
	// TODO compare by question rules
	return actual == expected
}
