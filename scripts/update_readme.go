package main

import (
	"bytes"
	"os"

	"github.com/fatih/color"
	"github.com/j178/leetgo/cmd"
	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
	"github.com/jedib0t/go-pretty/v6/table"
)

var (
	beginUsageMark  = []byte("<!-- BEGIN USAGE -->")
	endUsageMark    = []byte("<!-- END USAGE -->")
	beginConfigMark = []byte("<!-- BEGIN CONFIG -->")
	endConfigMark   = []byte("<!-- END CONFIG -->")
	beginMatrixMark = []byte("<!-- BEGIN MATRIX -->")
	endMatrixMark   = []byte("<!-- END MATRIX -->")
)

func main() {
	for _, f := range []string{"README.md", "README_zh.md"} {
		readme, _ := os.ReadFile(f)
		readme = updateUsage(readme)
		readme = updateConfig(readme)
		readme = updateSupportMatrix(readme)
		_ = os.WriteFile(f, readme, 0644)
	}
}

func updateUsage(readme []byte) []byte {
	color.NoColor = true
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
	_ = config.Default().WriteTo(buf)
	configStr := buf.String()
	configStr = "\n```yaml\n" + configStr + "```\n"

	configStart := bytes.Index(readme, beginConfigMark) + len(beginConfigMark)
	configEnd := bytes.Index(readme, endConfigMark)
	result := append([]byte(nil), readme[:configStart]...)
	result = append(result, configStr...)
	result = append(result, readme[configEnd:]...)
	return result
}

func updateSupportMatrix(readme []byte) []byte {
	w := table.NewWriter()
	w.AppendHeader(table.Row{"", "Generate", "Local Test"})
	q := &leetcode.QuestionData{}
	for _, l := range lang.SupportedLanguages {
		_, err := l.GenerateTest(q)
		localTest := ":white_check_mark:"
		if err == lang.NotSupported || err == lang.NotImplemented {
			localTest = ":x:"
		}
		w.AppendRow(
			table.Row{
				l.Name(),
				":white_check_mark:",
				localTest,
			},
		)
	}
	matrixStr := w.RenderMarkdown()
	matrixStr = "\n" + matrixStr + "\n"

	matrixStart := bytes.Index(readme, beginMatrixMark) + len(beginMatrixMark)
	matrixEnd := bytes.Index(readme, endMatrixMark)
	result := append([]byte(nil), readme[:matrixStart]...)
	result = append(result, matrixStr...)
	result = append(result, readme[matrixEnd:]...)
	return result
}
