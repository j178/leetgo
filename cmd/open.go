package cmd

import (
	"github.com/cli/browser"
	"github.com/spf13/cobra"

	"github.com/j178/leetgo/leetcode"
)

var openCmd = &cobra.Command{
	Use:   "open qid",
	Short: "Open one or multiple question pages in a browser",
	Example: `leetgo open today
leetgo open 549
leetgo open w330/`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"today", "last"},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := leetcode.NewClient(leetcode.ReadCredentials())
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
