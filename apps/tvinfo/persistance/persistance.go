package persistance

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

type Writer interface {
	Append(str string) error
	Contains(str string) (bool, error)
}

type fileWriter struct {
	fileName string
	mu       sync.Mutex
}

func NewFileWriter(path string) *fileWriter {
	fw := new(fileWriter)
	fw.fileName = path
	return fw
}

func (fw *fileWriter) Append(str string) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	f, err := os.OpenFile(fw.fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintln(f, str)
	return err
}

func (fw *fileWriter) Contains(str string) (bool, error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	readFile, err := os.Open(fw.fileName)
	if err != nil {
		return false, err
	}
	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		if fileScanner.Text() == str {
			return true, nil
		}
	}

	return false, nil
}
