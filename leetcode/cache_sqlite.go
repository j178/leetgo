package leetcode

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/j178/leetgo/utils"
	_ "github.com/mattn/go-sqlite3"
)

type sqliteCache struct {
	path string
	once sync.Once
	db   *sql.DB
}

func newSqliteCache(path string) QuestionsCache {
	return &sqliteCache{path: path}
}

func (c *sqliteCache) load() {
	c.once.Do(
		func() {
			var err error
			c.db, err = sql.Open("sqlite3", c.path)
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

func (c *sqliteCache) GetBySlug(slug string) *questionRecord {
	c.load()
	if c.db == nil {
		return nil
	}
	st, err := c.db.Prepare("select frontendId,slug,title,cnTitle,difficulty,tags,paidOnly from questions where slug = ?")
	if err != nil {
		return nil
	}
	var q questionRecord
	var tagsStr string
	err = st.QueryRow(slug).Scan(&q.FrontendId, &q.Slug, &q.Title, &q.CnTitle, &q.Difficulty, &tagsStr, &q.PaidOnly)
	if err != nil {
		return nil
	}
	q.Tags = strings.Split(tagsStr, ",")
	if err != nil {
		return nil
	}
	return &q
}

func (c *sqliteCache) GetById(id string) *questionRecord {
	c.load()
	if c.db == nil {
		return nil
	}
	st, err := c.db.Prepare("select frontendId,slug,title,cnTitle,difficulty,tags,paidOnly from questions where frontendId = ?")
	if err != nil {
		return nil
	}
	var q questionRecord
	var tagsStr string
	err = st.QueryRow(id).Scan(&q.FrontendId, &q.Slug, &q.Title, &q.CnTitle, &q.Difficulty, &tagsStr, &q.PaidOnly)
	if err != nil {
		return nil
	}
	q.Tags = strings.Split(tagsStr, ",")
	if err != nil {
		return nil
	}
	return &q
}

func (c *sqliteCache) createTable() error {
	err := utils.Truncate(c.path)
	if err != nil {
		return err
	}
	c.db, err = sql.Open("sqlite3", c.path)
	if err != nil {
		return err
	}
	_, err = c.db.Exec(
		`
create table questions (
    frontendId varchar(128) unique not null,
    slug varchar(128) primary key not null,
    title varchar(128) not null,
    cnTitle varchar(128) not null,
    difficulty varchar(16) not null,
    tags varchar(128) not null,
    paidOnly tinyint not null
);

create table lastUpdate (
    timestamp bigint not null
);

insert into lastUpdate values (0);
`,
	)
	return err
}

func (c *sqliteCache) Update(client Client) error {
	err := c.createTable()
	if err != nil {
		return err
	}
	all, err := client.GetAllQuestions()
	if err != nil {
		return err
	}
	questions := make([]any, 0, len(all))
	questionsStr := make([]string, 0, len(all))
	for _, q := range all {
		tags := make([]string, 0, len(q.TopicTags))
		for _, t := range q.TopicTags {
			tags = append(tags, t.Slug)
		}
		questionsStr = append(questionsStr, "(?, ?, ?, ?, ?, ?, ?)")
		questions = append(
			questions,
			q.QuestionFrontendId,
			q.TitleSlug,
			q.Title,
			q.TranslatedTitle,
			q.Difficulty,
			strings.Join(tags, ","),
			q.IsPaidOnly,
		)
	}
	stmt := fmt.Sprintf(
		"insert into questions (frontendId, slug, title, cnTitle, difficulty, tags, paidOnly) values %s",
		strings.Join(questionsStr, ","),
	)
	_, err = c.db.Exec(stmt, questions...)
	if err != nil {
		return err
	}
	err = c.updateTime()
	if err != nil {
		return err
	}
	hclog.L().Info("cache updated", "path", c.path)
	return nil
}
