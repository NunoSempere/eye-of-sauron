package main

import (
	"bufio"
	"os"
	"strings"
)

func ReadFileLines(filename string) ([]string, error) {
	var lines []string

	// Open the file for reading
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a new scanner to read the file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		clean := strings.ReplaceAll(line, "<b>", "")
		clean = strings.ReplaceAll(line, "</b>", "")
		clean = strings.ReplaceAll(line, "&#39;", "'")
		lines = append(lines, clean) // Append each line to the slice
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
