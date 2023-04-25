package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

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
	Aliases: []string{"s"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		cred := leetcode.CredentialsFromConfig()
		c := leetcode.NewClient(leetcode.WithCredentials(cred))
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

		for _, q := range qs {
			log.Info("submitting solution", "question", q.TitleSlug, "user", user.Whoami(c))
			result, err := submitSolution(cmd, q, c, gen, limiter)
			if err != nil {
				log.Error("failed to submit solution", "question", q.TitleSlug, "err", err)
				continue
			}
			cmd.Print(result.Display(qs[0]))

			if !result.Accepted() {
				added, _ := appendToTestCases(q, result)
				if added {
					log.Info("added failed case to testcases.txt", "question", q.TitleSlug)
				}
			}
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
	solution, err := lang.GetSolutionCode(q)
	if err != nil {
		return nil, fmt.Errorf("failed to get solution code: %w", err)
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

	testResult, err := waitResult(c, submissionId)
	if err != nil {
		return nil, fmt.Errorf("failed to wait submit result: %w", err)
	}
	return testResult.(*leetcode.SubmitCheckResult), nil
}

func appendToTestCases(q *leetcode.QuestionData, result *leetcode.SubmitCheckResult) (bool, error) {
	genResult, err := lang.GeneratePathsOnly(q)
	if err != nil {
		return false, err
	}
	testCasesFile := genResult.GetFile(lang.TestCasesFile)
	if !utils.IsExist(testCasesFile.GetPath()) {
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
