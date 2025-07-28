package index

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
)

type Index struct {
	Key    []byte
	Offset int64
}

type IndexBlock struct {
	indexes []Index
	writer  *bufio.Writer
}

func New(writer *bufio.Writer) IndexBlock {
	return IndexBlock{
		indexes: make([]Index, 0),
		writer:  writer,
	}
}

func (indexBlock *IndexBlock) Append(index Index) {
	indexBlock.indexes = append(indexBlock.indexes, index)
}

func (indexBlock *IndexBlock) AppendBlockIntoFile() error {

	for _, index := range indexBlock.indexes {
		if err := binary.Write(indexBlock.writer, binary.BigEndian, uint32(len(index.Key))); err != nil {
			return fmt.Errorf("writing index key length: %v", err)
		}
		if _, err := indexBlock.writer.Write(index.Key); err != nil {
			return fmt.Errorf("writing index key: %v", err)
		}
		if err := binary.Write(indexBlock.writer, binary.BigEndian, index.Offset); err != nil {
			return fmt.Errorf("writing index offset: %v", err)
		}
	}

	return nil
}

func (block *IndexBlock) Search(key []byte) (bool, int64) {
	lo, hi := 0, len(block.indexes)-1

	for lo <= hi {
		mid := (lo + hi) / 2
		cmp := bytes.Compare(block.indexes[mid].Key, key)

		if cmp == 0 {
			return true, block.indexes[mid].Offset
		} else if cmp < 0 {
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}

	if hi >= 0 {
		return false, block.indexes[hi].Offset
	}

	return false, -1
}
