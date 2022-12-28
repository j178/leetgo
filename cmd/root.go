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
}

var rootCmd = &cobra.Command{
    Use:     "leet SLUG_OR_ID...",
    Short:   "Leetcode",
    Long:    "Leetcode command line tool.",
    Args:    cobra.MinimumNArgs(1),
    Version: Version,
    RunE: func(cmd *cobra.Command, args []string) error {
        c := leetcode.NewClient()
        for _, p := range args {
            fmt.Println(c.GetQuestionData(p))
        }

        return nil
    },
}

func Execute() {
    cobra.CheckErr(rootCmd.Execute())
}

func init() {
    cobra.OnInitialize(initConfig)

    rootCmd.InitDefaultVersionFlag()
    rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file path")
    rootCmd.PersistentFlags().Bool("cn", true, "use Chinese")
    for _, l := range lang.SupportedLanguages {
        rootCmd.PersistentFlags().Bool(strings.ToLower(l.Name()), false, fmt.Sprintf("generate %s output", l.Name()))
    }
    _ = rootCmd.MarkPersistentFlagFilename("config", "yml", "yaml")
    _ = viper.BindPFlag("cn", rootCmd.PersistentFlags().Lookup("cn"))

    rootCmd.AddCommand(initCmd)
    rootCmd.AddCommand(updateCmd)
    rootCmd.AddCommand(todayCmd)
}
