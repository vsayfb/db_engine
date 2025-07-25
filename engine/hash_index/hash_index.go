package hashindex

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

type HashIndex struct {
	indexes map[string]map[string]int64
}

var instance *HashIndex
var once sync.Once

func NewHashIndex() *HashIndex {
	once.Do(func() {
		instance = &HashIndex{indexes: make(map[string]map[string]int64)}
	})

	return instance
}

func (index *HashIndex) IndexKey(segment, key string, offset int64) {
	index.indexes[segment][key] = offset
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
