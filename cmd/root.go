package cmd

import (
	"fmt"
	"os"

	"lin_cli/internal/git"
	"lin_cli/internal/linear"
	"lin_cli/internal/tui"
	"lin_cli/internal/util"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/spf13/cobra"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

// sessionState is used to track which model is focused
type sessionState uint

type Issue struct {
	data linear.Issue
}

func (i Issue) Title() string       { return i.data.Identifier }
func (i Issue) Description() string { return i.data.Title }
func (i Issue) Data() linear.Issue  { return i.data }
func (i Issue) FilterValue() string { return i.Title() + i.Description() }

type model struct {
	list list.Model
	keys tui.KeyMap

	gqlClient linear.GqlClient
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var issue linear.Issue

	selectedItem := m.list.SelectedItem()
	if selectedItem != nil {
		issue = selectedItem.(Issue).Data()
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.C):
			// TODO: handle multiple branches (based on issue attachments)
			err := git.CheckoutBranch(issue.BranchName)
			if err != nil {
				fmt.Printf("%s", err)
			}
			return m, tea.Quit
		case key.Matches(msg, m.keys.CtrlR):
			m.refresh()
		case key.Matches(msg, m.keys.Enter):
			util.OpenURL(issue.Url)
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func (m *model) updateList(issues []linear.Issue) {
	items := []list.Item{}

	for _, issue := range issues {
		fmt.Printf("%s\n", issue.Identifier)
		items = append(items, Issue{
			data: issue,
		})
	}

	fmt.Printf("setting items: %d", len(items))
	m.list.SetItems(items)
}

func (m *model) refresh() {
	issues := make(chan []linear.Issue, 1)
	go func() {
		i, err := linear.GetIssues(m.gqlClient)
		if err != nil {
			fmt.Printf("Error retrieving issues: %v", err)
		}
		fmt.Printf("Retrieved %d issues\n", len(i))
		issues <- i
	}()
	m.updateList(<-issues)
}

var rootCmd = &cobra.Command{
	Use:   "lin",
	Short: "lin is a CLI tool to interact with Linear",
	Run: func(cmd *cobra.Command, args []string) {
		/*
			issues, err := store.ReadObjectFromFile[linear.Issue]("./cache")
			if err != nil {
				log.Fatalf("Failed to open cache file: %v", err)
			}
		*/

		m := model{
			list:      list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
			keys:      tui.Keys,
			gqlClient: linear.GetClient(),
		}
		m.list.AdditionalShortHelpKeys = func() []key.Binding {
			return m.keys.ShortHelp()
		}
		m.list.Title = "Assigned Issues"

		// if len(issues) > 0 {
		//	m.updateList(issues)
		//} else {
		m.refresh()
		//}

		p := tea.NewProgram(m, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
