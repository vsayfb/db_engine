package engine

import (
	"bufio"
	"db_engine/paths"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/oklog/ulid/v2"
)

var storage string = paths.GetPath("store/segments")

var currentSegmentFile string = filepath.Join(storage, NewSegmentName())

func Put(key string, val []byte) error {

	file, err := os.OpenFile(currentSegmentFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}

	info, err := os.Stat(file.Name())

	if err != nil {
		return fmt.Errorf("error stat file: %v", err)
	}

	/*
		Segmentation:

		- When the file reaches threshold create a new segment file and
		continue appending there. Make older segments immutable (frozen).

	*/
	if (info.Size() + (int64(len(key)) + int64(len(val)))) >= THRESHOLD {
		newSegmentName := NewSegmentName()

		absPath := filepath.Join(storage, newSegmentName)

		err = os.Chmod(currentSegmentFile, 0444)

		if err != nil {
			return fmt.Errorf("error making file read-only: %v", err)
		}

		currentSegmentFile = absPath

		return Put(key, val)
	}

	defer file.Close()

	bytes := []byte(key + "," + string(val) + "\n")

	_, err = file.Write(bytes)

	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	return nil
}

func Get(path, key string) (string, error) {
	file, err := os.Open(path)

	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ",", 2)

		if len(parts) != 2 {
			continue // malformed line, skip
		}

		if parts[0] == key {
			return parts[1], nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}

	return "", fmt.Errorf("key not found: %s", key)
}

func GetAll(path string) (string, error) {
	file, err := os.Open(path)

	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}

	info, err := os.Stat(path)

	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}

	bytes := make([]byte, info.Size())

	file.Read(bytes)

	return string(bytes), nil
}

func NewSegmentName() string {
	return ulid.Make().String() + ".log"
}
