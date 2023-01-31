package cmd

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/j178/leetgo/leetcode"
)

var whoamiCmd = &cobra.Command{
	Use:    "whoami",
	Short:  "Print the current user",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cred := leetcode.CredentialsFromConfig()
		c := leetcode.NewClient(leetcode.WithCredentials(cred))
		u, err := c.GetUserStatus()
		if err != nil {
			return err
		}
		url, _ := url.Parse(c.BaseURI())
		cmd.Println(u.Username + "@" + url.Host)
		return nil
	},
}
