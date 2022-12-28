package cmd

import (
	"os"
	"path/filepath"

	"github.com/j178/leetgo/leetcode"
	"github.com/j178/leetgo/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Init a leetcode workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]
		if !utils.IsExist(dir) {
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return err
			}
		}
		createConfigFile(dir)
		createQuestionDB(dir)
		// 生成目录
		// 写入基础库代码

		return nil
	},
}

func createConfigFile(dir string) error {
	f, err := os.Create(filepath.Join(dir, defaultConfigFile))
	if err != nil {
		return err
	}
	enc := yaml.NewEncoder(f)
	return enc.Encode(Opts)
}

func createQuestionDB(dir string) error {
	leetcode.QuestionsCachePath = filepath.Join(dir, defaultLeetcodeQuestionsCachePath)
	c := leetcode.NewClient()
	return leetcode.GetCache().Update(c)
}
