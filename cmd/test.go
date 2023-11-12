package cmd

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/utils"
)

var (
	runLocally  bool
	runRemotely bool = true
	runBoth     bool
	autoSubmit  bool
	targetCase  string
	forceSubmit bool
)

func init() {
	testCmd.Flags().BoolVarP(
		&runLocally,
		"local",
		"L",
		false,
		"run test locally",
	)
	testCmd.Flags().BoolVarP(
		&runBoth,
		"both",
		"B",
		false,
		"run test both locally and remotely",
	)
	testCmd.Flags().BoolVarP(&autoSubmit, "submit", "s", false, "auto submit if all tests passed")
	testCmd.Flags().BoolVarP(&forceSubmit, "force", "f", false, "force submit even if local test failed")
	testCmd.Flags().StringVarP(&targetCase, "target", "t", "-", "only run the specified test case, e.g. 1, 1-3, -1, 1-")
}

var testCmd = &cobra.Command{
	Use:       "test qid",
	Aliases:   []string{"t"},
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"today", "last", "last/"},
	Short:     "Run question test cases",
	Example: `leetgo test 244
leetgo test last
leetgo test w330/1
leetgo test w330/`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if runLocally {
			runRemotely = false
		}
		if runBoth {
			runLocally = true
			runRemotely = true
		}

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
		_, supportLocalTest := gen.(lang.LocalTestable)
		if runLocally && !supportLocalTest {
			return fmt.Errorf("local test not supported for %s", cfg.Code.Lang)
		}

		user, err := c.GetUserStatus()
		if err != nil {
			user = &leetcode.UserStatus{}
		}
		testLimiter := newLimiter(user)
		submitLimiter := newLimiter(user)

		var hasFailedCase bool
		for _, q := range qs {
			var (
				localPassed    = true
				remotePassed   = true
				submitAccepted = true
			)
			if runLocally {
				log.Info("running test locally", "question", q.TitleSlug)
				localPassed, err = lang.RunLocalTest(q, targetCase)
				if err != nil {
					log.Error("failed to run test locally", "err", err)
				}
			}
			if runRemotely {
				log.Info("running test remotely", "question", q.TitleSlug)
				result, err := runTestRemotely(cmd, q, c, gen, testLimiter)
				if err != nil {
					log.Error("failed to run test remotely", "err", err)
					remotePassed = false
				} else {
					cmd.Print(result.Display(q))
					remotePassed = result.CorrectAnswer
				}
			}

			if autoSubmit && remotePassed && (localPassed || forceSubmit) {
				log.Info("submitting solution", "user", user.Whoami(c))
				result, err := submitSolution(cmd, q, c, gen, submitLimiter)
				if err != nil {
					submitAccepted = false
					log.Error("failed to submit solution", "err", err)
				} else {
					cmd.Print(result.Display(q))
					if !result.Accepted() {
						submitAccepted = false
						added, _ := appendToTestCases(q, result)
						if added {
							log.Info("added failed case to testcases.txt")
						}
					}
				}
			}

			if !localPassed || !remotePassed || !submitAccepted {
				hasFailedCase = true
			}
		}

		if hasFailedCase {
			return exitCode(1)
		}
		return nil
	},
}

func runTestRemotely(
	cmd *cobra.Command,
	q *leetcode.QuestionData,
	c leetcode.Client,
	gen lang.Lang,
	limiter *utils.RateLimiter,
) (
	*leetcode.RunCheckResult,
	error,
) {
	solution, err := lang.GetSolutionCode(q)
	if err != nil {
		return nil, fmt.Errorf("failed to get solution code: %w", err)
	}
	err = q.Fulfill()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch question: %s", err)
	}

	exampleCases := q.GetExampleTestCases()
	casesStr := strings.Join(exampleCases, "\n")

	var (
		cases             lang.TestCases
		fromTestCasesFile bool
	)
	testCasesFile, err := lang.GetFileOutput(q, lang.TestCasesFile)
	if err == nil {
		cases, err = lang.ParseTestCases(q, testCasesFile)
		if err == nil && len(cases.Cases) > 0 {
			fromTestCasesFile = true
			casesStr = cases.InputString()
		}
	}

	spin := newSpinner(cmd.ErrOrStderr())
	spin.Suffix = " Running tests..."
	spin.Reverse()
	spin.Start()
	defer spin.Stop()

	limiter.Take()
	spin.Reverse()

	interResult, err := c.RunCode(q, gen.Slug(), solution, casesStr)
	if err != nil {
		return nil, fmt.Errorf("failed to run test: %w", err)
	}

	spin.Lock()
	spin.Suffix = " Waiting for result..."
	spin.Unlock()

	testResult, err := waitResult(c, interResult.InterpretId)
	if err != nil {
		return nil, fmt.Errorf("failed to wait test result: %w", err)
	}
	r := testResult.(*leetcode.RunCheckResult)
	r.InputData = interResult.TestCase

	if r.Accepted() && fromTestCasesFile {
		updated, err := cases.UpdateOutputs(r.ExpectedCodeAnswer)
		if err != nil {
			log.Debug("failed to update test cases", "err", err)
		} else if updated {
			content := []byte(cases.String())
			err = utils.WriteFile(testCasesFile.GetPath(), content)
			if err != nil {
				log.Debug("failed to update test cases", "err", err)
			} else {
				log.Info("testcases.txt updated")
			}
		}
	}

	return r, nil
}

func waitResult(c leetcode.Client, submissionId string) (
	leetcode.CheckResult,
	error,
) {
	for {
		result, err := c.CheckResult(submissionId)
		if err != nil {
			return nil, err
		}
		if result.GetState() == "SUCCESS" {
			return result, nil
		}
		time.Sleep(1 * time.Second)
	}
}

func newLimiter(user *leetcode.UserStatus) *utils.RateLimiter {
	if user.IsPremium {
		return utils.NewRateLimiter(1 * time.Second)
	}
	return utils.NewRateLimiter(10 * time.Second)
}

func newSpinner(w io.Writer) *spinner.Spinner {
	spin := spinner.New(
		spinner.CharSets[11],
		125*time.Millisecond,
		spinner.WithHiddenCursor(false),
		spinner.WithWriter(w),
		spinner.WithColor("fgHiCyan"),
	)
	return spin
}
