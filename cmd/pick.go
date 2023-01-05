package cmd

import (
	"fmt"

	"github.com/j178/leetgo/editor"
	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
	"github.com/spf13/cobra"
)

var pickCmd = &cobra.Command{
	Use:   "pick [qid]",
	Short: "Generate a new question",
	Example: `leetgo pick  # show a list of questions to pick
leetgo pick today
leetgo pick 549`,
	Args:    cobra.MaximumNArgs(1),
	Aliases: []string{"p"},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := leetcode.NewClient()
		if len(args) == 0 {
			// 	TODO Start tea TUI to pick a question
			//   looks like https://leetcode.cn/problemset/all/
			args = append(args, "two-sum")
		}
		qid := args[0]
		qs, err := leetcode.ParseQID(qid, c)
		if err != nil {
			return err
		}
		if len(qs) > 1 {
			return fmt.Errorf("multiple questions found")
		}

		files, err := lang.Generate(qs[0])
		if err != nil {
			return err
		}
		err = editor.Open(files)
		return err
	},
}
