package cmd

import (
    "fmt"

    "github.com/j178/leetgo/leetcode"
    "github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
    Use:   "new",
    Short: "Generate a new question",
    Args:  cobra.MinimumNArgs(1),
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
