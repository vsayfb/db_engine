package main

import (
	"fmt"
	"os"
)

func main() {

	dirPath := "/store/segments"

	// Check if the directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// If the directory does not exist, create it
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}
	}

}
