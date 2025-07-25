package hashindex

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
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

	defer file.Close()

	var offset int64 = 0
	const headerSize int64 = 8

	for {
		keyLenBuf := make([]byte, 4)
		valLenBuf := make([]byte, 4)

		_, err := file.ReadAt(keyLenBuf, offset)

		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading key length: %v", err)
		}

		_, err = file.ReadAt(valLenBuf, offset+4)

		if err != nil {
			return fmt.Errorf("error reading value length: %v", err)
		}

		keyLen := binary.BigEndian.Uint32(keyLenBuf)
		valLen := binary.BigEndian.Uint32(valLenBuf)

		key := make([]byte, keyLen)
		val := make([]byte, valLen)

		_, err = file.ReadAt(key, offset+headerSize)
		if err != nil {
			return fmt.Errorf("error reading key: %v", err)
		}

		_, err = file.ReadAt(val, offset+headerSize+int64(keyLen))
		if err != nil {
			return fmt.Errorf("error reading value: %v", err)
		}

		index.IndexKey(path, string(key), offset)

		offset += headerSize + int64(keyLen) + int64(valLen)
	}

	return nil
}

func (index *HashIndex) GetOffsetOfKey(key string) (string, int64) {

	path := ""
	var offset int64 = -1

	for s, pairs := range index.indexes {

		for k, off := range pairs {
			if k == key {
				path = s
				offset = off
			}
		}
	}

	return path, int64(offset)
}

func (index *HashIndex) GetIndex() map[string]map[string]int64 {
	return index.indexes
}
