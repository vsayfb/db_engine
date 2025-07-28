package memtable

import (
	"bytes"
	"db_engine/engine/lsm-tree/sstable"
	"sync"

	"github.com/emirpasic/gods/trees/redblacktree"
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

	oldVal, found := memtable.tree.Get(key)

	if found {
		memtable.totalBytes -= (len(key) + len(oldVal.([]byte)))
	}

	memtable.tree.Put(key, val)
	memtable.totalBytes += (len(key) + len(val))

	if memtable.SizeBytes() >= THRESHOLD {

		sstable := sstable.Construct(memtable.tree)

		if err := sstable.Create(); err != nil {
			return err
		}

		memtable.totalBytes = 0
		memtable.tree = redblacktree.NewWith(byteComparator)
	}

	return nil
}

func (memtable *Memtable) SizeBytes() int {
	return memtable.totalBytes
}

func byteComparator(a, b any) int {
	return bytes.Compare(a.([]byte), b.([]byte))
}
