package memtable

import (
	"bufio"
	"bytes"
	"db_engine/engine/format"
	"db_engine/engine/lsm-tree/block"
	"db_engine/paths"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/emirpasic/gods/trees/redblacktree"
	"github.com/oklog/ulid/v2"
)

type Memtable struct {
	tree       *redblacktree.Tree
	totalBytes int
}

const THRESHOLD = 1024 * 1024 * 4

var instance *Memtable
var once sync.Once

func GetInstance() *Memtable {
	once.Do(func() {
		instance = &Memtable{
			tree: redblacktree.NewWith(byteComparator),
		}
	})

	return instance
}

func (memtable *Memtable) Put(key, val []byte) error {

	bytes := len(key) + len(val)

	if (memtable.SizeBytes() + bytes) >= THRESHOLD {
		if err := memtable.flushDisk(); err != nil {
			return err
		}

		memtable.totalBytes = 0
	}

	memtable.tree = redblacktree.NewWith(byteComparator)

	oldVal, found := memtable.tree.Get(key)

	memtable.tree.Put(key, val)

	memtable.totalBytes += (len(key) + len(val))

	if found {
		memtable.totalBytes -= (len(key) + len(oldVal.([]byte)))
	}

	return nil
}

func (memtable *Memtable) SizeBytes() int {
	return memtable.totalBytes
}

func (memtable *Memtable) flushDisk() error {

	it := memtable.tree.Iterator()

	path := filepath.Join(paths.GetPath("store/sstable"), newSortableFileName())

	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_TRUNC, 0644)

	if err != nil {
		os.Remove(path)
		return fmt.Errorf("error creating file: %v", err)
	}

	writer := bufio.NewWriter(file)

	defer file.Close()

	block := block.New()

	offset := int64(0)

	for it.Next() {

		key := it.Key().([]byte)
		value := it.Value().([]byte)
		data := format.EncodeBinary(key, value)

		if block.IsEmpty() {
			block.SetFirstKey(key)
		}

		block.AddToBlock(data)

		if block.GetSize() >= block.GetThreshold() {
			n, err := writer.Write(block.GetBlock())

			if err != nil {
				return fmt.Errorf("error writing block: %v", err)
			}

			block.Reset()
			offset += int64(n)
			block.SetOffset(offset)
		}
	}

	// write final block
	if block.GetSize() > 0 {
		n, err := writer.Write(block.GetBlock())

		if err != nil {
			return fmt.Errorf("error writing final block: %v", err)
		}

		block.Reset()
		offset += int64(n)
		block.SetOffset(offset)
	}

	return nil
}

func byteComparator(a, b any) int {
	return bytes.Compare(a.([]byte), b.([]byte))
}

func newSortableFileName() string {
	return ulid.Make().String() + ".bin"
}
