package cmd

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
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
			hclog.L().Info("submitting solution", "question", q.TitleSlug, "user", user.Whoami(c))
			result, err := submitSolution(cmd, q, c, gen, limiter)
			if err != nil {
				hclog.L().Error("failed to submit solution", "question", q.TitleSlug, "err", err)
				continue
			}
			cmd.Print(result.Display(qs[0]))
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
