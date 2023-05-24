package lang

import (
	"math"
	"strconv"
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

	a, _ := goutils.SplitArray(actual)
	b, _ := goutils.SplitArray(expected)
	if len(a) != len(b) {
		return false
	}

	if j.ignoreOrder {
		return j.compareIgnoringOrder(a, b)
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TODO improve the detection of "any order"
func shouldIgnoreOrder(q *leetcode.QuestionData) bool {
	content := q.GetEnglishContent()
	content = strip.StripTags(content)
	if strings.Contains(content, "return the answer in any order") {
		return true
	}

	// try translated content
	content = q.TranslatedContent
	content = strip.StripTags(content)
	// nolint: gosimple
	if strings.Contains(content, "任意顺序返回答案") {
		return true
	}
	return false
}

func (j *sliceJudger) compareIgnoringOrder(actual, expected []string) bool {
	cnt := map[string]int{}
	for _, v := range actual {
		cnt[v]++
	}
	for _, v := range expected {
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

// floatCompare compares two float numbers. Returns true if the difference is less than 1e-5.
func floatCompare(actual, expected string) bool {
	a, _ := strconv.ParseFloat(actual, 64)
	b, _ := strconv.ParseFloat(expected, 64)
	return math.Abs(a-b) < 1e-5
}

func GetJudger(q *leetcode.QuestionData) Judger {
	// TODO compare by question rules

	var judger Judger = judgeFunc(stringCompare)
	if q.MetaData.SystemDesign {
		judger = &systemDesignJudger{q}
	} else {
		resultType := q.MetaData.ResultType()
		switch resultType {
		case "double":
			judger = judgeFunc(floatCompare)
		default:
			if strings.HasSuffix(resultType, "[]") {
				judger = newSliceJudger(q)
			}
		}
	}
	return judger
}
