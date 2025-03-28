package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/dop251/goja"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/utils"
)

var submitCmd = &cobra.Command{
	Use:   "submit qid",
	Short: "Submit solution",
	Example: `leetgo submit 1
leetgo submit two-sum
leetgo submit last
leetgo submit w330/1
leetgo submit w330/
`,
	Aliases:   []string{"s"},
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"today", "last", "last/"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		c := leetcode.NewClient(leetcode.ReadCredentials())
		qs, err := leetcode.ParseQID(args[0], c)
		if err != nil {
			return err
		}
		gen, err := lang.GetGenerator(cfg.Code.Lang)
		if err != nil {
			return err
		}

		user, err := c.GetUserStatus()
		if err != nil {
			return err
		}
		limiter := newLimiter(user)

		var hasFailedCase bool
		for _, q := range qs {
			log.Info("submitting solution", "question", q.TitleSlug, "user", user.Whoami(c))
			result, err := submitSolution(cmd, q, c, gen, limiter)
			if err != nil {
				hasFailedCase = true
				log.Error("failed to submit solution", "err", err)
				continue
			}
			cmd.Print(result.Display(qs[0]))

			if !result.Accepted() {
				hasFailedCase = true
				added, _ := appendToTestCases(q, result)
				if added {
					log.Info("added failed case to testcases.txt")
				}
			}
		}

		err = showTodayStreak(c, cmd)
		if err != nil {
			log.Debug("failed to show today's streak", "err", err)
		}

		if hasFailedCase {
			return exitCode(1)
		}

		return nil
	},
}

func submitSolution(
	cmd *cobra.Command,
	q *leetcode.QuestionData,
	c leetcode.Client,
	gen lang.Lang,
	limiter *utils.RateLimiter,
) (
	*leetcode.SubmitCheckResult,
	error,
) {
	modifiers, err := getPostModifiers(gen)
	if err != nil {
		return nil, fmt.Errorf("failed to get post modifiers: %w", err)
	}

	solution, err := lang.GetSolutionCode(q)
	if err != nil {
		return nil, fmt.Errorf("failed to get solution code: %w", err)
	}
	for _, m := range modifiers {
		solution = m(solution)
	}

	spin := newSpinner(cmd.ErrOrStderr())
	spin.Suffix = " Submitting solution..."
	spin.Reverse()
	spin.Start()
	defer spin.Stop()

	limiter.Take()
	spin.Reverse()

	submissionId, err := c.SubmitCode(q, gen.Slug(), solution)
	if err != nil {
		return nil, fmt.Errorf("failed to submit solution: %w", err)
	}

	spin.Lock()
	spin.Suffix = " Waiting for result..."
	spin.Unlock()

	testResult, err := waitResult(c, submissionId)
	if err != nil {
		return nil, fmt.Errorf("failed to wait submit result: %w", err)
	}
	return testResult.(*leetcode.SubmitCheckResult), nil
}

func getPostModifiers(lang lang.Lang) ([]func(string) string, error) {
	modifiers := viper.Get("code." + lang.Slug() + ".post-modifiers")
	if modifiers == nil || len(modifiers.([]any)) == 0 {
		modifiers = viper.Get("code." + lang.ShortName() + ".post-modifiers")
	}
	if modifiers == nil || len(modifiers.([]any)) == 0 {
		modifiers = viper.Get("code.post-modifiers")
	}
	if modifiers == nil {
		return nil, nil
	}

	var funcs []func(string) string
	for _, m := range modifiers.([]any) {
		m := m.(map[string]any)
		name, script := "", ""

		if m["script"] != nil {
			script = m["script"].(string)
			vm := goja.New()
			_, err := vm.RunString(script)
			if err != nil {
				return nil, fmt.Errorf("failed to run script: %w", err)
			}
			var jsFn func(string) string
			if vm.Get("modify") == nil {
				return nil, fmt.Errorf("failed to get modify function")
			}
			err = vm.ExportTo(vm.Get("modify"), &jsFn)
			if err != nil {
				return nil, fmt.Errorf("failed to export function: %w", err)
			}
			f := func(s string) string {
				return jsFn(s)
			}
			funcs = append(funcs, f)
			continue
		}
		log.Warn("invalid modifier, ignored", "name", name, "script", script)
	}
	return funcs, nil
}

func appendToTestCases(q *leetcode.QuestionData, result *leetcode.SubmitCheckResult) (bool, error) {
	genResult, err := lang.GeneratePathsOnly(q)
	if err != nil {
		return false, err
	}
	testCasesFile := genResult.GetFile(lang.TestCasesFile)
	if testCasesFile == nil || !utils.IsExist(testCasesFile.GetPath()) {
		return false, nil
	}

	failedCase := lang.TestCase{
		Question: q,
		Input:    strings.Split(result.LastTestcase, "\n"),
		Output:   result.ExpectedOutput,
	}
	// some test cases are hidden during contest, they can be excluded by checking
	err = failedCase.Check()
	if err != nil {
		return false, err
	}

	tc, err := lang.ParseTestCases(q, testCasesFile)
	if err != nil {
		return false, err
	}
	if tc.Contains(failedCase) {
		return false, nil
	}
	tc.AddCase(failedCase)

	content := []byte(tc.String())
	err = utils.WriteFile(testCasesFile.GetPath(), content)
	return true, err
}

func showTodayStreak(c leetcode.Client, cmd *cobra.Command) error {
	streak, err := c.GetStreakCounter()
	if err != nil {
		return err
	}
	today := ""
	if streak.TodayCompleted {
		today = config.PassedStyle.Render("+1")
	}
	cmd.Printf("\nTotal streak:  %d%s\n", streak.StreakCount, today)
	return nil
}
