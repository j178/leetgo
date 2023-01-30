//go:build cgo

package leetcode

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/hashicorp/go-hclog"
	_ "github.com/mattn/go-sqlite3"

	"github.com/j178/leetgo/utils"
)

const (
	ddl = `
create table questions (
    titleSlug text not null,
	questionId text not null,
	questionFrontendId text not null,
	categoryTitle text not null,
	title text not null,
	translatedTitle text not null,
	difficulty text not null,
	topicTags text not null,
	isPaidOnly tinyint not null,
	content text not null,
	translatedContent text not null,
	status text not null,
	stats text not null,
	hints text not null,
	similarQuestions text not null,
	sampleTestCase text not null,
	exampleTestcases text not null,
	jsonExampleTestcases text not null,
	metaData text not null,
	codeSnippets text not null
);

create table lastUpdate (
    timestamp bigint not null
);

insert into lastUpdate values (0);
`
	columns = "titleSlug,questionId,questionFrontendId,categoryTitle,title,translatedTitle,difficulty,topicTags,isPaidOnly," +
		"content,translatedContent,status,stats,hints,similarQuestions,sampleTestCase,exampleTestcases,jsonExampleTestcases,metaData,codeSnippets"
)

type sqliteCache struct {
	path   string
	client Client
	once   sync.Once
	db     *sql.DB
}

func newCache(path string, c Client) QuestionsCache {
	return &sqliteCache{path: path, client: c}
}

func (c *sqliteCache) GetCacheFile() string {
	return c.path + ".db"
}

func (c *sqliteCache) load() {
	c.once.Do(
		func() {
			defer func(now time.Time) {
				hclog.L().Trace("cache loaded", "path", c.GetCacheFile(), "time", time.Since(now))
			}(time.Now())
			var err error
			c.db, err = sql.Open("sqlite3", c.GetCacheFile())
			if err != nil {
				hclog.L().Warn("failed to load cache, try updating with `leetgo cache update`")
				return
			}
			c.checkUpdateTime()
		},
	)
}

func (c *sqliteCache) checkUpdateTime() {
	if c.db == nil {
		return
	}
	st, err := c.db.Prepare("select timestamp from lastUpdate")
	if err != nil {
		return
	}
	var ts int64
	err = st.QueryRow().Scan(&ts)
	if err != nil {
		return
	}
	if time.Since(time.Unix(ts, 0)) >= 14*24*time.Hour {
		hclog.L().Warn("cache is too old, try updating with `leetgo cache update`")
	}
}

func (c *sqliteCache) updateTime() error {
	st, err := c.db.Prepare("update lastUpdate set timestamp = ? ")
	if err != nil {
		return err
	}
	_, err = st.Exec(time.Now().Unix())
	return err
}

func (c *sqliteCache) unmarshal(rows *sql.Rows) ([]*QuestionData, error) {
	result := make([]*QuestionData, 0)
	for rows.Next() {
		q := QuestionData{
			partial: 1,
			client:  c.client,
		}
		var (
			topicTagsStr            []byte
			statsStr                []byte
			hintsStr                []byte
			similarQuestionsStr     []byte
			jsonExampleTestcasesStr []byte
			metaDataStr             []byte
			codeSnippetsStr         []byte
		)
		err := rows.Scan(
			&q.TitleSlug,
			&q.QuestionId,
			&q.QuestionFrontendId,
			&q.CategoryTitle,
			&q.Title,
			&q.TranslatedTitle,
			&q.Difficulty,
			&topicTagsStr,
			&q.IsPaidOnly,
			&q.Content,
			&q.TranslatedContent,
			&q.Status,
			&statsStr,
			&hintsStr,
			&similarQuestionsStr,
			&q.SampleTestCase,
			&q.ExampleTestcases,
			&jsonExampleTestcasesStr,
			&metaDataStr,
			&codeSnippetsStr,
		)
		if err != nil {
			return nil, err
		}
		_ = json.Unmarshal(topicTagsStr, &q.TopicTags)
		_ = json.Unmarshal(statsStr, &q.Stats)
		_ = json.Unmarshal(hintsStr, &q.Hints)
		_ = json.Unmarshal(similarQuestionsStr, &q.SimilarQuestions)
		_ = json.Unmarshal(jsonExampleTestcasesStr, &q.JsonExampleTestcases)
		_ = json.Unmarshal(metaDataStr, &q.MetaData)
		_ = json.Unmarshal(codeSnippetsStr, &q.CodeSnippets)

		result = append(result, &q)
	}

	return result, nil
}

func (c *sqliteCache) marshal(q *QuestionData) []any {
	topicTagsStr, _ := json.Marshal(q.TopicTags)
	statsStr, _ := json.Marshal(q.Stats)
	hintsStr, _ := json.Marshal(q.Hints)
	similarQuestionsStr, _ := json.Marshal(q.SimilarQuestions)
	jsonExampleTestcasesStr, _ := json.Marshal(q.JsonExampleTestcases)
	metaDataStr, _ := json.Marshal(q.MetaData)
	codeSnippetsStr, _ := json.Marshal(q.CodeSnippets)
	return []any{
		q.TitleSlug,
		q.QuestionId,
		q.QuestionFrontendId,
		q.CategoryTitle,
		q.Title,
		q.TranslatedTitle,
		q.Difficulty,
		topicTagsStr,
		q.IsPaidOnly,
		q.Content,
		q.TranslatedContent,
		q.Status,
		statsStr,
		hintsStr,
		similarQuestionsStr,
		q.SampleTestCase,
		q.ExampleTestcases,
		jsonExampleTestcasesStr,
		metaDataStr,
		codeSnippetsStr,
	}
}

func (c *sqliteCache) GetBySlug(slug string) *QuestionData {
	c.load()
	if c.db == nil {
		return nil
	}
	st, err := c.db.Prepare("select * from questions where titleSlug = ?")
	if err != nil {
		return nil
	}
	rows, err := st.Query(slug)
	if err != nil {
		return nil
	}
	q, err := c.unmarshal(rows)
	if err != nil {
		return nil
	}
	if len(q) == 0 {
		return nil
	}
	return q[0]
}

func (c *sqliteCache) GetById(id string) *QuestionData {
	c.load()
	if c.db == nil {
		return nil
	}
	st, err := c.db.Prepare("select * from questions where questionFrontendId = ?")
	if err != nil {
		return nil
	}
	rows, err := st.Query(id)
	if err != nil {
		return nil
	}
	q, err := c.unmarshal(rows)
	if err != nil {
		return nil
	}
	if len(q) == 0 {
		return nil
	}
	return q[0]
}

func (c *sqliteCache) GetAllQuestions() []*QuestionData {
	c.load()
	if c.db == nil {
		return nil
	}
	st, err := c.db.Prepare("select * from questions")
	if err != nil {
		return nil
	}
	rows, err := st.Query()
	if err != nil {
		return nil
	}
	qs, err := c.unmarshal(rows)
	if err != nil {
		return nil
	}
	return qs
}

func (c *sqliteCache) createTable() error {
	err := utils.Truncate(c.GetCacheFile())
	if err != nil {
		return err
	}
	c.db, err = sql.Open("sqlite3", c.GetCacheFile())
	if err != nil {
		return err
	}
	_, err = c.db.Exec(ddl)
	return err
}

func (c *sqliteCache) Update() error {
	err := c.createTable()
	if err != nil {
		return err
	}
	all, err := c.client.GetAllQuestions()
	if err != nil {
		return err
	}
	placeholder := "(" + strings.Repeat("?,", 19) + "?)"
	batch := 100
	for len(all) > 0 {
		size := min(batch, len(all))
		questions := make([]any, 0, size*20)
		questionsStr := make([]string, 0, size)
		for i := 0; i < size; i++ {
			questionsStr = append(questionsStr, placeholder)
			questions = append(questions, c.marshal(all[i])...)
		}
		stmt := fmt.Sprintf(
			"insert into questions (%s) values %s",
			columns,
			strings.Join(questionsStr, ","),
		)
		_, err = c.db.Exec(stmt, questions...)
		if err != nil {
			return err
		}
		all = all[size:]
	}

	err = c.updateTime()
	if err != nil {
		return err
	}
	hclog.L().Info("cache updated", "path", c.GetCacheFile())
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
