package main

import (
	"lin_cli/cmd"
	"lin_cli/internal/config"
)

func main() {
	// Read in globally-available config
	config.GetConfig()

	cmd.Execute()

	config.SaveConfig()
}
