package cmd

import (
	"github.com/j178/leetgo/leetcode"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update local questions cache",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := leetcode.NewClient()
		return leetcode.GetCache().Update(c)
	},
}
