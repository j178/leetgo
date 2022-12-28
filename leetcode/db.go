package leetcode

import (
    "encoding/json"
    "errors"
    "os"
    "path/filepath"
)

type QuestionRecord struct {
    FrontendId string   `json:"frontendId"`
    Slug       string   `json:"slug"`
    Title      string   `json:"title"`
    CnTitle    string   `json:"cnTitle"`
    Difficulty string   `json:"difficulty"`
    Tags       []string `json:"tags"`
    PaidOnly   bool     `json:"paidOnly"`
}

type QuestionsDB interface {
    GetBySlug(slug string) *QuestionRecord
    GetById(id string) *QuestionRecord
    Update() error
}

type cache struct {
    path     string
    client   *Client
    slugs    map[string]*QuestionRecord
    frontIds map[string]*QuestionRecord
}

func (c *cache) load() error {
    c.slugs = make(map[string]*QuestionRecord)
    c.frontIds = make(map[string]*QuestionRecord)

    var records []QuestionRecord
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

func (c *cache) Update() error {
    dir := filepath.Dir(c.path)
    if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
        err = os.MkdirAll(dir, os.ModePerm)
        if err != nil {
            return err
        }
    }

    all, err := c.client.GetAllQuestions()
    if err != nil {
        return err
    }
    questions := make([]QuestionRecord, 0, len(all.Array()))
    for _, q := range all.Array() {
        topicTags := q.Get("topicTags").Array()
        tags := make([]string, 0, len(topicTags))
        for _, t := range topicTags {
            tags = append(tags, t.Get("slug").String())
        }
        questions = append(
            questions, QuestionRecord{
                FrontendId: q.Get("questionFrontendId").Str,
                Slug:       q.Get("titleSlug").Str,
                Title:      q.Get("title").Str,
                CnTitle:    q.Get("translatedTitle").String(),
                Difficulty: q.Get("difficulty").Str,
                Tags:       tags,
                PaidOnly:   q.Get("isPaidOnly").Bool(),
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

func (c *cache) GetBySlug(slug string) *QuestionRecord {
    return c.slugs[slug]
}

func (c *cache) GetById(id string) *QuestionRecord {
    return c.frontIds[id]
}

func NewDB(path string, client *Client) QuestionsDB {
    c := &cache{path: path, client: client}
    _ = c.load()
    return c
}
