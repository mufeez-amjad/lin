package main

import (
	rootCmd "lin_cli/cmd/root"
	"lin_cli/internal/config"
)

func main() {
	// Read in globally-available config
	config.GetConfig()

	rootCmd.Execute()

	config.SaveConfig()
}
