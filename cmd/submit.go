package cmd

import (
	"errors"
	"fmt"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
	"github.com/spf13/cobra"
)

var submitCmd = &cobra.Command{
	Use:     "submit qid",
	Short:   "Submit solution",
	Aliases: []string{"s"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		gen := lang.GetGenerator(cfg.Code.Lang)
		cred := leetcode.CredentialsFromConfig()
		c := leetcode.NewClient(leetcode.WithCredentials(cred))
		qs, err := leetcode.ParseQID(args[0], c)
		if err != nil {
			return err
		}
		if len(qs) > 1 {
			return errors.New("multiple questions found")
		}
		result, err := submitSolution(qs[0], c, gen)
		if err != nil {
			return fmt.Errorf("failed to submit solution: %w", err)
		}
		showSubmitResult(result, qs[0])
		return nil
	},
}
