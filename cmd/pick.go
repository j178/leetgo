package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/j178/leetgo/editor"
	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/tui"
)

var skipEditor bool

func init() {
	pickCmd.Flags().BoolVarP(&skipEditor, "skip-editor", "", false, "Skip opening the editor")
}

var pickCmd = &cobra.Command{
	Use:   "pick [qid]",
	Short: "Generate a new question",
	Example: `leetgo pick  # show a list of questions to pick
leetgo pick today
leetgo pick 549
leetgo pick two-sum`,
	Args:      cobra.MaximumNArgs(1),
	Aliases:   []string{"p"},
	ValidArgs: []string{"today", "yesterday"},
	RunE: func(cmd *cobra.Command, args []string) error {
		c := leetcode.NewClient(leetcode.ReadCredentials())
		var q *leetcode.QuestionData

		if len(args) > 0 {
			qid := args[0]
			qs, err := leetcode.ParseQID(qid, c)
			if err != nil {
				return err
			}
			if len(qs) > 1 {
				return fmt.Errorf("`leetgo pick` cannot handle multiple contest questions, use `leetgo contest` instead")
			}
			q = qs[0]
		} else {
			var err error
			q, err = tui.Pick(c)
			if err != nil {
				return err
			}
			if q == nil {
				return nil
			}
		}

		result, err := lang.Generate(q)
		if err != nil {
			return err
		}
		if !skipEditor {
			err = editor.Open(result)
			return err
		}
		return nil
	},
}
