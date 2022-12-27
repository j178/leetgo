package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var todayCmd = &cobra.Command{
	Use:   "today",
	Short: "Generate the question of today",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("today!")
	},
}
