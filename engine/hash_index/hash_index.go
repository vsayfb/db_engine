package hashindex

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type HashIndex struct {
	indexes map[string]map[string]int64
}

func NewHashIndex() *HashIndex {
	return &HashIndex{indexes: make(map[string]map[string]int64)}
}

func (index *HashIndex) GetIndexOfSegment(key string) map[string]int64 {
	return index.indexes[key]
}

func (index *HashIndex) CreateIndexForSegment(path string) error {

	file, err := os.Open(path)

	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}

	scanner := bufio.NewScanner(file)

	offset := 0

	for scanner.Scan() {
		line := scanner.Text()

		parts := strings.SplitN(line, ",", 2)

		if len(parts) != 2 {
			continue // malformed line, skip
		}

		index.indexes[path][parts[0]] = int64(offset)

		offset = offset + len(line) + 1
	}

	return nil
}

func (index *HashIndex) GetOffsetOfKey(path, key string) int64 {

	val, exist := index.indexes[path][key]

	if !exist {
		return -1
	}

	return val
}
