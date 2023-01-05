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

var testCmd = &cobra.Command{
	Use:     "test qid",
	Aliases: []string{"t"},
	Args:    cobra.ExactArgs(1),
	Short:   "Run question test cases",
	Example: `leetgo test 244`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if mode != "auto" && mode != "local" && mode != "remote" {
			return fmt.Errorf("invalid test mode: %s", mode)
		}
		cfg := config.Get()
		cred := leetcode.CredentialsFromConfig()
		c := leetcode.NewClient(leetcode.WithCredentials(cred))
		qs, err := leetcode.ParseQID(args[0], c)
		if err != nil {
			return err
		}

		gen := lang.GetGenerator(cfg.Code.Lang)
		for _, q := range qs {
			if mode == "auto" || mode == "local" {
				if gen, ok := gen.(lang.LocalTester); ok {
					hclog.L().Info("running local test")
					return gen.RunTest(q)
				}
			}
			if mode == "local" {
				return nil
			}

			solution, err := lang.GetSolutionCode(q)
			if err != nil {
				hclog.L().Error("failed to get solution code", "question", q.TitleSlug, "err", err)
				continue
			}
			err = q.Fulfill()
			if err != nil {
				hclog.L().Error("failed to fetch question", "question", q.TitleSlug, "err", err)
				continue
			}
			cases := q.GetTestCases()
			cases = append(cases, customCases...)
			if len(cases) == 0 {
				hclog.L().Warn("no test cases found", "question", q.TitleSlug)
				continue
			}
			casesStr := strings.Join(cases, "\n")
			hclog.L().Info("running remote test")
			interResult, err := c.InterpretSolution(q, gen.Slug(), solution, casesStr)
			if err != nil {
				hclog.L().Error("failed to interpret solution", "question", q.TitleSlug, "err", err)
				continue
			}
			testResult, err := waitResult(c, interResult.InterpretId)
			if err != nil {
				hclog.L().Error("failed to wait test result", "question", q.TitleSlug, "err", err)
				continue
			}
			fmt.Printf("%+v", testResult)
		}
		return nil
	},
}

func waitResult(c leetcode.Client, submissionId string) (*leetcode.SubmissionCheckResult, error) {
	for {
		result, err := c.CheckSubmissionResult(submissionId)
		if err != nil {
			hclog.L().Error("failed to get submission result", "submissionId", submissionId, "err", err)
			return nil, err
		}
		fmt.Println(result)
		if result.State == "SUCCESS" {
			return result, nil
		}
		time.Sleep(2 * time.Second)
	}
}

var (
	mode        string
	submit      bool
	customCases []string
)

func init() {
	testCmd.Flags().StringVarP(
		&mode,
		"mode",
		"m",
		"auto",
		"test mode, one of: [auto, local, remote]. `auto` mode will try to run test locally, if not supported then submit to Leetcode to test.",
	)
	testCmd.Flags().StringSliceVarP(&customCases, "cases", "c", nil, "custom test customCases")
	testCmd.Flags().BoolVarP(&submit, "submit", "s", false, "auto submit if all tests passed")
}
