package utils

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// IsExist checks if a file or directory exists
func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return false
}

func MakeDir(dir string) error {
	return os.MkdirAll(dir, 0o755)
}

// CreateIfNotExists creates a file or a directory only if it does not already exist.
func CreateIfNotExists(path string, isDir bool) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if isDir {
				return os.MkdirAll(path, 0o755)
			}
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			f, err := os.OpenFile(path, os.O_CREATE, 0o755)
			if err != nil {
				return err
			}
			f.Close()
		}
	}
	return nil
}

func WriteFile(file string, content []byte) error {
	err := CreateIfNotExists(file, false)
	if err != nil {
		return err
	}
	err = os.WriteFile(file, content, 0o644)
	if err != nil {
		return err
	}
	return nil
}

func RemoveIfExist(path string) error {
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func Truncate(filename string) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return nil
}

func RelToCwd(path string) string {
	wd, err := os.Getwd()
	if err != nil {
		return path
	}
	relPath, err := filepath.Rel(wd, path)
	if err != nil {
		relPath = path
	}
	return filepath.ToSlash(relPath)
}

// CopyFS copies a file system to a directory.
func CopyFS(dir string, fsys fs.FS) error {
	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		targ := filepath.Join(dir, filepath.FromSlash(path))
		if d.IsDir() {
			if err := os.MkdirAll(targ, 0777); err != nil {
				return err
			}
			return nil
		}
		r, err := fsys.Open(path)
		if err != nil {
			return err
		}
		defer r.Close()
		info, err := r.Stat()
		if err != nil {
			return err
		}
		w, err := os.OpenFile(targ, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666|info.Mode()&0777)
		if err != nil {
			return err
		}
		if _, err := io.Copy(w, r); err != nil {
			w.Close()
			return fmt.Errorf("copying %s: %v", path, err)
		}
		if err := w.Close(); err != nil {
			return err
		}
		return nil
	})
}
