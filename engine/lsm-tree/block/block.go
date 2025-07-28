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

func (block *Block) AddToBlock(data []byte) {
	block.buffer = append(block.buffer, data...)

	block.size += len(data)
}

func (block *Block) GetBlock() []byte {
	return block.buffer
}

func (block *Block) GetSize() int {
	return block.size
}

func (block *Block) GetThreshold() int {
	return 1024 * 16
}

func (block *Block) IsEmpty() bool {
	return block.size == 0
}

func (block *Block) SetFirstKey(key []byte) {
	block.firstKey = key
}

func (block *Block) SetOffset(offset int64) {
	block.offset = offset
}

func (block *Block) Reset() {
	block.buffer = make([]byte, 0)
	block.firstKey = make([]byte, 0)
	block.size = 0
	block.offset = 0
}
