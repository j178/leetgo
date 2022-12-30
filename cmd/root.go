package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/go-hclog"
	cc "github.com/ivanpirog/coloredcobra"
	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/lang"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// TODO set when building
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
			c.TagName = "yaml"
		},
	)
	if err != nil {
		return err
	}
	if err = config.Verify(cfg); err != nil {
		return fmt.Errorf("config file is invalid: %w", err)
	}

	config.Init(cfg)
	return err
}

var rootCmd = &cobra.Command{
	Use:               config.CmdName,
	Short:             "Leetcode",
	Long:              "Leetcode friend for geek.",
	Version:           Version,
	PersistentPreRunE: loadConfig,
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func UsageString() string {
	return rootCmd.UsageString()
}

var langFlags map[string]*pflag.Flag

func addLangFlags(cmd *cobra.Command) {
	if langFlags == nil {
		langFlags = make(map[string]*pflag.Flag)
		flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
		for _, l := range lang.SupportedLanguages {
			entry := strings.ToLower(l.Name())
			flagSet.Bool(entry, false, fmt.Sprintf("generate %s code", entry))
			langFlags[entry] = flagSet.Lookup(entry)
		}
	}

	for _, l := range lang.SupportedLanguages {
		entry := strings.ToLower(l.Name())
		flag := langFlags[entry]
		cmd.Flags().AddFlag(flag)
		_ = viper.BindPFlag(entry+".enable", flag)
	}
}

func initLogger() {
	opts := &hclog.LoggerOptions{
		Level:           hclog.Info,
		DisableTime:     true,
		Color:           hclog.AutoColor,
		ColorHeaderOnly: true,
	}
	if config.Debug {
		opts.Level = hclog.Trace
		opts.DisableTime = false
		opts.Color = hclog.ColorOff
	}
	hclog.SetDefault(hclog.New(opts))
}

func initCommands() {
	cobra.EnableCommandSorting = false

	rootCmd.Flags().SortFlags = false
	rootCmd.InitDefaultVersionFlag()

	commands := []*cobra.Command{
		initCmd,
		pickCmd,
		todayCmd,
		infoCmd,
		testCmd,
		submitCmd,
		contestCmd,
		updateCmd,
	}
	for _, cmd := range commands {
		cmd.Flags().SortFlags = false
		rootCmd.AddCommand(cmd)
	}

	cc.Init(
		&cc.Config{
			RootCmd:         rootCmd,
			Headings:        cc.HiCyan + cc.Bold + cc.Underline,
			Commands:        cc.HiYellow + cc.Bold,
			Example:         cc.Italic,
			ExecName:        cc.Bold,
			Flags:           cc.Bold,
			NoExtraNewlines: true,
			NoBottomNewline: true,
		},
	)
}

func init() {
	initCommands()
	initLogger()
}
