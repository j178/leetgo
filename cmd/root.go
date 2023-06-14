package cmd

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/charmbracelet/log"
	cc "github.com/ivanpirog/coloredcobra"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/constants"
)

func buildVersion() string {
	result := constants.Version
	if constants.Commit != "" {
		result = fmt.Sprintf("%s\ncommit: %s", result, constants.Commit)
	}
	if constants.BuildDate != "" {
		result = fmt.Sprintf("%s\nbuilt at: %s", result, constants.BuildDate)
	}
	result = fmt.Sprintf("%s\ngoos: %s\ngoarch: %s", result, runtime.GOOS, runtime.GOARCH)
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Sum != "" {
		result = fmt.Sprintf("%s\nmodule version: %s, checksum: %s", result, info.Main.Version, info.Main.Sum)
	}
	return result
}

var rootCmd = &cobra.Command{
	Use:           constants.CmdName,
	Short:         "Leetcode",
	Long:          "Leetcode friend for geek.",
	Version:       buildVersion() + "\n\n" + constants.ProjectURL,
	SilenceErrors: true,
	SilenceUsage:  true,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
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
	if config.Debug {
		log.SetReportTimestamp(true)
		log.SetLevel(log.DebugLevel)
	} else {
		log.DebugLevelStyle = log.DebugLevelStyle.SetString("●")
		log.InfoLevelStyle = log.InfoLevelStyle.SetString("●")
		log.WarnLevelStyle = log.WarnLevelStyle.SetString("●")
		log.ErrorLevelStyle = log.ErrorLevelStyle.SetString("×")
		log.FatalLevelStyle = log.FatalLevelStyle.SetString("×")
		log.SetReportTimestamp(false)
		log.SetLevel(log.InfoLevel)
	}
}

func preRun(cmd *cobra.Command, args []string) error {
	initLogger()
	err := initWorkDir()
	if err != nil {
		return err
	}
	err = config.Load(cmd == initCmd)
	return err
}

func initCommands() {
	cobra.EnableCommandSorting = false

	rootCmd.SetOut(os.Stdout)
	rootCmd.InitDefaultVersionFlag()
	rootCmd.Flags().SortFlags = false
	rootCmd.PersistentFlags().StringP("lang", "l", "", "language of code to generate: cpp, go, python ...")
	rootCmd.PersistentFlags().StringP("site", "", "", "leetcode site: cn, us")
	rootCmd.PersistentFlags().BoolP("yes", "y", false, "answer yes to all prompts")
	rootCmd.InitDefaultHelpFlag()
	_ = viper.BindPFlag("code.lang", rootCmd.PersistentFlags().Lookup("lang"))
	_ = viper.BindPFlag("leetcode.site", rootCmd.PersistentFlags().Lookup("site"))
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
		debugCmd,
		gitCmd,
		inspectCmd,
		whoamiCmd,
		openCmd,
	}
	for _, cmd := range commands {
		cmd.Flags().SortFlags = false
		cmd.PersistentPreRunE = preRun
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
