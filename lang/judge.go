package lang

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/goccy/go-json"
	strip "github.com/grokify/html-strip-tags-go"

	"github.com/j178/leetgo/leetcode"
	goutils "github.com/j178/leetgo/testutils/go"
)

type JudgeResult interface {
	IsAccepted() bool
	GetInfo() string
}

type simpleResult struct {
	accepted bool
	info     string
}

func failed(info string) JudgeResult {
	return simpleResult{false, info}
}

func accepted() JudgeResult {
	return simpleResult{true, ""}
}

func (r simpleResult) IsAccepted() bool {
	return r.accepted
}

func (r simpleResult) GetInfo() string {
	return r.info
}

type Judger interface {
	Judge(input []string, output, actualOutput string) JudgeResult
}

type stringJudger struct{}

func (stringJudger) Judge(input []string, output, actualOutput string) JudgeResult {
	if output != actualOutput {
		return failed(fmt.Sprintf("expected %q, got %q", output, actualOutput))
	}
	return accepted()
}

type sliceJudger struct {
	ignoreOrder bool
	subJudger   Judger
}

func newSliceJudger(ignoreOrder bool, subJudger Judger) *sliceJudger {
	return &sliceJudger{ignoreOrder, subJudger}
}

func (j *sliceJudger) Judge(input []string, output, actualOutput string) JudgeResult {
	if output == actualOutput {
		return accepted()
	}

	a, _ := goutils.SplitArray(output)
	b, _ := goutils.SplitArray(actualOutput)
	if len(a) != len(b) {
		return failed(fmt.Sprintf("expected %d elements, got %d", len(a), len(b)))
	}

	if j.ignoreOrder {
		if !j.compareIgnoringOrder(a, b) {
			return failed(fmt.Sprintf("expected %q, got %q", output, actualOutput))
		}
		return accepted()
	}

	for i := range a {
		if r := j.subJudger.Judge(input, a[i], b[i]); !r.IsAccepted() {
			rr := failed(r.GetInfo() + " at index " + strconv.Itoa(i))
			return rr
		}
	}
	return accepted()
}

// TODO improve the detection of "any order"
func shouldIgnoreOrder(q *leetcode.QuestionData) bool {
	content := q.GetEnglishContent()
	content = strip.StripTags(content)
	if strings.Contains(content, "return the answer in any order") {
		return true
	}
	if strings.Contains(content, "return the result in any order") {
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

type floatJudger struct{}

func (floatJudger) Judge(input []string, output, actualOutput string) JudgeResult {
	a, _ := strconv.ParseFloat(output, 64)
	b, _ := strconv.ParseFloat(actualOutput, 64)
	if math.Abs(a-b) >= 1e-5 {
		return failed(fmt.Sprintf("expected %f, got %f", a, b))
	}
	return accepted()
}

type systemDesignJudger struct {
	judgers map[string]Judger
}

func newSystemDesignJudger(q *leetcode.QuestionData) *systemDesignJudger {
	judgers := map[string]Judger{}
	for _, m := range q.MetaData.Methods {
		// TODO: if two functions both return a slice, we can't distinguish them
		//  We just compare both function results ignoring order.
		judgers[m.Name] = getJudger(q, m.Return.Type, true)
	}
	return &systemDesignJudger{judgers}
}

func (s systemDesignJudger) Judge(input []string, output, actualOutput string) JudgeResult {
	// First line of the input is the function names
	var funcs []string
	_ = json.Unmarshal([]byte(input[0]), &funcs)
	inputs, _ := goutils.SplitArray(input[1])
	a, _ := goutils.SplitArray(output)
	b, _ := goutils.SplitArray(actualOutput)

	if len(a) != len(b) || len(a) != len(funcs) {
		panic("system design input/output not match")
	}

	// i == 0 is the constructor, its output is always "null", skip it.
	for i := 1; i < len(a); i++ {
		judger := s.judgers[funcs[i]]
		if judger == nil {
			panic(fmt.Sprintf("judger for %s not found", funcs[i]))
		}
		if r := judger.Judge(input, a[i], b[i]); !r.IsAccepted() {
			param := inputs[i][1 : len(inputs[i])-1] // remove []
			rr := failed(fmt.Sprintf("%s at index %d [%s(%s)]", r.GetInfo(), i, funcs[i], param))
			return rr
		}
	}
	return accepted()
}

func GetJudger(q *leetcode.QuestionData) Judger {
	if q.MetaData.SystemDesign {
		return newSystemDesignJudger(q)
	}
	resultType := q.MetaData.ResultType()
	return getJudger(q, resultType, true)
}

func getJudger(q *leetcode.QuestionData, tp string, topLevel bool) Judger {
	switch tp {
	case "double":
		return floatJudger{}
	case "string":
		return stringJudger{}
	default:
		if strings.HasSuffix(tp, "[]") {
			// Support top-level slice out-of-order comparison only.
			ignoreOrder := topLevel && shouldIgnoreOrder(q)
			subJudger := getJudger(q, tp[:len(tp)-2], false)
			return newSliceJudger(ignoreOrder, subJudger)
		}
		// void, bool, int, long, TreeNode, etc.
		return stringJudger{}
	}
}
