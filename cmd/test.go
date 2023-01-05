package cmd

import (
	"fmt"

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
		cred := leetcode.CredentialsFromConfig()
		c := leetcode.NewClient(leetcode.WithCredentials(cred))
		qs, err := leetcode.ParseQID(args[0], c)
		if err != nil {
			return err
		}
		fmt.Println(qs)
		return nil
	},
}

var (
	mode   string
	submit bool
)

func init() {
	testCmd.Flags().StringVarP(
		&mode,
		"mode",
		"m",
		"auto",
		"test mode, one of: [auto, local, remote]. `auto` mode will try to run test locally, if not supported then submit to Leetcode to test.",
	)
	testCmd.Flags().BoolVarP(&submit, "submit", "s", false, "auto submit if all tests passed")
}
