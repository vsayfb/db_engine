package hashindex

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"
)

type HashIndex struct {
	indexes sync.Map // map[string]sync.Map -> segment => (key => offset)
}

var instance *HashIndex
var once sync.Once

func NewHashIndex() *HashIndex {
	once.Do(func() {
		instance = &HashIndex{}
	})
	return instance
}

func (hi *HashIndex) IndexKey(segment, key string, offset int64) {

	val, _ := hi.indexes.LoadOrStore(segment, &sync.Map{})

	segmentMap := val.(*sync.Map)

	segmentMap.Store(key, offset)
}

func (hi *HashIndex) GetIndexOfSegment(segment string) *sync.Map {
	val, ok := hi.indexes.Load(segment)

	if !ok {
		return nil
	}

	return val.(*sync.Map)
}

func (hi *HashIndex) CreateIndexForSegment(path string) error {
	file, err := os.Open(path)

	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}

	defer file.Close()

	var offset int64 = 0
	const headerSize int64 = 8

	val, _ := hi.indexes.LoadOrStore(path, &sync.Map{})
	segmentMap := val.(*sync.Map)

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
		_, err = file.ReadAt(key, offset+headerSize)
		if err != nil {
			return fmt.Errorf("error reading key: %v", err)
		}

		segmentMap.Store(string(key), offset)

		offset += headerSize + int64(keyLen) + int64(valLen)
	}

	return nil
}

func (hi *HashIndex) GetOffsetOfData(key string) (string, int64) {
	var foundPath string
	var foundOffset int64 = -1

	hi.indexes.Range(func(segKey, segVal any) bool {
		segment := segKey.(string)
		segmentMap := segVal.(*sync.Map)

		if offsetAny, ok := segmentMap.Load(key); ok {
			foundPath = segment
			foundOffset = offsetAny.(int64)
			return false
		}
		return true
	})

	return foundPath, foundOffset
}

func (hi *HashIndex) Dump() map[string]map[string]int64 {
	result := make(map[string]map[string]int64)

	hi.indexes.Range(func(segKey, segVal any) bool {
		segment := segKey.(string)
		segmentMap := segVal.(*sync.Map)

		inner := make(map[string]int64)

		segmentMap.Range(func(k, v any) bool {
			inner[k.(string)] = v.(int64)
			return true
		})

		result[segment] = inner

		return true
	})

	return result
}
