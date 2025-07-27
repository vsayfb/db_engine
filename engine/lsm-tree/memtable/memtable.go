package memtable

import (
	"bytes"
	"sync"

	"github.com/emirpasic/gods/trees/redblacktree"
)

type Memtable struct {
	tree *redblacktree.Tree
}

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

func byteComparator(a, b any) int {
	return bytes.Compare(a.([]byte), b.([]byte))
}
