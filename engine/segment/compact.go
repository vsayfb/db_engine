package segment

import (
	"bufio"
	"db_engine/engine/format"
	hashindex "db_engine/engine/hash_index"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type SegmentInfo struct {
	Name      string
	Path      string
	CreatedAt time.Time
}

// AN ATOMIC COMPACTION & MERGING OPERATION
func CompactSegment() error {
	source := filepath.Join("store", "segments")

	entries, err := os.ReadDir(source)
	if err != nil {
		return fmt.Errorf("error reading directory /segments: %v", err)
	}
	if len(entries) <= 2 {
		return nil
	}

	var segments []SegmentInfo
	for _, entry := range entries {
		if entry.IsDir() || strings.Contains(entry.Name(), "write") {
			continue
		}

		fullPath := filepath.Join(source, entry.Name())
		info, err := os.Stat(fullPath)
		if err != nil {
			return fmt.Errorf("error getting file stat: %v", err)
		}

		segments = append(segments, SegmentInfo{
			Name:      entry.Name(),
			Path:      fullPath,
			CreatedAt: info.ModTime(),
		})
	}

	// Sort oldest to newest (newest overwrites oldest)
	sort.Slice(segments, func(i, j int) bool {
		return segments[i].CreatedAt.Before(segments[j].CreatedAt)
	})

	latest := make(map[string][]byte)
	var readSegments []string

	for _, seg := range segments {
		file, err := os.Open(seg.Path)

		if err != nil {
			return fmt.Errorf("error opening segment %s: %v", seg.Name, err)
		}

		offset := int64(0)

		for {

			key, val, nextOffset, err := format.DecodeBinary(file, offset)

			if err != nil {

				if err == io.EOF {
					break
				}

				return err
			}

			latest[string(key)] = val

			offset = nextOffset
		}

		file.Close()
		readSegments = append(readSegments, seg.Path)
	}

	// Write latest key-values into new segments
	buffer := make(map[string][]byte)
	var bufferSize int

	for k, v := range latest {
		size := 8 + len(k) + len(v)
		if bufferSize+size >= THRESHOLD {
			if err := mergeSegments(source, buffer); err != nil {
				return fmt.Errorf("error writing segment: %v", err)
			}
			buffer = make(map[string][]byte)
			bufferSize = 0
		}
		buffer[k] = v
		bufferSize += size
	}

	if bufferSize > 0 {
		if err := mergeSegments(source, buffer); err != nil {
			return fmt.Errorf("error writing remaining segment: %v", err)
		}
	}

	if err := cleanUp(readSegments); err != nil {
		return fmt.Errorf("cleanup error: %v", err)
	}

	return nil
}

func mergeSegments(dir string, pairs map[string][]byte) error {

	if len(pairs) == 0 {
		return nil
	}

	tmpFile, err := os.CreateTemp(dir, "compact.*.bin")

	if err != nil {
		return fmt.Errorf("error creating temp file: %v", err)

	}

	tmpPath := tmpFile.Name()

	defer func() {
		if err != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
		}
	}()

	writer := bufio.NewWriter(tmpFile)

	for k, v := range pairs {
		if _, err := writer.Write(format.EncodeBinary([]byte(k), v)); err != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("error writing to temp file: %v", err)
		}
	}

	if err := writer.Flush(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("error flushing writer: %v", err)
	}
	if err := tmpFile.Sync(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("error syncing file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("error closing temp file: %v", err)
	}

	newSegment := path.Join(dir, newSegmentFileName())

	if err := os.Rename(tmpPath, newSegment); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("error renaming temp file: %v", err)
	}

	if err := hashindex.NewHashIndex().CreateIndexForSegment(newSegment); err != nil {
		os.Remove(tmpPath)
		os.Remove(newSegment)
		return fmt.Errorf("error creating index for segment: %v", err)
	}

	return nil
}

func cleanUp(segments []string) error {
	for _, seg := range segments {
		if err := os.Remove(seg); err != nil && !os.IsNotExist(err) {
			fmt.Printf("warning: could not remove old segment %s: %v\n", seg, err)
		}
	}

	return nil
}
