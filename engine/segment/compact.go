package segment

import (
	"bufio"
	"fmt"
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
		if entry.IsDir() {
			continue
		}

		fullPath := filepath.Join(source, entry.Name())

		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			return fmt.Errorf("error getting file stat: %v", err)
		}

		segments = append(segments, SegmentInfo{
			Name:      entry.Name(),
			Path:      fullPath,
			CreatedAt: fileInfo.ModTime(),
		})
	}

	// sort segments oldest to newest to keep most recent value for each key
	sort.Slice(segments, func(i, j int) bool {
		return segments[i].CreatedAt.Before(segments[j].CreatedAt)
	})

	var written int
	readSegments := make([]string, 0)
	pairs := make(map[string]string, 0)

	for _, s := range segments {

		// do not include writable segment file in process
		if strings.Contains(s.Name, "write") {
			continue
		}

		file, err := os.Open(path.Join(source, s.Name))

		if err != nil {
			return fmt.Errorf("error opening segment %s: %v", s.Name, err)
		}

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {

			line := scanner.Text()

			parts := strings.SplitN(line, ",", 2)

			if len(parts) != 2 {
				continue
			}

			key, value := parts[0], parts[1]

			// discard the length of bytes in older pairs
			if existing, exists := pairs[key]; exists {
				written -= len(key) + len(existing)
			}

			pairs[key] = value

			written += len(line)

			if written >= THRESHOLD {

				if err := mergeSegments(source, pairs); err != nil {

					file.Close()

					return err
				}

				written = 0
			}
		}

		if err := scanner.Err(); err != nil {
			file.Close()
			return fmt.Errorf("error reading segment %s: %v", s.Name, err)
		}

		readSegments = append(readSegments, path.Join(source, s.Name))

		file.Close()

	}

	// merge remaining segments
	if written > 0 {
		if err := mergeSegments(source, pairs); err != nil {
			return fmt.Errorf("error merging remaining segments %v", err)
		}
	}

	if err := cleanUp(readSegments); err != nil {
		return fmt.Errorf("error during segment clean up %v", err)
	}

	return nil
}

func mergeSegments(dir string, pairs map[string]string) error {
	if len(pairs) == 0 {
		return nil
	}

	tmpFile, err := os.CreateTemp(dir, "compact.*.tmp")

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
		if _, err := writer.WriteString(k + "," + v + "\n"); err != nil {
			return fmt.Errorf("error writing to temp file: %v", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("error flushing writer: %v", err)
	}
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("error syncing file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("error closing temp file: %v", err)
	}

	newSegment := path.Join(dir, newSegmentFileName())

	if err := os.Rename(tmpPath, newSegment); err != nil {
		return fmt.Errorf("error renaming temp file: %v", err)
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
