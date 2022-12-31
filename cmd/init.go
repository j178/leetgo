package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var initCmd = &cobra.Command{
	Use:     "init DIR",
	Short:   "Init a leetcode workspace",
	Example: "leetgo init . --gen go",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		gen := viper.GetString("gen")
		if gen == "" {
			return fmt.Errorf("--gen is required for init")
		}
		dir := args[0]
		err := utils.CreateIfNotExists(dir, true)
		if err != nil {
			return err
		}
		err = createConfigDir()
		if err != nil {
			return err
		}
		err = createConfigFiles(dir, gen)
		if err != nil {
			return err
		}
		err = createQuestionDB()
		if err != nil {
			return err
		}
		// 生成目录
		// 写入基础库代码
		return nil
	},
}

func createConfigDir() error {
	dir := config.Get().ConfigDir()
	if utils.IsExist(dir) {
		return nil
	}
	err := utils.MakeDir(dir)
	if err != nil {
		return err
	}
	hclog.L().Info("config dir created", "dir", dir)
	return nil
}

func createConfigFiles(dir string, gen string) error {
	cfg := config.Default()
	globalFile := cfg.GlobalConfigFile()
	if !utils.IsExist(globalFile) {
		f, err := os.Create(globalFile)
		if err != nil {
			return err
		}
		err = cfg.Write(f)
		if err != nil {
			return err
		}
		hclog.L().Info("global config file created", "file", globalFile)
	}

	projectFile := filepath.Join(dir, cfg.ProjectConfigFilename())
	f, err := os.Create(projectFile)
	if err != nil {
		return err
	}
	_, _ = f.WriteString("gen: " + gen + "\n")
	hclog.L().Info("project config file created", "file", projectFile)

	return nil
}

func createQuestionDB() error {
	if utils.IsExist(config.Get().LeetCodeCacheFile()) {
		return nil
	}
	c := leetcode.NewClient()
	err := leetcode.GetCache().Update(c)
	if err != nil {
		return err
	}
	return nil
}
