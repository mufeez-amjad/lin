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
	Short: "Print the version number of Hugo",
	Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(config.InitialModel())
		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}
	},
}
