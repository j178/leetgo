package cmd

import "github.com/spf13/cobra"

var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Init a leetcode workspace",
    Run: func(cmd *cobra.Command, args []string) {
        // 生成配置文件
        // 生成数据库
        // 生成目录
        // 写入基础库代码
    },
}
