package cmd

import (
	"fmt"
	"io"
	"log"
	"os"

	"lin_cli/internal/git"
	"lin_cli/internal/linear"
	"lin_cli/internal/tui/styles"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

type checkout struct {
	list list.Model

	choice   *linear.Issue
	quitting bool
}

type checkoutItemDelegate struct{}

func (d checkoutItemDelegate) Height() int                             { return 1 }
func (d checkoutItemDelegate) Spacing() int                            { return 0 }
func (d checkoutItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d checkoutItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	style := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(styles.LinearPurple)).
		Foreground(lipgloss.Color("#ffffff"))

	i, ok := listItem.(Issue)
	if !ok {
		return
	}

	pageItems := m.VisibleItems()

	longestIdentifier := 0
	for _, item := range pageItems {
		longestIdentifier = max(longestIdentifier, len(item.(Issue).data.Identifier))
	}

	padding := longestIdentifier - len(i.data.Identifier) + 1
	choice := i.data.Identifier
	for i := 0; i < padding; i++ {
		choice += " "
	}
	choice += ": " + i.data.BranchName

	if index == m.Index() {
		fmt.Fprint(w, style.Render("› "+choice))
	} else {
		fmt.Fprint(w, "  "+choice)
	}
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
			if !m.list.SettingFilter() {
				m.choice = m.list.SelectedItem().(Issue).data
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m checkout) View() string {
	return m.list.View()
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

		list := list.New(issuesToItems(issues), checkoutItemDelegate{}, 0, len(issues))
		list.Paginator.Type = paginator.Dots
		list.Paginator.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
		list.Paginator.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•")
		list.Title = "Select a branch to checkout:"
		list.SetShowHelp(false)
		list.SetShowStatusBar(false)
		list.Styles.TitleBar = lipgloss.NewStyle().PaddingBottom(1).PaddingTop(1)
		list.Styles.Title = lipgloss.NewStyle().Padding(0)

		p := tea.NewProgram(checkout{
			list:     list,
			choice:   nil,
			quitting: false,
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
