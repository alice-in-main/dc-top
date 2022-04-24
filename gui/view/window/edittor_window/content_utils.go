package edittor_window

import (
	"os"
	"strings"
)

func contentsEquals(content1 []string, content2 []string) bool {
	if len(content1) != len(content2) {
		return false
	}
	for i := range content1 {
		if content1[i] != content2[i] {
			return false
		}
	}
	return true
}

func writeNewContent(content []string, file *os.File) error {
	new_raw_content := []byte(strings.Join(content, "\n"))
	err := os.WriteFile(file.Name(), new_raw_content, 0664)
	if err != nil {
		return err
	}
	return nil
}
