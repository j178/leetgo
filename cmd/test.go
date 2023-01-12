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
		"local",
		"L",
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

		limitC := make(chan struct{})
		defer func() { close(limitC) }()
		go func() {
			for {
				if _, ok := <-limitC; !ok {
					return
				}
				time.Sleep(2 * time.Second)
			}
		}()

		for _, q := range qs {
			passed := false
			if runLocally {
				hclog.L().Info("running test locally", "question", q.TitleSlug)
				err = lang.RunLocalTest(q)
				if err != nil {
					hclog.L().Error("failed to run test locally", "question", q.TitleSlug, "err", err)
				} else {
					passed = true
				}
			} else {
				result, err := runTestRemotely(q, c, gen, limitC)
				if err != nil {
					hclog.L().Error("failed to run test remotely", "question", q.TitleSlug, "err", err)
				} else {
					cmd.Print(result.Display(q))
					passed = result.CorrectAnswer
				}
			}

			if passed && autoSubmit {
				result, err := submitSolution(q, c, gen, limitC)
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

func runTestRemotely(q *leetcode.QuestionData, c leetcode.Client, gen lang.Lang, wait chan<- struct{}) (
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
	spin.Start()
	defer spin.Stop()

	wait <- struct{}{}

	interResult, err := c.Test(q, gen.Slug(), solution, casesStr)
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
