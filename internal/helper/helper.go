package helper

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"
)

// Helper Methods

func CreateFolder(dirPath string) error {
	_, err := os.Stat(dirPath)

	if os.IsNotExist(err) {
		err = os.MkdirAll(dirPath, 0700)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

func DeleteFolder(dirPath string) error {
	// Check if the folder exists
	_, err := os.Stat(dirPath)

	if err == nil {
		// Delete the folder and its contents
		err := os.RemoveAll(dirPath)

		return err
	} else if os.IsNotExist(err) {
		return nil
	} else {
		return err
	}
}

func ReadLine(filename string, lineNumber int) (string, error) {
	if lineNumber < 0 {
		return "", errors.New("invalid line number")
	}

	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	currentLine := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return "", err
		}

		if currentLine == lineNumber {
			return line, nil
		}

		if err == io.EOF {
			break
		}

		currentLine++
	}

	return "", errors.New("line not found")
}

func CountLinesInFile(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	var lineCount int
	for {
		_, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}
		lineCount++
	}
	return lineCount, nil
}

func IsValidPath(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		zap.L().Error(fmt.Sprintf("Path '%s' does not exist or is not accessible.", path))
		return false
	}

	// TODO [LP]: Is there any way to check read/write permissions? If path is a directory, we will need
	//  read/write permissions. Whereas, only read permission would suffice for a file.

	return true
}

func Contains[T comparable](value T, arr []T) bool {
	for i := 0; i < len(arr); i++ {
		if arr[i] == value {
			return true
		}
	}
	return false
}

func Find[T comparable](arr []T, value T) *T {
	for _, v := range arr {
		if v == value {
			return &v
		}
	}
	return nil
}
