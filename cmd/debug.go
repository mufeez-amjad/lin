package cmd

import (
	"fmt"
	"lin_cli/internal/config"
	"lin_cli/internal/store"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(debugCmd)
}

var debugCmd = &cobra.Command{
	Use:    "debug",
	Short:  "Prints debugging information",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Cache directory: %s\n", store.CacheDirectory)
		fmt.Printf("Config directory: %s\n", config.ConfigDir)
	},
}
