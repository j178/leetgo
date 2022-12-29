package main

import (
	"bytes"
	"os"

	"github.com/j178/leetgo/cmd"
	"github.com/j178/leetgo/config"
	"gopkg.in/yaml.v3"
)

var (
	beginUsageMark  = []byte("<!-- BEGIN USAGE -->")
	endUsageMark    = []byte("<!-- END USAGE -->")
	beginConfigMark = []byte("<!-- BEGIN CONFIG -->")
	endConfigMark   = []byte("<!-- END CONFIG -->")
)

func main() {
	readme, _ := os.ReadFile("README.md")
	readme = updateUsage(readme)
	readme = updateConfig(readme)
	_ = os.WriteFile("README.md", readme, 0644)
}

func updateUsage(readme []byte) []byte {
	usage := cmd.UsageString()
	usage = "\n```\n" + usage + "```\n"

	usageStart := bytes.Index(readme, beginUsageMark) + len(beginUsageMark)
	usageEnd := bytes.Index(readme, endUsageMark)
	result := append([]byte(nil), readme[:usageStart]...)
	result = append(result, usage...)
	result = append(result, readme[usageEnd:]...)
	return result
}

func updateConfig(readme []byte) []byte {
	buf := new(bytes.Buffer)
	cfg := config.Default()
	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)
	_ = enc.Encode(cfg)
	configStr := buf.String()
	configStr = "\n```yaml\n" + configStr + "```\n"

	configStart := bytes.Index(readme, beginConfigMark) + len(beginConfigMark)
	configEnd := bytes.Index(readme, endConfigMark)
	result := append([]byte(nil), readme[:configStart]...)
	result = append(result, configStr...)
	result = append(result, readme[configEnd:]...)
	return result
}
