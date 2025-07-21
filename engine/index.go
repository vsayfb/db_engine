package engine

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func BuildHashIndex(path string) (*map[int64]string, error) {

	var index map[int64]string = map[int64]string{}

	file, err := os.Open(path)

	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}

	scanner := bufio.NewScanner(file)

	offset := 0

	for scanner.Scan() {
		line := scanner.Text()

		parts := strings.SplitN(line, ",", 2)

		if len(parts) != 2 {
			continue // malformed line, skip
		}

		index[int64(offset)] = string(parts[0] + parts[1])

		offset = offset + len(line) + 1
	}

	return &index, nil
}

func GetByIndex(path string, offset int64) (string, error) {

	file, err := os.Open(path)

	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}

	defer file.Close()

	_, err = file.Seek(offset, 0) // moves the file pointer to the given offset.

	if err != nil {
		return "", fmt.Errorf("error seeking file: %v", err)
	}

	buf := make([]byte, 1024) // Read 1024 bytes at a time

	var result []byte

	for {
		n, err := file.Read(buf)

		if err != nil && err.Error() != "EOF" {
			return "", fmt.Errorf("error reading file: %v", err)
		}

		// Process each byte
		for i := range n {
			result = append(result, buf[i])

			if buf[i] == '\n' {
				return string(result), nil
			}
		}

		if err == nil && n == 0 {
			return "", fmt.Errorf("EOF reached without finding '\\n': %v", err)
		}
	}
}
