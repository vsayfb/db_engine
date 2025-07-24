package segment

import (
	"db_engine/engine/disk"
	"db_engine/engine/format"
	"db_engine/paths"
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

		file.Close()
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

	_, err = disk.AppendFile(writableSegmentFile, format.FormatBinary(key, val))

	if err != nil {
		return -1, fmt.Errorf("error writing to disk: %v", err)
	}

	return info.Size() + 1, nil
}
