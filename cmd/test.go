package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
	"github.com/spf13/cobra"
)

var (
	runLocally  bool
	autoSubmit  bool
	customCases []string
)

func init() {
	testCmd.Flags().BoolVarP(
		&runLocally,
		"mode",
		"m",
		false,
		"run test locally",
	)
	testCmd.Flags().StringSliceVarP(&customCases, "cases", "c", nil, "custom test cases")
	testCmd.Flags().BoolVarP(&autoSubmit, "submit", "s", false, "auto submit if all tests passed")
}

var testCmd = &cobra.Command{
	Use:     "test qid",
	Aliases: []string{"t"},
	Args:    cobra.ExactArgs(1),
	Short:   "Run question test cases",
	Example: `leetgo test 244`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		gen := lang.GetGenerator(cfg.Code.Lang)
		cred := leetcode.CredentialsFromConfig()
		c := leetcode.NewClient(leetcode.WithCredentials(cred))
		qs, err := leetcode.ParseQID(args[0], c)
		if err != nil {
			return err
		}

		localTestRunner, supportLocalTest := gen.(lang.LocalTester)
		if runLocally && !supportLocalTest {
			return fmt.Errorf("local test not supported for %s", cfg.Code.Lang)
		}

		for _, q := range qs {
			passed := false
			if runLocally {
				hclog.L().Info("running test locally", "question", q.TitleSlug)
				err = localTestRunner.RunTest(q)
				if err != nil {
					hclog.L().Error("failed to run test locally", "question", q.TitleSlug, "err", err)
				} else {
					passed = true
				}
			} else {
				result, err := runTestRemotely(q, c, gen)
				if err != nil {
					hclog.L().Error("failed to run test remotely", "question", q.TitleSlug, "err", err)
				} else {
					showTestResult(result, q)
					passed = result.CorrectAnswer
				}
			}

			if passed && autoSubmit {
				result, err := submitSolution(q, c, gen)
				if err != nil {
					hclog.L().Error("failed to submit solution", "question", q.TitleSlug, "err", err)
				} else {
					showSubmitResult(result, q)
				}
			}
		}
		return nil
	},
}

func runTestRemotely(q *leetcode.QuestionData, c leetcode.Client, gen lang.Generator) (
	*leetcode.SubmitCheckResult,
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
	cases := q.GetTestCases()
	cases = append(cases, customCases...)
	if len(cases) == 0 {
		return nil, fmt.Errorf("no test cases found")
	}

	casesStr := strings.Join(cases, "\n")
	hclog.L().Info("running remote test", "question", q.TitleSlug)
	interResult, err := c.InterpretSolution(q, gen.Slug(), solution, casesStr)
	if err != nil {
		return nil, fmt.Errorf("failed to interpret solution: %w", err)
	}

	testResult, err := waitResult(c, interResult.InterpretId)
	if err != nil {
		return nil, fmt.Errorf("failed to wait test result: %w", err)
	}
	return testResult, nil
}

func submitSolution(q *leetcode.QuestionData, c leetcode.Client, gen lang.Generator) (
	*leetcode.SubmitCheckResult,
	error,
) {
	solution, err := lang.GetSolutionCode(q)
	if err != nil {
		return nil, fmt.Errorf("failed to get solution code: %w", err)
	}
	hclog.L().Info("submitting solution", "question", q.TitleSlug)
	submissionId, err := c.Submit(q, gen.Slug(), solution)
	if err != nil {
		return nil, fmt.Errorf("failed to submit solution: %w", err)
	}

	testResult, err := waitResult(c, submissionId)
	if err != nil {
		return nil, fmt.Errorf("failed to wait submit result: %w", err)
	}
	return testResult, nil
}

func waitResult(c leetcode.Client, submissionId string) (*leetcode.SubmitCheckResult, error) {
	for {
		result, err := c.CheckSubmissionResult(submissionId)
		if err != nil {
			return nil, err
		}
		if result.State == "SUCCESS" {
			return result, nil
		}
		time.Sleep(1 * time.Second)
	}
}

func showTestResult(result *leetcode.SubmitCheckResult, q *leetcode.QuestionData) {
	if result.CorrectAnswer {
		hclog.L().Info(result.StatusMsg, "question", q.TitleSlug)
	} else {
		hclog.L().Error(result.StatusMsg, "question", q.TitleSlug, "compare", result.CompareResult)
	}
}

func showSubmitResult(result *leetcode.SubmitCheckResult, q *leetcode.QuestionData) {
	if result.State == "SUCCESS" {
		hclog.L().Info("solution submitted", "question", q.TitleSlug, "status", result.State)
	} else {
		hclog.L().Error("failed to submit solution", "question", q.TitleSlug, "status", result.State)
	}
}
