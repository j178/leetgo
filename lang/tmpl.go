package lang

import (
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/leetcode"
	"github.com/spf13/viper"
)

const contentTemplate = `
{{- block "header" . -}}
{{ .LineComment }} Created by {{ .Author }} at {{ .Time }}
{{ .LineComment }} {{ .Question.Url }}
{{ if .Question.IsContest }}{{ .LineComment }} {{ .Question.ContestUrl }}
{{ end }}
{{ end }}
{{ block "description" . -}}
{{ .BlockCommentStart }}
{{ block "title" . }}{{ .Question.QuestionFrontendId }}. {{ .Question.GetTitle }} ({{ .Question.Difficulty }}){{ end }}

{{ .Question.GetFormattedContent }}
{{ .BlockCommentEnd }}
{{ end }}
{{ block "beforeMarker" . }}{{ end }}
{{ .LineComment }} {{ .CodeBeginMarker }}
{{ block "beforeCode" . }}{{ end }}
{{ block "code" . }}{{ .Code | runModifiers }}{{ end }}
{{ block "afterCode" . }}{{ end }}
{{ .LineComment }} {{ .CodeEndMarker }}
{{ block "afterMarker" . }}{{ end }}
`

type contentData struct {
	Question          *leetcode.QuestionData
	Author            string
	Time              string
	LineComment       string
	BlockCommentStart string
	BlockCommentEnd   string
	CodeBeginMarker   string
	CodeEndMarker     string
	Code              string
	NeedsDefinition   bool
}

var validBlocks = map[string]bool{
	"header":       true,
	"description":  true,
	"title":        true,
	"beforeMarker": true,
	"beforeCode":   true,
	"code":         true,
	"afterCode":    true,
	"afterMarker":  true,
}

var builtinModifiers = map[string]ModifierFunc{
	"removeUselessComments": removeUselessComments,
}

type ModifierFunc func(string, *leetcode.QuestionData) string

func getBlocks(lang Lang) (ans []config.Block) {
	blocks := viper.Get("code." + lang.Slug() + ".blocks")
	if blocks == nil || len(blocks.([]any)) == 0 {
		blocks = viper.Get("code." + lang.ShortName() + ".blocks")
	}
	if blocks == nil || len(blocks.([]any)) == 0 {
		blocks = viper.Get("code.blocks")
	}
	if blocks == nil {
		return
	}
	for _, b := range blocks.([]any) {
		ans = append(
			ans, config.Block{
				Name:     b.(map[string]any)["name"].(string),
				Template: b.(map[string]any)["template"].(string),
			},
		)
	}
	return
}

func getModifiers(lang Lang, modifiersMap map[string]ModifierFunc) (ans []ModifierFunc) {
	modifiers := viper.Get("code." + lang.Slug() + ".modifiers")
	if modifiers == nil || len(modifiers.([]any)) == 0 {
		modifiers = viper.Get("code." + lang.ShortName() + ".modifiers")
	}
	if modifiers == nil || len(modifiers.([]any)) == 0 {
		modifiers = viper.Get("code.modifiers")
	}
	if modifiers == nil {
		return
	}

	for _, m := range modifiers.([]any) {
		name := m.(map[string]any)["name"].(string)
		fun := m.(map[string]any)["func"]
		if f, ok := modifiersMap[name]; ok {
			ans = append(ans, f)
		} else if fun != nil {
			// TODO support js func
			_ = fun
			hclog.L().Warn("custom modifier not supported yet, ignored", "modifier", name)
		} else {
			hclog.L().Warn("modifier not supported, ignored", "modifier", name)
		}
	}
	return
}

func needsDefinition(code string) bool {
	return strings.Contains(code, "Definition for")
}

func needsMod(content string) bool {
	return strings.Contains(content, "<code>10<sup>9</sup> + 7</code>") || strings.Contains(content, "10^9 + 7")
}

func removeUselessComments(code string, q *leetcode.QuestionData) string {
	lines := strings.Split(code, "\n")
	var newLines []string
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "/**") && (strings.Contains(
			lines[i+1],
			"object will be instantiated and called",
		) || strings.Contains(lines[i+1], "Definition for")) {
			for {
				i++
				if strings.HasSuffix(lines[i], "*/") {
					break
				}
			}
			continue
		}
		newLines = append(newLines, line)
	}
	return strings.Join(newLines, "\n")
}
