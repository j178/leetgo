package lang

import (
	"bufio"
	"os"
	"strings"
)

func ReadVersion(file string) (string, error) {
	headerFile, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer headerFile.Close()

	scanner := bufio.NewScanner(headerFile)
	if scanner.Scan() {
		versionLine := scanner.Text()
		version := versionLine[strings.Index(versionLine, "version: ")+len("version: "):]
		version = strings.TrimSpace(version)
		return version, nil
	}

	return "", scanner.Err()
}
