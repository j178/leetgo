package cmd

import (
	"github.com/j178/leetgo/leetcode"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open [qid]",
	Short: "open a question in browser",
	Example: `leetgo open
leetgo open today
leetgo open 549
leetgo open two-sum`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cred := leetcode.CredentialsFromConfig()
		c := leetcode.NewClient(leetcode.WithCredentials(cred))
		var err error
		if len(args) > 0 {
			qid := args[0]
			qs, err := leetcode.ParseQID(qid, c)
			if err != nil {
				return err
			}
			if len(qs) > 1 {
				for _, r := range qs {
					err = browser.OpenURL(r.ContestUrl())
				}
			} else {
				err = browser.OpenURL(qs[0].Url())

			}

		}
		return err
	},
}
