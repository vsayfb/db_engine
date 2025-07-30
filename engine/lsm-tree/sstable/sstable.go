package sstable

import (
	"bufio"
	"bytes"
	"db_engine/engine/format"
	"db_engine/engine/lsm-tree/block"
	"db_engine/engine/lsm-tree/index"
	"db_engine/paths"
	"encoding/binary"
	"fmt"
	"io"
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

func (sstable *SSTable) Search(key []byte) (value []byte, err error) {

	file, err := os.Open(sstable.path)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	indexOffset, err := getIndexOffset(file)

	if err != nil {
		return nil, fmt.Errorf("error getting index offset of segment")
	}

	_, err = file.Seek(indexOffset, 0)

	if err != nil {
		return nil, err
	}

	indexBlock := index.New(nil)

	for {
		keyLenBuf := make([]byte, 4)

		_, err := file.ReadAt(keyLenBuf, indexOffset)

		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		keyLen := binary.BigEndian.Uint32(keyLenBuf)

		key := make([]byte, keyLen)

		file.ReadAt(key, indexOffset+4)

		offsetBuff := make([]byte, 8)

		file.ReadAt(offsetBuff, indexOffset+int64(keyLen)+int64(len(key)))

		offset := binary.BigEndian.Uint64(offsetBuff)

		offsetLen := 8

		indexBlock.Append(index.Index{Key: key, Offset: int64(offset)})

		indexOffset += (int64(keyLen) + int64(len(key)) + int64(offsetLen))
	}

	found, offset := indexBlock.Search(key)

	if found && offset != -1 {

		_, val, _, _ := format.DecodeBinary(file, offset)

		return val, nil
	}

	if !found && offset == -1 {
		return nil, fmt.Errorf("key not found")
	}

	_, err = file.Seek(offset, 0)

	if err != nil {
		return nil, fmt.Errorf("error seeking at block %v", err)
	}

	for {

		currKey, val, nextOffset, err := format.DecodeBinary(file, offset)

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("error decoding block binary %v", err)
		}

		if bytes.Equal(currKey, key) {
			return val, nil
		}

		offset = nextOffset

	}

	return nil, fmt.Errorf("key not found")
}

func newSortableFileName() string {
	return ulid.Make().String() + ".bin"
}

func getIndexOffset(file *os.File) (int64, error) {

	footer := make([]byte, 8)

	_, err := file.Seek(-8, 0)

	if err != nil {
		return -1, err
	}

	_, err = file.Read(footer)

	if err != nil {
		return -1, err
	}

	return int64(binary.BigEndian.Uint64(footer)), nil
}
