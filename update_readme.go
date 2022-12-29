package main

import (
	"os"
	"strings"

	"github.com/j178/leetgo/cmd"
)

const (
	beginMark = "<!-- BEGIN USAGE -->"
	endMark   = "<!-- END USAGE -->"
)

func main() {
	help := cmd.HelpText()
	help = "\n```\n" + help + "```\n"
	readmeBytes, _ := os.ReadFile("README.md")
	readme := string(readmeBytes)
	usageStart := strings.Index(readme, beginMark) + len(beginMark)
	usageEnd := strings.Index(readme, endMark)
	readme = strings.Replace(readme, readme[usageStart:usageEnd], help, 1)
	_ = os.WriteFile("README.md", []byte(readme), 0644)
}
