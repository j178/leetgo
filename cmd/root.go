package cmd

import (
    "fmt"
    "os"
    "strings"

    "github.com/j178/leetgo/lang"
    "github.com/j178/leetgo/leetcode"
    "github.com/mitchellh/mapstructure"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var (
    configFile = ""
    Version    = "0.0.1"
    Opts       Config
)

type Config struct {
    QuestionsDB string        `json:"questions_db"`
    Go          lang.GoConfig `json:"go"`
}

func initConfig() {
    Opts = Config{
        QuestionsDB: "./data/questions.json",
        Go: lang.GoConfig{
            SeparatePackage:  true,
            FilenameTemplate: ``,
        },
    }

    viper.SetConfigName("leet")
    viper.AddConfigPath(".")
    if configFile != "" {
        viper.SetConfigFile(configFile)
    }
    err := viper.ReadInConfig()
    if err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            _, _ = fmt.Fprintln(os.Stderr, err)
            os.Exit(1)
        }
    }
    err = viper.Unmarshal(
        &Opts, func(c *mapstructure.DecoderConfig) {
            c.TagName = "json"
        },
    )
    cobra.CheckErr(err)

    leetcode.DbPath = Opts.QuestionsDB
}

var rootCmd = &cobra.Command{
    Use:     "leet",
    Short:   "Leetcode",
    Long:    "Leetcode command line tool.",
    Version: Version,
}

func Execute() {
    cobra.CheckErr(rootCmd.Execute())
}

func addLangFlags(cmd *cobra.Command) {
    for _, l := range lang.SupportedLanguages {
        cmd.Flags().Bool(strings.ToLower(l.Name()), false, fmt.Sprintf("generate %s output", l.Name()))
    }
}

func init() {
    cobra.OnInitialize(initConfig)
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
