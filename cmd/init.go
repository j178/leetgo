package cmd

import (
	"os"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var initCmd = &cobra.Command{
	Use:     "init DIR",
	Short:   "Init a leetcode workspace",
	Example: "leet init . --go",
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
	return utils.CreateIfNotExists(config.Get().ConfigDir(), true)
}

func createConfigFile() error {
	if utils.IsExist(config.Get().ConfigFile()) {
		return nil
	}
	f, err := os.Create(config.Get().ConfigFile())
	if err != nil {
		return err
	}
	enc := yaml.NewEncoder(f)
	return enc.Encode(config.Default())
}

func createQuestionDB() error {
	if utils.IsExist(config.Get().LeetCodeCacheFile()) {
		return nil
	}
	c := leetcode.NewClient()
	return leetcode.GetCache().Update(c)
}

func init() {
	addLangFlags(initCmd)
}
