package leetcode

import (
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "sync"
    "time"
)

type questionRecord struct {
    FrontendId string   `json:"frontendId"`
    Slug       string   `json:"slug"`
    Title      string   `json:"title"`
    CnTitle    string   `json:"cnTitle"`
    Difficulty string   `json:"difficulty"`
    Tags       []string `json:"tags"`
    PaidOnly   bool     `json:"paidOnly"`
}

type QuestionsDB interface {
    GetBySlug(slug string) *questionRecord
    GetById(id string) *questionRecord
    Update(client Client) error
}

type cache struct {
    path     string
    slugs    map[string]*questionRecord
    frontIds map[string]*questionRecord
}

func (c *cache) load() error {
    c.slugs = make(map[string]*questionRecord)
    c.frontIds = make(map[string]*questionRecord)

    var records []questionRecord
    if _, err := os.Stat(c.path); errors.Is(err, os.ErrNotExist) {
        return err
    }
    s, err := os.ReadFile(c.path)
    if err != nil {
        return err
    }
    err = json.Unmarshal(s, &records)
    if err != nil {
        return err
    }
    for _, r := range records {
        r := r
        c.slugs[r.Slug] = &r
        c.frontIds[r.FrontendId] = &r
    }
    return nil
}

func (c *cache) checkUpdateTime() {
    stat, err := os.Stat(c.path)
    if errors.Is(err, os.ErrNotExist) {
        return
    }
    if time.Since(stat.ModTime()) >= 14*24*time.Hour {
        _, _ = fmt.Fprintf(os.Stderr, "database is too old, try updating with `leet update`")
    }
}

func (c *cache) Update(client Client) error {
    dir := filepath.Dir(c.path)
    if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
        err = os.MkdirAll(dir, os.ModePerm)
        if err != nil {
            return err
        }
    }

    all, err := client.GetAllQuestions()
    if err != nil {
        return err
    }
    questions := make([]questionRecord, 0, len(all))
    for _, q := range all {
        tags := make([]string, 0, len(q.TopicTags))
        for _, t := range q.TopicTags {
            tags = append(tags, t.Slug)
        }
        questions = append(
            questions, questionRecord{
                FrontendId: q.QuestionFrontendId,
                Slug:       q.TitleSlug,
                Title:      q.Title,
                CnTitle:    q.TranslatedTitle,
                Difficulty: q.Difficulty,
                Tags:       tags,
                PaidOnly:   q.IsPaidOnly,
            },
        )
    }

    f, err := os.Create(c.path)
    if err != nil {
        return err
    }
    enc := json.NewEncoder(f)
    enc.SetIndent("", "\t")
    err = enc.Encode(questions)
    return err
}

func (c *cache) GetBySlug(slug string) *questionRecord {
    return c.slugs[slug]
}

func (c *cache) GetById(id string) *questionRecord {
    return c.frontIds[id]
}

func GetCache() QuestionsDB {
    once.Do(
        func() {
            c := &cache{path: DbPath}
            c.checkUpdateTime()
            err := c.load()
            if err != nil {
                _, _ = fmt.Fprintf(os.Stderr, "failed to load cache: %v", DbPath)
            }
            lazyCache = c
        },
    )

    return lazyCache
}

var (
    once      sync.Once
    lazyCache QuestionsDB
    DbPath    string
)
