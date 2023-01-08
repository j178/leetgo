package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/j178/leetgo/cmd"
	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/lang"
	"github.com/jedib0t/go-pretty/v6/table"
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

func replace(mark string, origin []byte, new []byte) []byte {
	beginMark := fmt.Appendf(nil, "<!-- BEGIN %s -->", mark)
	endMark := fmt.Appendf(nil, "<!-- END %s -->", mark)
	begin := bytes.Index(origin, beginMark) + len(beginMark)
	end := bytes.Index(origin, endMark)
	result := append([]byte(nil), origin[:begin]...)
	result = append(result, new...)
	result = append(result, origin[end:]...)
	return result
}

func updateUsage(readme []byte) []byte {
	color.NoColor = true
	usage := cmd.UsageString()
	usage = "\n```\n" + usage + "```\n"

	return replace("USAGE", readme, []byte(usage))
}

func updateConfig(readme []byte) []byte {
	buf := new(bytes.Buffer)
	_ = config.Default().Write(buf, true)
	configStr := buf.String()
	configStr = "\n```yaml\n" + configStr + "```\n"

	return replace("CONFIG", readme, []byte(configStr))
}

func updateSupportMatrix(readme []byte) []byte {
	w := table.NewWriter()
	w.AppendHeader(table.Row{"", "Generate", "Local Test"})
	for _, l := range lang.SupportedLangs {
		localTest := ":white_check_mark:"
		if _, ok := l.(lang.LocalTester); !ok {
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

	return replace("MATRIX", readme, []byte(matrixStr))
}
