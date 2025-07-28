package index

type Index struct {
	key    []byte
	offset int64
}

func New(key []byte, offset int64) Index {
	return Index{
		key:    key,
		offset: offset,
	}
}
