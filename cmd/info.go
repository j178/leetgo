package cmd

import (
    "fmt"

    "github.com/j178/leetgo/leetcode"
    "github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
    Use:   "info",
    Short: "Show question info",
    Args:  cobra.MinimumNArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        c := leetcode.NewClient()
        for _, s := range args {
            q, _ := leetcode.Question(s, c)
            fmt.Printf("%v\n", q)
        }
        return nil
    },
}
