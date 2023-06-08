package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/j178/leetgo/editor"
	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
)

var editCmd = &cobra.Command{
	Use:     "edit qid",
	Short:   "Open solution in editor",
	Aliases: []string{"e"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := leetcode.NewClient(leetcode.ReadCredentials())
		qs, err := leetcode.ParseQID(args[0], c)
		if err != nil {
			return err
		}
		if len(qs) > 1 {
			return fmt.Errorf("multiple questions found")
		}
		result, err := lang.GeneratePathsOnly(qs[0])
		if err != nil {
			return err
		}
		return editor.Open(result)
	},
}

var extractCmd = &cobra.Command{
	Use:    "extract qid",
	Short:  "Extract solution code from generated file",
	Args:   cobra.ExactArgs(1),
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := leetcode.NewClient(leetcode.ReadCredentials())
		qs, err := leetcode.ParseQID(args[0], c)
		if err != nil {
			return err
		}
		if len(qs) > 1 {
			return fmt.Errorf("multiple questions found")
		}
		code, err := lang.GetSolutionCode(qs[0])
		if err != nil {
			return err
		}
		cmd.Println(code)
		return nil
	},
}
