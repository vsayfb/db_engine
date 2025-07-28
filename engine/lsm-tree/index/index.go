package index

import (
	"bufio"
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
