package cmd

import (
	"github.com/spf13/cobra"

	"github.com/j178/leetgo/leetcode"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage local questions cache",
}

var cacheUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update local questions cache",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := leetcode.NewClient(leetcode.ReadCredentials())
		return leetcode.GetCache(c).Update()
	},
}

func init() {
	cacheCmd.AddCommand(cacheUpdateCmd)
}
