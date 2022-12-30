package cmd

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	cc "github.com/ivanpirog/coloredcobra"
	"github.com/j178/leetgo/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// TODO set when building
var (
	Version = "0.0.1"
)

func loadConfig(cmd *cobra.Command, args []string) error {
	if cmd == initCmd {
		return nil
	}
	cfg := config.Default()
	viper.SetConfigFile(cfg.ConfigFile())
	err := viper.ReadInConfig()
	if err != nil {
		if os.IsNotExist(err) {
			hclog.L().Debug("config file not found, have you ran `leetgo init`?")
			return nil
		}
		return err
	}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return err
	}
	if err = config.Verify(cfg); err != nil {
		return fmt.Errorf("config file is invalid: %w", err)
	}

	config.Set(cfg)
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
	rootCmd.PersistentFlags().StringP("gen", "g", "", "language to generate: cpp, go, python ...")
	_ = viper.BindPFlag("gen", rootCmd.PersistentFlags().Lookup("gen"))

	commands := []*cobra.Command{
		initCmd,
		pickCmd,
		todayCmd,
		infoCmd,
		testCmd,
		submitCmd,
		contestCmd,
		cacheCmd,
		configCmd,
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
