package cmd

import (
	"fmt"
	"lin_cli/internal/git"
	"lin_cli/internal/linear"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// TODO: switch to using list with custom itemDelegate
type checkout struct {
	cursor  int
	choice  *linear.Issue
	choices []*linear.Issue

	paginator paginator.Model
	quitting  bool
}

func (m checkout) Init() tea.Cmd {
	return nil
}

func (m checkout) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			m.choice = m.choices[m.cursor]
			return m, tea.Quit

		case "left":
			m.paginator.PrevPage()
			start, _ := m.paginator.GetSliceBounds(len(m.choices))
			m.cursor = start

		case "right":
			m.paginator.NextPage()
			start, _ := m.paginator.GetSliceBounds(len(m.choices))
			m.cursor = start

		case "down", "j":
			_, end := m.paginator.GetSliceBounds(len(m.choices))

			if m.cursor < len(m.choices)-1 {
				m.cursor += 1

				if m.cursor >= end {
					m.paginator.NextPage()
				}
			}

		case "up", "k":
			start, _ := m.paginator.GetSliceBounds(len(m.choices))

			if m.cursor > 0 {
				m.cursor -= 1

				if m.cursor <= start {
					m.paginator.PrevPage()
				}
			}
		}
	}

	return m, nil
}

func (m checkout) View() string {
	s := strings.Builder{}
	s.WriteString("Select a branch to checkout:\n\n")

	style := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(linearPurple)).
		Foreground(lipgloss.Color("#ffffff"))

	start, end := m.paginator.GetSliceBounds(len(m.choices))
	pageItems := m.choices[start:end]

	longestIdentifier := 0
	for _, item := range pageItems {
		longestIdentifier = max(longestIdentifier, len(item.Identifier))
	}

	for i, item := range pageItems {
		padding := longestIdentifier - len(item.Identifier) + 1
		choice := item.Identifier
		for i := 0; i < padding; i++ {
			choice += " "
		}
		choice += ": " + item.BranchName

		if m.cursor == i+start {
			s.WriteString(style.Render("› "))
			s.WriteString(style.Render(choice))
		} else {
			s.WriteString("› ")
			s.WriteString(choice)
		}
		s.WriteString("\n")
	}

	s.WriteString("\n" + m.paginator.View() + "\n")
	s.WriteString("\n(press q to quit)\n")

	return s.String()
}

func init() {
	rootCmd.AddCommand(checkoutCmd)
}

var checkoutCmd = &cobra.Command{
	Use:     "checkout",
	Aliases: []string{"co"},
	Short:   "List available branches to checkout",
	Run: func(cmd *cobra.Command, args []string) {
		issues, _, err := linear.LoadIssues(linear.GetClient())
		if err != nil {
			log.Fatalf("Could not load issues: %v", err)
		}

		pg := paginator.New()
		pg.Type = paginator.Dots
		pg.PerPage = 10
		pg.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
		pg.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•")
		pg.SetTotalPages(len(issues))

		p := tea.NewProgram(checkout{
			choices:   issues,
			paginator: pg,
		})

		// Run returns the model as a tea.Model.
		m, err := p.Run()
		if err != nil {
			fmt.Println("Oh no:", err)
			os.Exit(1)
		}

		// Assert the final tea.Model to our local model and print the choice.
		if m, ok := m.(checkout); ok && m.choice != nil {
			git.CheckoutBranch(m.choice.BranchName)
		}
	},
}
