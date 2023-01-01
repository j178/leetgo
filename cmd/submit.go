package cmd

import (
	"fmt"

	"github.com/j178/leetgo/leetcode"
	"github.com/spf13/cobra"
)

var submitCmd = &cobra.Command{
	Use:   "submit SLUG_OR_ID",
	Short: "Submit solution",
	RunE: func(cmd *cobra.Command, args []string) error {
		cred, err := leetcode.CredentialsFromConfig()
		if err != nil {
			return err
		}
		c := leetcode.NewClient(leetcode.WithCredentials(cred))
		fmt.Println(c.GetUserStatus())
		return nil
	},
}
