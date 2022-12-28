package cmd

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	configFile                        = ""
	Version                           = "0.0.1"
	defaultConfigFile                 = "leet.yml"
	defaultLeetcodeQuestionsCachePath = "./data/leetcode-questions.json"
)

type Config struct {
	Cn       bool           `json:"cn" yaml:"cn"`
	LeetCode LeetCodeConfig `json:"leetcode" yaml:"leetcode"`
	Go       lang.GoConfig  `json:"go" yaml:"go"`
}

type LeetCodeConfig struct {
	QuestionsCachePath string `json:"questions_cache_path" yaml:"questions_cache_path"`
}

var Opts = Config{
	Cn: true,
	LeetCode: LeetCodeConfig{
		QuestionsCachePath: defaultLeetcodeQuestionsCachePath,
	},
	Go: lang.GoConfig{
		SeparatePackage:  true,
		FilenameTemplate: ``,
	},
}

func initConfig(cmd *cobra.Command, args []string) error {
	if cmd == initCmd {
		return nil
	}
	viper.SetConfigName("leet")
	viper.AddConfigPath(".")
	if configFile != "" {
		viper.SetConfigFile(configFile)
	}
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}
	err = viper.Unmarshal(
		&Opts, func(c *mapstructure.DecoderConfig) {
			c.TagName = "json"
		},
	)
	if err != nil {
		return err
	}

	leetcode.QuestionsCachePath = Opts.LeetCode.QuestionsCachePath
	return err
}

var rootCmd = &cobra.Command{
	Use:               "leet",
	Short:             "Leetcode",
	Long:              "Leetcode command line tool.",
	Version:           Version,
	PersistentPreRunE: initConfig,
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func HelpText() string {
	out := new(bytes.Buffer)
	rootCmd.SetOut(out)
	_ = rootCmd.Help()
	return out.String()
}

func addLangFlags(cmd *cobra.Command) {
	for _, l := range lang.SupportedLanguages {
		cmd.Flags().Bool(strings.ToLower(l.Name()), false, fmt.Sprintf("generate %s output", l.Name()))
	}
}

func init() {
	cobra.EnableCommandSorting = false

	rootCmd.InitDefaultVersionFlag()
	rootCmd.UsageTemplate()
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().Bool("cn", true, "use Chinese")

	_ = rootCmd.MarkPersistentFlagFilename("config", "yml", "yaml")
	_ = viper.BindPFlag("cn", rootCmd.PersistentFlags().Lookup("cn"))

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(todayCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(contestCmd)
	rootCmd.AddCommand(updateCmd)
}
