package engine

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func CompactSegment(path string) error {
	file, err := os.Open(path)

	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}

	defer file.Close()

	pairs := make(map[string]string)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {

		line := scanner.Text()

		parts := strings.SplitN(line, ",", 2)

		if len(parts) == 2 {
			pairs[parts[0]] = parts[1]
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	dir := filepath.Dir(path)

	tmpFile, err := os.CreateTemp(dir, "compact.*.tmp")

	if err != nil {
		return fmt.Errorf("error creating temporary file: %v", err)
	}

	tmpPath := tmpFile.Name()

	writer := bufio.NewWriter(tmpFile)

	for k, v := range pairs {

		if _, err := writer.WriteString(k + "," + v + "\n"); err != nil {

			tmpFile.Close()

			os.Remove(tmpPath)

			return fmt.Errorf("error writing to temporary file: %v", err)
		}
	}

	if err := writer.Flush(); err != nil {
		tmpFile.Close()

		os.Remove(tmpPath)

		return fmt.Errorf("error flushing temporary file: %v", err)
	}

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()

		os.Remove(tmpPath)

		return fmt.Errorf("error syncing temporary file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)

		return fmt.Errorf("error closing temporary file: %v", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)

		return fmt.Errorf("error replacing original file: %v", err)
	}

	return nil
}
