package cmd

import (
	"fmt"
	"os"

	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configFile string

var rootCmd = &cobra.Command{
	Use:   "leet SLUG_OR_ID...",
	Short: "Leetcode",
	Long:  "Leetcode command line tool.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.SetConfigFile(configFile)
		err := viper.ReadInConfig()
		if err != nil {
			if err != nil {
				switch err.(type) {
				case viper.ConfigParseError:
					return err
				}
			}
		}

		c := leetcode.NewClient()
		for _, p := range args {
			fmt.Println(c.GetQuestionData(p))
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "leet.yml", "config file path")
	rootCmd.PersistentFlags().Bool("cn", true, "use Chinese")
	for _, l := range lang.SupportedLanguages {
		rootCmd.PersistentFlags().Bool(l.Name(), false, fmt.Sprintf("generate %s output", l.Name()))
	}
	rootCmd.MarkPersistentFlagFilename("config", "yml", "yaml")
	_ = viper.BindPFlag("cn", rootCmd.PersistentFlags().Lookup("cn"))

	rootCmd.AddCommand(todayCmd)
}
