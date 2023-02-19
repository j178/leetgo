package cmd

import (
	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	"github.com/j178/leetgo/leetcode"
)

var openCmd = &cobra.Command{
	Use:   "open [qid]",
	Short: "open a question in browser",
	Example: `leetgo open
leetgo open today
leetgo open 549
leetgo open two-sum`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cred := leetcode.CredentialsFromConfig()
		c := leetcode.NewClient(leetcode.WithCredentials(cred))
		qid := args[0]
		qs, err := leetcode.ParseQID(qid, c)
		if err != nil {
			return err
		}
		for _, q := range qs {
			if q.IsContest() {
				err = browser.OpenURL(q.ContestUrl())
			} else {
				err = browser.OpenURL(q.Url())
			}
		}

		return err
	},
}
