package cmd

import (
	"fmt"

	"github.com/j178/leetgo/leetcode"
	"github.com/spf13/cobra"
)

var contestCmd = &cobra.Command{
	Use:     "contest [qid]",
	Short:   "Generate contest questions",
	Aliases: []string{"c"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var contestSlug string
		if len(args) == 0 {
			// get upcoming contest
			// select to register / unregister
			// register then wait for contest to start
			contestSlug = "weekly-contest-327"
		} else {
			contestSlug = args[0]
		}
		cred := leetcode.CredentialsFromConfig()
		c := leetcode.NewClient(leetcode.WithCredentials(cred))
		contest, err := c.GetContest(contestSlug)
		if err != nil {
			return err
		}
		_, _ = contest.GetAllQuestions()
		fmt.Printf("%+v", contest)

		return nil
	},
}
