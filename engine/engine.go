package engine

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

func Put(path, key string, val []byte) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}

	info, err := file.Stat()

	if err != nil {
		return fmt.Errorf("error getting stat of the file: %v", err)
	}

	if info.Size() > 4096 {

		fileName := fmt.Sprintf("%d.log", time.Now().UnixMilli())

		filePath := fmt.Sprintf("%s/%s", "/store/segments", fileName)

		return Put(filePath, key, val)
	}

	defer file.Close()

	bytes := []byte(key + "," + string(val) + "\n")

	n, err := file.Write(bytes)

	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	fmt.Printf("%d bytes are written. \n", n)

	return nil
}

func Get(path, key string) (string, error) {
	file, err := os.Open(path)

	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ",", 2)

		if len(parts) != 2 {
			continue // malformed line, skip
		}

		if parts[0] == key {
			return parts[1], nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}

	return "", fmt.Errorf("key not found: %s", key)
}

func GetAll(path string) (string, error) {
	file, err := os.Open(path)

	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}

	info, err := os.Stat(path)

	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}

	bytes := make([]byte, info.Size())

	file.Read(bytes)

	return string(bytes), nil
}
