package block

type IndexEntry struct {
	Key    []byte
	Offset int64
}

/*
+----------------+     <-- block 1
| key1,val1      |
| key2,val2      |
| ...            |
+----------------+     <-- block 2
| keyN,valN      |
| ...            |
+----------------+     <-- ... more blocks
*/

type Block struct {
	offset   int64
	size     int
	buffer   []byte
	firstKey []byte
}

func New() *Block {
	return &Block{
		offset:   0,
		size:     0,
		buffer:   make([]byte, 0),
		firstKey: make([]byte, 0),
	}
}

func (block *Block) addToBlock(data []byte) {
	block.buffer = append(block.buffer, data...)

	block.size += len(data)
}

func (block *Block) getBlock() []byte {
	return block.buffer
}

func (block *Block) getSize() int {
	return block.size
}

func (block *Block) getThreshold() int {
	return 1024 * 4
}
