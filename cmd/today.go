package cmd

import (
	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
	"github.com/spf13/cobra"
)

var todayCmd = &cobra.Command{
	Use:     "today",
	Short:   "Generate the question of today",
	Example: `leetgo today`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := leetcode.NewClient()
		q, err := c.GetTodayQuestion()
		if err != nil {
			return err
		}
		_, err = lang.Generate(q)
		return err
	},
}
