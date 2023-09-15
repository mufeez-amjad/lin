package cmd

import (
	"lin_cli/internal/config"
	"log"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(authCmd)
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Configure credentials to connect to Linear's API",
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(config.InitialModel())
		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}
	},
}
