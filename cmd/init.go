package cmd

import (
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/utils"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:     "init DIR",
	Short:   "Init a leetcode workspace",
	Example: "leetgo init . --go",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]
		err := utils.CreateIfNotExists(dir, true)
		if err != nil {
			return err
		}
		err = createConfigDir()
		if err != nil {
			return err
		}
		err = createConfigFile()
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

func createConfigFile() error {
	file := config.Get().ConfigFile()
	if utils.IsExist(file) {
		return nil
	}
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	err = config.Default().WriteTo(f)
	if err != nil {
		return err
	}
	hclog.L().Info("config file created", "file", file)
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
