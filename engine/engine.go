package engine

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func Put(key string, val []byte) error {
	file, err := os.OpenFile("data.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
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

func Get(key string) (string, error) {
	file, err := os.Open("data.log")

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

func GetAll() (string, error) {
	file, err := os.Open("data.log")

	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}

	info, err := os.Stat("data.log")

	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}

	bytes := make([]byte, info.Size())

	file.Read(bytes)

	return string(bytes), nil
}
