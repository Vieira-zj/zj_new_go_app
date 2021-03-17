package utils

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

/*
Common
*/

// TimeFormat formats timestamp and returns datetime.
func TimeFormat(now time.Time) string {
	return now.Format("2006-01-02 15:04:05")
}

// TimeFormatWithUnderline formats timestamp and returns datetime with underline.
func TimeFormatWithUnderline(now time.Time) string {
	year, month, day := now.Date()
	hour, min, sec := now.Clock()
	return fmt.Sprintf("%d_%d_%d-%d_%d_%d", year, month, day, hour, min, sec)
}

// GetCurPath returns current run abs path.
func GetCurPath() string {
	dir, _ := filepath.Split(os.Args[0])
	return dir
}

// IsFileExist check file exists.
func IsFileExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// DeleteFile deletes a file or an empty directory.
func DeleteFile(path string) error {
	exist, err := IsFileExist(path)
	if err != nil {
		return err
	}
	if !exist {
		return nil
	}
	return os.Remove(path)
}

/*
File Read
*/

// ReadFileText reads file and returns file content string.
func ReadFileText(path string) ([]byte, error) {
	exist, err := IsFileExist(path)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("File (%s) not found", path)
	}

	return ioutil.ReadFile(path)
}

// ReadFileLines reads file and returns all lines.
func ReadFileLines(path string) ([]string, error) {
	exist, err := IsFileExist(path)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("File (%s) not found", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	lines := []string{}
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				lines = append(lines, string(line))
				break
			}
			return nil, err
		}
		lines = append(lines, string(line))
	}
	return lines, nil
}

/*
File Write
*/

// WriteTextToFile writes text to file.
func WriteTextToFile(path, text string, isOverwrite bool) error {
	exist, err := IsFileExist(path)
	if err != nil {
		return err
	}
	if !isOverwrite && exist {
		return fmt.Errorf("File (%s) is exist", path)
	}

	return ioutil.WriteFile(path, []byte(text), 0644)
}

// AppendTextToFile if file exist, appends text at the end, or writes to a new file.
func AppendTextToFile(path, text string) (int, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return f.WriteString(text)
}
