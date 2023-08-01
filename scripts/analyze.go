package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-json"

	"github.com/j178/leetgo/config"
	"github.com/j178/leetgo/lang"
	"github.com/j178/leetgo/leetcode"
)

func main() {
	f, err := os.Open("misc/questions.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_ = os.Chdir(os.Getenv("LEETGO_WORKDIR"))
	err = config.Load(false)
	if err != nil {
		panic(err)
	}

	var questions []*leetcode.QuestionData
	err = json.NewDecoder(f).Decode(&questions)
	if err != nil {
		panic(err)
	}
	c := leetcode.NewClient(leetcode.NonAuth())

	categories := map[leetcode.CategoryTitle]int{}
	for _, q := range questions {
		q.SetClient(c)

		categories[q.CategoryTitle]++
		if q.MetaData.Manual && q.CategoryTitle == leetcode.CategoryAlgorithms {
			fmt.Printf("%s.%s\n", q.QuestionFrontendId, q.TitleSlug)
			out, err := lang.Generate(q)
			if err != nil {
				fmt.Println(err)
			}
			f, _ := os.Create(filepath.Join(out.TargetDir(), "question.json"))
			enc := json.NewEncoder(f)
			enc.SetIndent("", "  ")
			enc.Encode(q)
			f.Close()
		}
	}

	fmt.Printf(
		"total: %d, %v, manual: %d\n",
		len(questions),
		categories,
	)
}
