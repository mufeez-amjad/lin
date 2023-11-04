package util

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

func RightAlignText(leftText, rightText string, totalWidth int) string {
	leftLen := lipgloss.Width(leftText)
	rightLen := lipgloss.Width(rightText)
	if leftLen+rightLen > totalWidth {
		return leftText
	}

	leftWidth := totalWidth - rightLen

	// Create a strings.Builder to build the resulting string.
	var builder strings.Builder

	// Append leftText with spaces to fill the remaining width.
	builder.WriteString(leftText)
	spaces := leftWidth - leftLen
	for i := 0; i < spaces; i++ {
		builder.WriteByte(' ')
	}

	// Append rightText.
	builder.WriteString(rightText)

	return builder.String()
}
