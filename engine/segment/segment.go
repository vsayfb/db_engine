package segment

import (
	"db_engine/engine/disk"
	"db_engine/engine/format"
	hashindex "db_engine/engine/hash_index"
	"db_engine/paths"
	"encoding/binary"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

var storage string = paths.GetPath("store/segments")

var writableSegmentFile string = path.Join(storage, "write.bin")

func Put(key []byte, val []byte) (int64, error) {

	info, err := os.Stat(writableSegmentFile)

	if os.IsNotExist(err) {
		file, err := os.Create(writableSegmentFile)

		if err != nil {
			return -1, fmt.Errorf("error creating new segment file: %v", err)
		}

		stat, err := file.Stat()

		if err != nil {
			return -1, fmt.Errorf("error getting stat of segment file: %v", err)
		}

		info = stat

		defer file.Close()
	} else if err != nil {
		return -1, fmt.Errorf("error stat file: %v", err)
	} else if (info.Size() + (int64(len(key)) + int64(len(val)))) >= THRESHOLD {
		/*
			Segmentation:

			- When the file reaches threshold create a new segment file and
			continue appending there. Make the current writable segment file immutable (frozen).
		*/
		segmentName := filepath.Join(storage, newSegmentFileName())

		err := os.Chmod(writableSegmentFile, 0444)

		if err != nil {
			return -1, fmt.Errorf("error making current writable segment read-only: %v", err)
		}

		if err := os.Rename(writableSegmentFile, segmentName); err != nil {
			return -1, fmt.Errorf("error renaming writable segment file: %v", err)
		}

		return Put(key, val)
	}

	_, err = disk.AppendFile(writableSegmentFile, format.EncodeBinary(key, val))

	if err != nil {
		return -1, fmt.Errorf("error writing to disk: %v", err)
	}

	offset := 0

	if info.Size() != 0 {
		offset = int(info.Size()) + 1
	}

	hashindex.NewHashIndex().IndexKey(writableSegmentFile, string(key), int64(offset))

	return info.Size(), nil
}

func GetValueByKey(key []byte) ([]byte, error) {

	filepath, offset := hashindex.NewHashIndex().GetOffsetOfData(string(key))

	if offset != -1 {
		return nil, fmt.Errorf("value not found by key: %s", key)
	}

	file, err := os.Open(filepath)

	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}

	defer file.Close()

	valueLen := make([]byte, 4)

	_, err = file.ReadAt(valueLen, offset+4)

	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	valueLenInt := binary.BigEndian.Uint32(valueLen)

	value := make([]byte, valueLenInt)

	_, err = file.ReadAt(value, offset+8+int64(len(key)))

	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return value, nil
}
