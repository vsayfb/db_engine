package sstable

import (
	"bufio"
	"db_engine/engine/format"
	"db_engine/engine/lsm-tree/block"
	"db_engine/engine/lsm-tree/index"
	"db_engine/paths"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"

	"github.com/emirpasic/gods/trees/redblacktree"
	"github.com/oklog/ulid/v2"
)

type SSTable struct {
	tree *redblacktree.Tree
	path string
}

func Construct(tree *redblacktree.Tree) *SSTable {
	return &SSTable{
		tree: tree,
	}
}

func New(path string) *SSTable {
	return &SSTable{
		path: path,
	}
}

func (sstable *SSTable) Create() error {

	path := filepath.Join(paths.GetPath("store/sstable"), newSortableFileName())

	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)

	if err != nil {
		os.Remove(path)
		return fmt.Errorf("error creating file: %v", err)
	}

	defer file.Close()

	writer := bufio.NewWriter(file)
	block := block.New()
	offset := int64(0)
	indexBlock := index.New(writer)

	it := sstable.tree.Iterator()

	for it.Next() {
		key := it.Key().([]byte)
		value := it.Value().([]byte)

		data := format.EncodeBinary(key, value)

		if block.IsEmpty() {
			block.SetFirstKey(key)
		}

		block.Append(data)

		if block.GetSize() >= block.GetThreshold() {
			n, err := writer.Write(block.GetBlock())

			if err != nil {
				return fmt.Errorf("error writing block: %v", err)
			}

			indexBlock.Append(index.Index{Key: block.GetFirstKey(), Offset: offset})

			block.Reset()
			offset += int64(n)
			block.SetOffset(offset)
		}
	}

	indexBlockOffset := offset

	// write final block
	if !block.IsEmpty() {
		n, err := writer.Write(block.GetBlock())

		if err != nil {
			return fmt.Errorf("error writing final block: %v", err)
		}

		indexBlock.Append(index.Index{Key: block.GetFirstKey(), Offset: offset})

		block.Reset()
		offset += int64(n)
		block.SetOffset(offset)
	}

	if err := indexBlock.AppendBlockIntoFile(); err != nil {
		return fmt.Errorf("error appending index block into file: %v", err)
	}

	// a pointer holds the beginning offset of the index block
	if err := binary.Write(writer, binary.BigEndian, indexBlockOffset); err != nil {
		return fmt.Errorf("error writing footer block: %v", err)
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("error flushing to disk: %v", err)
	}

	return nil
}

func newSortableFileName() string {
	return ulid.Make().String() + ".bin"
}
