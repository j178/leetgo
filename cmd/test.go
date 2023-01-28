package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/utils"
	"github.com/spf13/cobra"
)

var (
	runLocally  bool
	runRemotely bool = true
	runBoth     bool
	autoSubmit  bool
	customCases []string
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
	testCmd.Flags().StringSliceVarP(&customCases, "cases", "c", nil, "custom test cases")
	testCmd.Flags().BoolVarP(&autoSubmit, "submit", "s", false, "auto submit if all tests passed")
}

var testCmd = &cobra.Command{
	Use:     "test qid",
	Aliases: []string{"t"},
	Args:    cobra.ExactArgs(1),
	Short:   "Run question test cases",
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
		cred := leetcode.CredentialsFromConfig()
		c := leetcode.NewClient(leetcode.WithCredentials(cred))
		qs, err := leetcode.ParseQID(args[0], c)
		if err != nil {
			return err
		}

		gen := lang.GetGenerator(cfg.Code.Lang)
		if gen == nil {
			return fmt.Errorf("language %s is not supported yet", cfg.Code.Lang)
		}
		_, supportLocalTest := gen.(lang.LocalTestable)
		if runLocally && !supportLocalTest {
			return fmt.Errorf("local test not supported for %s", cfg.Code.Lang)
		}

		testLimiter := utils.NewRateLimiter(10 * time.Second)
		submitLimiter := utils.NewRateLimiter(10 * time.Second)
		defer testLimiter.Stop()
		defer submitLimiter.Stop()

		for _, q := range qs {
			localPassed, remotePassed := true, true
			if runLocally {
				hclog.L().Info("running test locally", "question", q.TitleSlug)
				localPassed, err = lang.RunLocalTest(q)
				if err != nil {
					hclog.L().Error("failed to run test locally", "question", q.TitleSlug, "err", err)
				}
			}
			if runRemotely {
				result, err := runTestRemotely(q, c, gen, testLimiter)
				if err != nil {
					hclog.L().Error("failed to run test remotely", "question", q.TitleSlug, "err", err)
					remotePassed = false
				} else {
					cmd.Print(result.Display(q))
					remotePassed = result.CorrectAnswer
				}
			}

			if localPassed && remotePassed && autoSubmit {
				result, err := submitSolution(q, c, gen, submitLimiter)
				if err != nil {
					hclog.L().Error("failed to submit solution", "question", q.TitleSlug, "err", err)
				} else {
					cmd.Print(result.Display(q))
				}
			}
		}
		return nil
	},
}

func runTestRemotely(q *leetcode.QuestionData, c leetcode.Client, gen lang.Lang, limiter *utils.RateLimiter) (
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
	cases := q.GetTestCases()
	cases = append(cases, getCustomCases()...)
	if len(cases) == 0 {
		return nil, fmt.Errorf("no test cases found")
	}
	casesStr := strings.Join(cases, "\n")

	hclog.L().Info("running test remotely", "question", q.TitleSlug)
	spin := spinner.New(spinner.CharSets[9], 250*time.Millisecond, spinner.WithSuffix(" Running test..."))
	spin.Reverse()
	spin.Start()
	defer spin.Stop()

	limiter.Wait()
	spin.Reverse()

	interResult, err := c.RunCode(q, gen.Slug(), solution, casesStr)
	if err != nil {
		return nil, fmt.Errorf("failed to run test: %w", err)
	}

	testResult, err := waitResult(c, interResult.InterpretId)
	if err != nil {
		return nil, fmt.Errorf("failed to wait test result: %w", err)
	}
	r := testResult.(*leetcode.RunCheckResult)
	r.InputData = interResult.TestCase
	return r, nil
}

func getCustomCases() []string {
	cases := make([]string, len(customCases))
	for i, c := range customCases {
		cases[i] = strings.ReplaceAll(c, `\n`, "\n")
	}
	return cases
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
		time.Sleep(500 * time.Millisecond)
	}
}
