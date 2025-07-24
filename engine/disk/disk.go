package disk

import (
	"fmt"
	"os"
)

func AppendFile(path string, data []byte) (int, error) {

	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		return -1, fmt.Errorf("error opening file: %v", err)
	}

	defer file.Close()

	n, err := file.Write(data)

	if err != nil {
		return -1, fmt.Errorf("error writing to file: %v", err)
	}

	return n, nil
}

func ReadLine(path string, offset int64) ([]byte, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}

	defer file.Close()

	_, err = file.Seek(offset, 0)

	if err != nil {
		return nil, fmt.Errorf("error seeking file: %v", err)
	}

	buffer := make([]byte, 1024)

	var result []byte

	for {

		n, err := file.Read(buffer)

		if err != nil && err.Error() != "EOF" {
			return nil, fmt.Errorf("error reading file: %v", err)
		}

		for i := range n {
			result = append(result, buffer[i])

			if buffer[i] == '\n' {
				return result, nil
			}
		}

		if err == nil && n == 0 {
			return nil, fmt.Errorf("EOF reached without finding '\\n': %v", err)
		}
	}

}

func ReadAll(path string) ([]byte, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}

	info, err := os.Stat(path)

	if err != nil {
		return nil, fmt.Errorf("error getting stat of file: %v", err)
	}

	bytes := make([]byte, info.Size())

	file.Read(bytes)

	return bytes, nil
}
