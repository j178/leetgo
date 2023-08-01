package main

import (
	"fmt"
	"os"
	"time"

	"github.com/goccy/go-json"

	"github.com/j178/leetgo/leetcode"
)

func main() {
	client := leetcode.NewClient(leetcode.ReadCredentials())
	cache := leetcode.GetCache(client)
	questions := cache.GetAllQuestions()
	paidOnly := 0
	for _, q := range questions {
		if q.IsPaidOnly {
			paidOnly++
		}
	}
	fmt.Printf("Total questions: %d, paid only: %d\n", len(questions), paidOnly)

	for i, q := range questions {
		if q.IsPaidOnly {
			continue
		}
		err := q.Fulfill()
		if err != nil {
			fmt.Printf("fetch error: %s, q=%s\n", err, q.TitleSlug)
			continue
		}
		if i > 0 && i%100 == 0 {
			fmt.Printf("\rfetching %d/%d", i+1, len(questions))
			save(questions)
		}
		time.Sleep(10 * time.Millisecond)
	}
	fmt.Println("\nDone")
}

func save(questions []*leetcode.QuestionData) {
	f, err := os.Create("./misc/questions.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	err = enc.Encode(questions)
	if err != nil {
		panic(err)
	}
}
