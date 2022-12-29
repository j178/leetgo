package cmd

import (
	"fmt"

	"github.com/j178/leetgo/leetcode"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:     "new SLUG_OR_ID...",
	Short:   "Generate a new question",
	Example: "leetgo new 450 --go\nleetgo new two-sum --go",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := leetcode.NewClient()
		for _, p := range args {
			q, _ := leetcode.Question(p, c)
			fmt.Println(q)
		}
		return nil
	},
}

func init() {
	addLangFlags(newCmd)
}
