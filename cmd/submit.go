package cmd

import (
	"fmt"

	"github.com/j178/leetgo/leetcode"
	"github.com/spf13/cobra"
)

var submitCmd = &cobra.Command{
	Use:     "submit qid",
	Short:   "Submit solution",
	Aliases: []string{"s"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cred := leetcode.CredentialsFromConfig()
		c := leetcode.NewClient(leetcode.WithCredentials(cred))
		fmt.Println(c.GetUserStatus())
		return nil
	},
}
