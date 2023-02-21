package cmd

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/charmbracelet/log"
	"github.com/fatih/color"
	cc "github.com/ivanpirog/coloredcobra"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/j178/leetgo/config"
)

var (
	version = "0.0.1"
	commit  = "HEAD"
	date    = "unknown"
)

func buildVersion() string {
	result := version
	if commit != "" {
		result = fmt.Sprintf("%s\ncommit: %s", result, commit)
	}
	if date != "" {
		result = fmt.Sprintf("%s\nbuilt at: %s", result, date)
	}
	result = fmt.Sprintf("%s\ngoos: %s\ngoarch: %s", result, runtime.GOOS, runtime.GOARCH)
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Sum != "" {
		result = fmt.Sprintf("%s\nmodule version: %s, checksum: %s", result, info.Main.Version, info.Main.Sum)
	}
	return result
}

var rootCmd = &cobra.Command{
	Use:           config.CmdName,
	Short:         "Leetcode",
	Long:          "Leetcode friend for geek.",
	Version:       buildVersion() + "\n\n" + config.ProjectURL,
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		initLogger()
		err := initWorkDir()
		if err != nil {
			return err
		}
		err = config.Load(cmd == initCmd)
		if err != nil {
			return fmt.Errorf(
				"%w\nSeems like your configuration is not a valid YAML file, please paste your configuration to tools like https://www.yamllint.com/ to fix it.",
				err,
			)
		}
		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, color.New(color.FgHiRed).Sprint("Error:"), err)
		os.Exit(1)
	}
}

func UsageString() string {
	return rootCmd.UsageString()
}

func initWorkDir() error {
	if dir := os.Getenv("LEETGO_WORKDIR"); dir != "" {
		log.Debug("change workdir to LEETGO_WORKDIR", "dir", dir)
		return os.Chdir(dir)
	}
	return nil
}

func initLogger() {
	log.SetReportTimestamp(false)
	log.SetLevel(log.InfoLevel)
	if config.Debug {
		log.SetReportTimestamp(true)
		log.SetLevel(log.DebugLevel)
	}
}

func initCommands() {
	cobra.EnableCommandSorting = false

	rootCmd.SetOut(os.Stdout)
	rootCmd.InitDefaultVersionFlag()
	rootCmd.Flags().SortFlags = false
	rootCmd.PersistentFlags().StringP("lang", "l", "", "language of code to generate: cpp, go, python ...")
	rootCmd.PersistentFlags().BoolP("yes", "y", false, "answer yes to all prompts")
	rootCmd.InitDefaultHelpFlag()
	_ = viper.BindPFlag("code.lang", rootCmd.PersistentFlags().Lookup("lang"))
	_ = viper.BindPFlag("yes", rootCmd.PersistentFlags().Lookup("yes"))

	commands := []*cobra.Command{
		initCmd,
		pickCmd,
		infoCmd,
		testCmd,
		submitCmd,
		fixCmd,
		editCmd,
		extractCmd,
		contestCmd,
		cacheCmd,
		configCmd,
		gitCmd,
		inspectCmd,
		whoamiCmd,
		openCmd,
	}
	for _, cmd := range commands {
		cmd.Flags().SortFlags = false
		rootCmd.AddCommand(cmd)
	}
	rootCmd.InitDefaultHelpCmd()

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
}
