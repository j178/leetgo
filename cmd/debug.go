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
		name, err := whoami(c)
		if err != nil {
			return err
		}
		cmd.Println(name)
		return nil
	},
}

var userCache map[leetcode.Client]string

func init() {
	userCache = make(map[leetcode.Client]string)
}

func whoami(c leetcode.Client) (string, error) {
	if userCache[c] == "" {
		u, err := c.GetUserStatus()
		if err != nil {
			return "", err
		}
		if !u.IsSignedIn {
			return "", leetcode.ErrUserNotSignedIn
		}
		uri, _ := url.Parse(c.BaseURI())
		userCache[c] = u.Username + "@" + uri.Host
	}
	return userCache[c], nil
}
