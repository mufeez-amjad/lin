package util

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
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
