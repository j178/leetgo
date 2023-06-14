package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/j178/leetgo/cmd"
	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/utils"
)

func main() {
	for _, f := range []string{"README_zh.md", "README_en.md", "README.md"} {
		if !utils.IsExist(f) {
			continue
		}
		readme, _ := os.ReadFile(f)
		readme = updateUsage(readme)
		readme = updateConfig(readme)
		readme = updateSupportMatrix(readme)
		_ = os.WriteFile(f, readme, 0o644)
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
	_ = config.Get().Write(buf, true)
	configStr := buf.String()
	configStr = "\n```yaml\n" + configStr + "```\n"

	return replace("CONFIG", readme, []byte(configStr))
}

func updateSupportMatrix(readme []byte) []byte {
	w := table.NewWriter()
	w.AppendHeader(table.Row{"", "Generation", "Local testing"})
	for _, l := range lang.SupportedLangs {
		localTest := ":white_check_mark:"
		if _, ok := l.(lang.LocalTestable); !ok {
			localTest = "Not yet"
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
