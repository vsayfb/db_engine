package compact

import (
	"bufio"
	"bytes"
	"db_engine/engine/format"
	"db_engine/engine/lsm-tree/block"
	"db_engine/engine/lsm-tree/index"
	"db_engine/paths"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/oklog/ulid/v2"
)

type SSTable struct {
	Name      string
	Path      string
	CreatedAt time.Time
}

func CompactSSTables() error {

	files, err := os.ReadDir(paths.GetPath("store/sstables"))

	if err != nil {
		return err
	}

	sstables, err := sortFiles(files)

	if err != nil {
		return err
	}

	oldest := sstables[0]

	for i := 1; i < len(sstables); i++ {

		curr := sstables[i]

		ss1, err := os.Open(oldest.Path)

		if err != nil {
			return fmt.Errorf("error opening sstable %v", err)
		}

		ss2, err := os.Open(curr.Path)

		if err != nil {
			return fmt.Errorf("error opening sstable %v", err)
		}

		mergedSSTable, err := mergeSSTables(ss1, ss2)

		if err != nil {
			return err
		}

		oldest = SSTable{Path: mergedSSTable}
	}

	return nil
}

func mergeSSTables(ss1, ss2 *os.File) (string, error) {
	l, r := int64(0), int64(0)

	lEnd, err := getIndexOffset(ss1)
	if err != nil {
		return "", err
	}
	rEnd, err := getIndexOffset(ss2)
	if err != nil {
		return "", err
	}

	tmpFile, err := os.CreateTemp(paths.GetPath("store/sstables"), "compact.*.bin")

	if err != nil {
		return "", err
	}

	block := block.New()
	writer := bufio.NewWriter(tmpFile)
	indexBlock := index.New(writer)
	var offset int64

	for l < (lEnd) && r < (rEnd) {

		key1, val1, nextOffset1, err := format.DecodeBinary(ss1, l)

		if err != nil {
			return "", err
		}

		key2, val2, nextOffset2, err := format.DecodeBinary(ss2, r)

		if err != nil {
			return "", err
		}

		cmp := bytes.Compare(key1, key2)

		var key []byte
		var val []byte

		switch cmp {
		case 0:
			key = key2
			val = val2
			r = nextOffset2
		case -1:
			key = key1
			val = val1
			l = nextOffset1
		default:
			key = key2
			val = val2
			r = nextOffset2
		}

		data := format.EncodeBinary(key, val)

		writer.Write(data)

		if block.IsEmpty() {
			block.SetFirstKey(key)
		}

		block.Append(data)

		if block.GetSize() >= block.GetThreshold() {
			n, _ := writer.Write(block.GetBlock())

			indexBlock.Append(index.Index{Key: block.GetFirstKey(), Offset: offset})

			block.Reset()
			offset += int64(n)
			block.SetOffset(offset)
		}
	}

	for l < lEnd {
		key, val, nextOffset, err := format.DecodeBinary(ss1, l)

		if err != nil {
			return "", err
		}

		n, _ := writer.Write(format.EncodeBinary(key, val))

		offset += int64(n)

		l = nextOffset
	}

	for r < rEnd {
		key, val, nextOffset, err := format.DecodeBinary(ss2, r)

		if err != nil {
			return "", err
		}

		n, _ := writer.Write(format.EncodeBinary(key, val))

		offset += int64(n)

		r = nextOffset
	}

	indexBlockOffset := offset

	if !block.IsEmpty() {
		n, err := writer.Write(block.GetBlock())

		if err != nil {
			return "", fmt.Errorf("error writing final block: %v", err)
		}

		indexBlock.Append(index.Index{Key: block.GetFirstKey(), Offset: offset})

		block.Reset()
		offset += int64(n)
		block.SetOffset(offset)
	}

	if err := indexBlock.AppendBlockIntoFile(); err != nil {
		return "", fmt.Errorf("error appending index block into file: %v", err)
	}

	if err := binary.Write(writer, binary.BigEndian, indexBlockOffset); err != nil {
		return "", fmt.Errorf("error writing footer block: %v", err)
	}

	if err := writer.Flush(); err != nil {
		return "", fmt.Errorf("error flushing to disk: %v", err)
	}

	ss1.Close()
	ss2.Close()

	os.Remove(ss1.Name())
	os.Remove(ss2.Name())

	newPath := filepath.Join(paths.GetPath("store/sstables"), (ulid.Make().String() + ".bin"))

	os.Rename(tmpFile.Name(), newPath)

	return newPath, nil
}

func sortFiles(files []os.DirEntry) ([]SSTable, error) {

	source := paths.GetPath("store/sstables")

	var sstables []SSTable

	for _, entry := range files {
		if entry.IsDir() {
			continue
		}

		fullPath := filepath.Join(source, entry.Name())

		info, err := os.Stat(fullPath)

		if err != nil {
			return nil, fmt.Errorf("error getting file stat: %v", err)
		}

		sstables = append(sstables, SSTable{
			Name:      entry.Name(),
			Path:      fullPath,
			CreatedAt: info.ModTime(),
		})
	}

	// Sort oldest to newest (newest overwrites oldest)
	sort.Slice(sstables, func(i, j int) bool {
		return sstables[i].CreatedAt.Before(sstables[j].CreatedAt)
	})

	return sstables, nil
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
