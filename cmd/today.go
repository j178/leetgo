package cmd

import (
	"fmt"

	"github.com/j178/leetgo/leetcode"
	"github.com/spf13/cobra"
)

var todayCmd = &cobra.Command{
	Use:     "today",
	Short:   "Generate the question of today",
	Example: `leetgo today --go`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := leetcode.NewClient()
		q, err := c.GetTodayQuestion()
		if err != nil {
			return err
		}
		fmt.Printf("%v\n", q.TitleSlug)
		return nil
	},
}

func init() {
	addLangFlags(todayCmd)
}
