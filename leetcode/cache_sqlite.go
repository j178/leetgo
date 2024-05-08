//go:build sqlite

package leetcode

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/goccy/go-json"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"

	"github.com/j178/leetgo/utils"
)

const (
	questionsDDL = `
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
);`

	timestampDDL = `
create table lastUpdate (
    timestamp bigint not null
);`

	initTimestamp = `
insert into lastUpdate values (0);`
	columns = "titleSlug,questionId,questionFrontendId,categoryTitle,title,translatedTitle,difficulty,topicTags,isPaidOnly," +
		"content,translatedContent,status,stats,hints,similarQuestions,sampleTestCase,exampleTestcases,jsonExampleTestcases,metaData,codeSnippets"
)

var cacheExt = ".db"

type sqliteCache struct {
	path   string
	client Client
	once   sync.Once
	db     *sqlite.Conn
}

func newCache(path string, c Client) QuestionsCache {
	return &sqliteCache{path: path, client: c}
}

func (c *sqliteCache) CacheFile() string {
	return c.path
}

func (c *sqliteCache) load() {
	c.once.Do(
		func() {
			defer func(start time.Time) {
				log.Debug("cache loaded", "path", c.path, "elapsed", time.Since(start))
			}(time.Now())
			var err error
			c.db, err = sqlite.OpenConn(c.path)
			if err != nil {
				log.Error("failed to load cache, try updating with `leetgo cache update`")
				return
			}
			if c.Outdated() {
				log.Warn("cache is too old, try updating with `leetgo cache update`")
			}
		},
	)
}

func (c *sqliteCache) Outdated() bool {
	// Cannot use c.load() here, because it will cause a deadlock.
	db, err := sqlite.OpenConn(c.path)
	if err != nil {
		return true
	}

	var ts int64
	err = sqlitex.Execute(
		db, "select timestamp from lastUpdate", &sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				ts = stmt.ColumnInt64(0)
				return nil
			},
		},
	)
	if err != nil {
		return true
	}

	return time.Since(time.Unix(ts, 0)) >= 14*24*time.Hour
}

func (c *sqliteCache) updateLastUpdate() error {
	err := sqlitex.Execute(
		c.db, "update lastUpdate set timestamp = ?", &sqlitex.ExecOptions{
			Args: []any{time.Now().Unix()},
		},
	)
	return err
}

func (c *sqliteCache) unmarshal(stmt *sqlite.Stmt) (*QuestionData, error) {
	q := QuestionData{
		partial: 1,
		client:  c.client,
	}

	q.TitleSlug = stmt.ColumnText(0)
	q.QuestionId = stmt.ColumnText(1)
	q.QuestionFrontendId = stmt.ColumnText(2)
	q.CategoryTitle = CategoryTitle(stmt.ColumnText(3))
	q.Title = stmt.ColumnText(4)
	q.TranslatedTitle = stmt.ColumnText(5)
	q.Difficulty = stmt.ColumnText(6)
	n := stmt.ColumnLen(7)
	topicTagsStr := make([]byte, n)
	stmt.ColumnBytes(7, topicTagsStr)
	q.IsPaidOnly = stmt.ColumnBool(8)
	q.Content = stmt.ColumnText(9)
	q.TranslatedContent = stmt.ColumnText(10)
	q.Status = stmt.ColumnText(11)
	n = stmt.ColumnLen(12)
	statsStr := make([]byte, n)
	stmt.ColumnBytes(12, statsStr)
	n = stmt.ColumnLen(13)
	hintsStr := make([]byte, n)
	stmt.ColumnBytes(13, hintsStr)
	n = stmt.ColumnLen(14)
	similarQuestionsStr := make([]byte, n)
	stmt.ColumnBytes(14, similarQuestionsStr)
	q.SampleTestCase = stmt.ColumnText(15)
	q.ExampleTestcases = stmt.ColumnText(16)
	n = stmt.ColumnLen(17)
	jsonExampleTestcasesStr := make([]byte, n)
	stmt.ColumnBytes(17, jsonExampleTestcasesStr)
	n = stmt.ColumnLen(18)
	metaDataStr := make([]byte, n)
	stmt.ColumnBytes(18, metaDataStr)
	n = stmt.ColumnLen(19)
	codeSnippetsStr := make([]byte, n)
	stmt.ColumnBytes(19, codeSnippetsStr)

	err := json.Unmarshal(topicTagsStr, &q.TopicTags)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(statsStr, &q.Stats)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(hintsStr, &q.Hints)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(similarQuestionsStr, &q.SimilarQuestions)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jsonExampleTestcasesStr, &q.JsonExampleTestcases)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(metaDataStr, &q.MetaData)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(codeSnippetsStr, &q.CodeSnippets)
	if err != nil {
		return nil, err
	}
	return &q, nil
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

	var (
		q   *QuestionData
		err error
	)
	err = sqlitex.Execute(
		c.db, "select * from questions where titleSlug = ?", &sqlitex.ExecOptions{
			Args: []any{slug},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				q, err = c.unmarshal(stmt)
				return err
			},
		},
	)
	if err != nil {
		return nil
	}
	return q
}

func (c *sqliteCache) GetById(id string) *QuestionData {
	defer func(start time.Time) {
		log.Debug("get by id", "elapsed", time.Since(start))
	}(time.Now())

	c.load()
	if c.db == nil {
		return nil
	}

	var (
		q   *QuestionData
		err error
	)
	err = sqlitex.Execute(
		c.db, "select * from questions where questionId = ?", &sqlitex.ExecOptions{
			Args: []any{id},
			ResultFunc: func(stmt *sqlite.Stmt) error {
				q, err = c.unmarshal(stmt)
				return err
			},
		},
	)
	if err != nil {
		return nil
	}
	return q
}

func (c *sqliteCache) GetAllQuestions() []*QuestionData {
	c.load()
	if c.db == nil {
		return nil
	}

	var qs []*QuestionData
	err := sqlitex.Execute(
		c.db, "select * from questions", &sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				q, err := c.unmarshal(stmt)
				if err != nil {
					return err
				}
				qs = append(qs, q)
				return nil
			},
		},
	)
	if err != nil {
		return nil
	}
	return qs
}

func (c *sqliteCache) createTable() error {
	err := utils.CreateIfNotExists(c.path, false)
	if err != nil {
		return err
	}
	err = utils.Truncate(c.path)
	if err != nil {
		return err
	}
	c.db, err = sqlite.OpenConn(c.path)
	if err != nil {
		return err
	}
	err = sqlitex.Execute(c.db, questionsDDL, nil)
	if err != nil {
		return err
	}
	err = sqlitex.Execute(c.db, timestampDDL, nil)
	if err != nil {
		return err
	}
	err = sqlitex.Execute(c.db, initTimestamp, nil)
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
	count := len(all)
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
		err = sqlitex.Execute(
			c.db, stmt, &sqlitex.ExecOptions{
				Args: questions,
			},
		)
		if err != nil {
			return err
		}
		all = all[size:]
	}

	err = c.updateLastUpdate()
	if err != nil {
		return err
	}
	log.Info("questions cache updated", "count", count, "path", c.path)
	return nil
}
