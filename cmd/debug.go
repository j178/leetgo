package cmd

import (
	"bytes"

	"github.com/goccy/go-json"
	"github.com/spf13/cobra"

	"github.com/j178/leetgo/leetcode"
)

var inspectCmd = &cobra.Command{
	Use:    "inspect",
	Short:  "Inspect LeetCode API, developer only",
	Args:   cobra.ExactArgs(1),
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := leetcode.NewClient()
		resp, err := c.Inspect(args[0])
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetIndent("", "  ")
		_ = enc.Encode(resp)
		cmd.Print(buf.String())
		return nil
	},
}

var whoamiCmd = &cobra.Command{
	Use:    "whoami",
	Short:  "Print the current user",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cred := leetcode.CredentialsFromConfig()
		c := leetcode.NewClient(leetcode.WithCredentials(cred))
		user, err := c.GetUserStatus()
		if err != nil {
			return err
		}
		cmd.Println(user.Whoami(c))
		return nil
	},
}
