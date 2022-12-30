package cmd

import (
	"github.com/j178/leetgo/leetcode"
	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage local questions cache",
}

var cacheUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update local questions cache",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := leetcode.NewClient()
		return leetcode.GetCache().Update(c)
	},
}

func init() {
	cacheCmd.AddCommand(cacheUpdateCmd)
}
