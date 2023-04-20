package leetcode

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/fatih/color"
)

type UserStatus struct {
	Username        string `json:"username"`
	UserSlug        string `json:"userSlug"`
	RealName        string `json:"realName"`
	Avatar          string `json:"avatar"`
	ActiveSessionId int    `json:"activeSessionId"`
	IsSignedIn      bool   `json:"isSignedIn"`
	IsPremium       bool   `json:"isPremium"`
}

func (u *UserStatus) Whoami(c Client) string {
	uri, _ := url.Parse(c.BaseURI())
	return u.Username + "@" + uri.Host
}

type InterpretSolutionResult struct {
	InterpretExpectedId string `json:"interpret_expected_id"`
	InterpretId         string `json:"interpret_id"`
	TestCase            string `json:"test_case"`
}

type CheckResult interface {
	Display(q *QuestionData) string
	GetState() string
}

type StatusCode int

const (
	Accepted            StatusCode = 10
	WrongAnswer         StatusCode = 11 // submit only
	MemoryLimitExceeded StatusCode = 12 // submit only?
	OutputLimitExceeded StatusCode = 13
	TimeLimitExceeded   StatusCode = 14
	RuntimeError        StatusCode = 15
	CompileError        StatusCode = 20
)

type SubmitCheckResult struct {
	CodeOutput        string  `json:"code_output"` // answers of our code
	CompareResult     string  `json:"compare_result"`
	ElapsedTime       int     `json:"elapsed_time"`
	ExpectedOutput    string  `json:"expected_output"`
	FastSubmit        bool    `json:"fast_submit"`
	Finished          bool    `json:"finished"`
	Lang              string  `json:"lang"`
	LastTestcase      string  `json:"last_testcase"`
	Memory            int     `json:"memory"`
	MemoryPercentile  float64 `json:"memory_percentile"`
	PrettyLang        string  `json:"pretty_lang"`
	QuestionId        string  `json:"question_id"`
	RunSuccess        bool    `json:"run_success"`
	RuntimePercentile float64 `json:"runtime_percentile"`
	State             string  `json:"state"`
	StatusCode        int     `json:"status_code"`
	StatusMemory      string  `json:"status_memory"`
	StatusMsg         string  `json:"status_msg"`
	StatusRuntime     string  `json:"status_runtime"`
	StdOutput         string  `json:"std_output"`
	SubmissionId      string  `json:"submission_id"`
	TaskFinishTime    int     `json:"task_finish_time"`
	TaskName          string  `json:"task_name"`
	TotalCorrect      int     `json:"total_correct"`
	TotalTestcases    int     `json:"total_testcases"`
	CompileError      string  `json:"compile_error"`
	FullCompileError  string  `json:"full_compile_error"`
	FullRuntimeError  string  `json:"full_runtime_error"`
}

var (
	// TODO replace with lipgloss
	colorGreen  = color.New(color.FgHiGreen, color.Bold)
	colorYellow = color.New(color.FgHiYellow, color.Bold)
	colorFaint  = color.New(color.Faint)
	colorRed    = color.New(color.FgHiRed, color.Bold)
)

func (r *SubmitCheckResult) Display(q *QuestionData) string {
	stdout := ""
	if len(r.CodeOutput) > 1 {
		stdout = "\nStdout:        " + strings.ReplaceAll(r.StdOutput, "\n", "↩ ")
	}
	switch StatusCode(r.StatusCode) {
	case Accepted:
		return fmt.Sprintf(
			"\n%s\n%s%s%s\n",
			colorGreen.Sprintf("√ %s", r.StatusMsg),
			fmt.Sprintf("\nPassed cases:  %d/%d", r.TotalCorrect, r.TotalTestcases),
			fmt.Sprintf("\nRuntime:       %s, better than %.0f%%", r.StatusRuntime, r.RuntimePercentile),
			fmt.Sprintf("\nMemory:        %s, better than %.0f%%", r.StatusMemory, r.RuntimePercentile),
		)
	case WrongAnswer:
		return fmt.Sprintf(
			"\n%s\n%s%s%s%s%s\n",
			colorRed.Sprint(" × Wrong Answer"),
			fmt.Sprintf("\nPassed cases:  %d/%d", r.TotalCorrect, r.TotalTestcases),
			fmt.Sprintf("\nLast case:     %s", strings.ReplaceAll(r.LastTestcase, "\n", "↩ ")),
			fmt.Sprintf("\nOutput:        %s", strings.ReplaceAll(r.CodeOutput, "\n", "↩ ")),
			stdout,
			fmt.Sprintf("\nExpected:      %s", strings.ReplaceAll(r.ExpectedOutput, "\n", "↩ ")),
		)
	case MemoryLimitExceeded, TimeLimitExceeded, OutputLimitExceeded:
		return fmt.Sprintf(
			"\n%s\n%s%s\n",
			colorYellow.Sprintf("\n × %s\n", r.StatusMsg),
			fmt.Sprintf("\nPassed cases:  %d/%d", r.TotalCorrect, r.TotalTestcases),
			fmt.Sprintf("\nLast case:     %s", r.LastTestcase),
		)
	case RuntimeError:
		return fmt.Sprintf(
			"\n%s\n%s\n\n%s\n",
			colorRed.Sprintf(" × %s", r.StatusMsg),
			fmt.Sprintf("Passed cases:   %s", formatCompare(r.CompareResult)),
			colorFaint.Sprint(r.FullRuntimeError),
		)
	case CompileError:
		return fmt.Sprintf(
			"\n%s\n\n%s\n",
			colorRed.Sprintf(" × %s", r.StatusMsg),
			colorFaint.Sprint(r.FullCompileError),
		)
	default:
		return fmt.Sprintf("\n%s\n", colorRed.Sprintf(" × %s", r.StatusMsg))
	}
}

func (r *SubmitCheckResult) GetState() string {
	return r.State
}

func (r *SubmitCheckResult) Accepted() bool {
	return r.StatusCode == int(Accepted)
}

type RunCheckResult struct {
	InputData              string
	State                  string   `json:"state"` // STARTED, SUCCESS
	StatusCode             int      `json:"status_code"`
	StatusMsg              string   `json:"status_msg"`         // Accepted, Wrong Answer, Time Limit Exceeded, Memory Limit Exceeded, Runtime Error, Compile Error, Output Limit Exceeded, Unknown Error
	Memory                 int      `json:"memory"`             // 内存消耗 in bytes
	StatusMemory           string   `json:"status_memory"`      // 内存消耗
	MemoryPercentile       float64  `json:"memory_percentile"`  // 内存消耗击败百分比
	StatusRuntime          string   `json:"status_runtime"`     // 执行用时
	RuntimePercentile      float64  `json:"runtime_percentile"` // 用时击败百分比
	Lang                   string   `json:"lang"`
	PrettyLang             string   `json:"pretty_lang"`
	CodeAnswer             []string `json:"code_answer"`   // return values of our code
	CompileError           string   `json:"compile_error"` //
	FullCompileError       string   `json:"full_compile_error"`
	FullRuntimeError       string   `json:"full_runtime_error"`
	CompareResult          string   `json:"compare_result"`  // "111", 1 means correct, 0 means wrong
	CorrectAnswer          bool     `json:"correct_answer"`  // true means all passed
	CodeOutput             []string `json:"code_output"`     // output to stdout of our code
	StdOutputList          []string `json:"std_output_list"` // list of output to stdout, same as code_output
	TaskName               string   `json:"task_name"`
	TotalCorrect           int      `json:"total_correct"`   // number of correct answers
	TotalTestcases         int      `json:"total_testcases"` // number of test cases
	ElapsedTime            int      `json:"elapsed_time"`
	TaskFinishTime         int      `json:"task_finish_time"`
	RunSuccess             bool     `json:"run_success"` // true if run success
	FastSubmit             bool     `json:"fast_submit"`
	Finished               bool     `json:"finished"`
	ExpectedOutput         string   `json:"expected_output"`
	ExpectedCodeAnswer     []string `json:"expected_code_answer"`
	ExpectedCodeOutput     []string `json:"expected_code_output"`
	ExpectedElapsedTime    int      `json:"expected_elapsed_time"`
	ExpectedLang           string   `json:"expected_lang"`
	ExpectedMemory         int      `json:"expected_memory"`
	ExpectedRunSuccess     bool     `json:"expected_run_success"`
	ExpectedStatusCode     int      `json:"expected_status_code"`
	ExpectedStatusRuntime  string   `json:"expected_status_runtime"`
	ExpectedStdOutputList  []string `json:"expected_std_output_list"`
	ExpectedTaskFinishTime int      `json:"expected_task_finish_time"`
	ExpectedTaskName       string   `json:"expected_task_name"`
}

func formatCompare(s string) string {
	var sb strings.Builder
	for _, c := range s {
		if c == '1' {
			sb.WriteString(colorGreen.Sprint("√"))
		} else {
			sb.WriteString(colorRed.Sprint("×"))
		}
	}
	return sb.String()
}

func (r *RunCheckResult) Display(q *QuestionData) string {
	stdout := ""
	if len(r.CodeOutput) > 1 {
		stdout = "\nStdout:        " + strings.Join(r.CodeOutput, "↩ ")
	}
	switch StatusCode(r.StatusCode) {
	case Accepted:
		if r.CorrectAnswer {
			return fmt.Sprintf(
				"\n%s\n%s%s%s%s%s\n",
				colorGreen.Sprintf("√ %s", r.StatusMsg),
				fmt.Sprintf("\nPassed cases:  %s", formatCompare(r.CompareResult)),
				fmt.Sprintf("\nInput:         %s", strings.ReplaceAll(r.InputData, "\n", "↩ ")),
				fmt.Sprintf("\nOutput:        %s", strings.Join(r.CodeAnswer, "↩ ")),
				stdout,
				fmt.Sprintf("\nExpected:      %s", strings.Join(r.ExpectedCodeAnswer, "↩ ")),
			)
		} else {
			return fmt.Sprintf(
				"\n%s\n%s%s%s%s%s\n",
				colorRed.Sprint(" × Wrong Answer"),
				fmt.Sprintf("\nPassed cases:  %s", formatCompare(r.CompareResult)),
				fmt.Sprintf("\nInput:         %s", strings.ReplaceAll(r.InputData, "\n", "↩ ")),
				fmt.Sprintf("\nOutput:        %s", strings.Join(r.CodeAnswer, "↩ ")),
				stdout,
				fmt.Sprintf("\nExpected:      %s", strings.Join(r.ExpectedCodeAnswer, "↩ ")),
			)
		}
	case MemoryLimitExceeded, TimeLimitExceeded, OutputLimitExceeded:
		return colorYellow.Sprintf("\n × %s\n", r.StatusMsg)
	case RuntimeError:
		return fmt.Sprintf(
			"\n%s\n%s\n\n%s\n",
			colorRed.Sprintf(" × %s", r.StatusMsg),
			fmt.Sprintf("Passed cases:   %s", formatCompare(r.CompareResult)),
			colorFaint.Sprint(r.FullRuntimeError),
		)
	case CompileError:
		return fmt.Sprintf(
			"\n%s\n\n%s\n",
			colorRed.Sprintf(" × %s", r.StatusMsg),
			colorFaint.Sprint(r.FullCompileError),
		)
	default:
		return fmt.Sprintf("\n%s\n", colorRed.Sprintf(" × %s", r.StatusMsg))
	}
}

func (r *RunCheckResult) GetState() string {
	return r.State
}

type QuestionList struct {
	Questions []*QuestionData `json:"questions"`
	HasMore   bool            `json:"hasMore"`
	Total     int             `json:"total"`
}

type QuestionTag struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	NameTranslated string `json:"nameTranslated"`
	Slug           string `json:"slug"`
	TypeName       string `json:"typeName"`
	TypeTransName  string `json:"typeTransName"`
}
