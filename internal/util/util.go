package util

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func OpenURL(href string) error {
	var cmd *exec.Cmd

	// Determine the operating system and open the URL accordingly
	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.Command("open", href)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", href)
	case "linux":
		cmd = exec.Command("xdg-open", href)
	default:
		return fmt.Errorf("unsupported platform")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func SplitIntoChunks(inputString string, chunkSize int) []string {
	words := strings.Fields(inputString) // Split the input into words
	chunks := []string{}
	currentChunk := ""

	for _, word := range words {
		// If adding the current word to the current chunk would exceed the chunk size,
		// add the current chunk to the list of chunks and start a new chunk.
		if len(currentChunk)+len(word)+1 > chunkSize {
			chunks = append(chunks, strings.TrimSpace(currentChunk))
			currentChunk = ""
		}

		// Add the word to the current chunk (with a space if not empty).
		if currentChunk != "" {
			currentChunk += " "
		}
		currentChunk += word
	}

	// Add the last chunk (if any).
	if currentChunk != "" {
		chunks = append(chunks, strings.TrimSpace(currentChunk))
	}

	return chunks
}
