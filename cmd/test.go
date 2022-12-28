package cmd

import "github.com/spf13/cobra"

var testCmd = &cobra.Command{
    Use:   "test",
    Short: "Run test cases",
    RunE: func(cmd *cobra.Command, args []string) error {
        return nil
    },
}
