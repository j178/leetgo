package cmd

import (
	"fmt"

	"github.com/j178/leetgo/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show leetgo config dir",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(config.Get().ConfigDir())
	},
}
