package lang

import (
	"errors"
	"os"

	"github.com/goccy/go-json"

	"github.com/j178/leetgo/config"
)

// If client dependency needs to be updated, update this version number.
var depVersions = map[string]int{
	cppGen.slug:     1,
	golangGen.slug:  3,
	python3Gen.slug: 1,
	rustGen.slug:    1,
}

func readDepVersions() (map[string]int, error) {
	depVersionFile := config.Get().DepVersionFile()
	records := make(map[string]int)
	f, err := os.Open(depVersionFile)
	if errors.Is(err, os.ErrNotExist) {
		return records, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&records)
	if err != nil {
		return nil, err
	}
	return records, nil
}

func IsDepUpdateToDate(lang Lang) (bool, error) {
	ver := depVersions[lang.Slug()]
	if ver == 0 {
		return true, nil
	}

	records, err := readDepVersions()
	if err != nil {
		return false, err
	}
	old := records[lang.Slug()]
	if old == 0 || old != ver {
		return false, nil
	}

	return true, nil
}

func UpdateDep(lang Lang) error {
	ver := depVersions[lang.Slug()]
	if ver == 0 {
		return nil
	}

	records, err := readDepVersions()
	if err != nil {
		return err
	}

	records[lang.Slug()] = ver

	depVersionFile := config.Get().DepVersionFile()
	f, err := os.Create(depVersionFile)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	err = enc.Encode(records)
	if err != nil {
		return err
	}

	return nil
}
