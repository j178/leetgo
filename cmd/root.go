package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/lang"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Version = "0.0.1"
)

func loadConfig(cmd *cobra.Command, args []string) error {
	cfg := config.Default()
	viper.SetConfigFile(cfg.ConfigFile())
	err := viper.ReadInConfig()
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	err = viper.Unmarshal(
		&cfg, func(c *mapstructure.DecoderConfig) {
			c.TagName = "json"
		},
	)
	if err != nil {
		return err
	}

	config.Init(cfg)
	return err
}

var rootCmd = &cobra.Command{
	Use:               "leet",
	Short:             "Leetcode",
	Long:              "Leetcode command line tool.",
	Version:           Version,
	PersistentPreRunE: loadConfig,
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func UsageString() string {
	return rootCmd.UsageString()
}

func addLangFlags(cmd *cobra.Command) {
	for _, l := range lang.SupportedLanguages {
		cmd.Flags().Bool(strings.ToLower(l.Name()), false, fmt.Sprintf("generate %s output", l.Name()))
	}
}

func init() {
	cobra.EnableCommandSorting = false

	rootCmd.InitDefaultVersionFlag()
	rootCmd.PersistentFlags().Bool("cn", true, "use Chinese")
	_ = viper.BindPFlag("cn", rootCmd.PersistentFlags().Lookup("cn"))

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(todayCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(submitCmd)
	rootCmd.AddCommand(contestCmd)
	rootCmd.AddCommand(updateCmd)
}
