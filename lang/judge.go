package lang

import (
	"strings"

	strip "github.com/grokify/html-strip-tags-go"
	"github.com/j178/leetgo/leetcode"
	goutils "github.com/j178/leetgo/testutils/go"
)

type Judger interface {
	Judge(actual, expected string) bool
}

type judgeFunc func(actual, expected string) bool

func (f judgeFunc) Judge(actual, expected string) bool {
	return f(actual, expected)
}

type systemDesignJudger struct {
	q *leetcode.QuestionData
}

func (systemDesignJudger) Judge(actual, expected string) bool {
	return stringCompare(actual, expected)
}

func stringCompare(actual, expected string) bool {
	return actual == expected
}

type sliceJudger struct {
	ignoreOrder bool
}

func newSliceJudger(q *leetcode.QuestionData) *sliceJudger {
	ignoreOrder := shouldIgnoreOrder(q)
	return &sliceJudger{ignoreOrder}
}

func (j *sliceJudger) Judge(actual, expected string) bool {
	if actual == expected {
		return true
	}

	if j.ignoreOrder {
		return j.compareIgnoringOrder(actual, expected)
	}

	return false
}

// TODO improve the detection of "any order"
func shouldIgnoreOrder(q *leetcode.QuestionData) bool {
	content := q.GetEnglishContent()
	content = strip.StripTags(content)
	// nolint: gosimple
	if strings.Contains(content, "return the answer in any order") {
		return true
	}
	return false
}

func (j *sliceJudger) compareIgnoringOrder(actual, expected string) bool {
	a, _ := goutils.SplitArray(actual)
	b, _ := goutils.SplitArray(expected)
	if len(a) != len(b) {
		return false
	}
	cnt := map[string]int{}
	for _, v := range a {
		cnt[v]++
	}
	for _, v := range b {
		cnt[v]--
		if cnt[v] < 0 {
			return false
		}
	}
	for _, v := range cnt {
		if v != 0 {
			return false
		}
	}
	return true
}

func GetJudger(q *leetcode.QuestionData) Judger {
	// TODO compare by question rules

	var judger Judger = judgeFunc(stringCompare)
	if q.MetaData.SystemDesign {
		judger = &systemDesignJudger{q}
	} else {
		resultType := q.MetaData.ResultType()
		if strings.HasSuffix(resultType, "[]") {
			judger = newSliceJudger(q)
		}
	}
	return judger
}
