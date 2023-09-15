package cmd

import (
	"fmt"
	"os"

	"lin_cli/internal/config"
	"lin_cli/internal/git"
	"lin_cli/internal/linear"
	"lin_cli/internal/tui"
	"lin_cli/internal/util"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/spf13/cobra"
)

var issueViewWidth = 65
var listStyle = lipgloss.NewStyle().
	Border(lipgloss.HiddenBorder()).
	Margin(2, 2).
	Width(100 - issueViewWidth)

var contentStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("0"))

var selectedItemStyle lipgloss.Style

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).MarginBottom(1)

var linearPurple = lipgloss.Color("#5e63d7")
var linearPurpleDarker = lipgloss.Color("#494b7b")

var delegate = list.NewDefaultDelegate()

type pane int

const (
	listPane pane = iota
	contentPane
)

type Issue struct {
	data linear.Issue
}

func (i Issue) Title() string       { return i.data.Identifier }
func (i Issue) Description() string { return i.data.Title }
func (i Issue) Data() linear.Issue  { return i.data }
func (i Issue) FilterValue() string { return i.Title() + i.Description() }

type model struct {
	list       list.Model
	keys       tui.KeyMap
	issueView  viewport.Model
	help       help.Model
	activePane pane

	gqlClient linear.GqlClient
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) updatePane() {
	if m.activePane == contentPane {
		m.activePane = listPane
		m.issueView.Style = contentStyle

		delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
			Foreground(linearPurple).BorderLeftForeground(linearPurple)
	} else {
		m.activePane = contentPane

		m.issueView.Style = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(linearPurple)

		delegate.Styles.SelectedTitle = selectedItemStyle.
			Foreground(lipgloss.Color("0")).
			BorderLeftForeground(lipgloss.Color("0"))
	}

	delegate.Styles.SelectedDesc = delegate.Styles.SelectedTitle
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
		case key.Matches(msg, m.keys.Tab):
			m.updatePane()
		case key.Matches(msg, m.keys.Up):
			if m.activePane == contentPane {
				var cmd tea.Cmd
				m.issueView, cmd = m.issueView.Update(msg)
				return m, cmd
			}

			idx := m.list.Index()
			items := m.list.Items()

			if idx == 0 {
				break
			}

			nextIdx := (idx - 1) % len(items)
			nextIssue := items[nextIdx].(Issue).data
			m.updateIssueView(nextIssue)
		case key.Matches(msg, m.keys.Down):
			if m.activePane == contentPane {
				var cmd tea.Cmd
				m.issueView, cmd = m.issueView.Update(msg)
				return m, cmd
			}

			idx := m.list.Index()
			items := m.list.Items()

			if idx == len(items)-1 {
				break
			}

			nextIdx := (idx + 1) % len(items)
			nextIssue := items[nextIdx].(Issue).data
			m.updateIssueView(nextIssue)
		case key.Matches(msg, m.keys.C):
			// TODO: handle multiple branches (based on issue attachments)
			err := git.CheckoutBranch(issue.BranchName)
			if err != nil {
				fmt.Printf("%s", err)
			}
			return m, tea.Quit
		case key.Matches(msg, m.keys.CtrlR):
			m.refresh()
			break
		case key.Matches(msg, m.keys.Enter):
			util.OpenURL(issue.Url)
			break
		}
	case tea.WindowSizeMsg:
		h, v := listStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	m.issueView.Update(msg)
	return m, cmd
}

func (m *model) updateList(issues []linear.Issue) {
	items := []list.Item{}

	for _, issue := range issues {
		items = append(items, Issue{
			data: issue,
		})
	}

	m.list.SetItems(items)
}

func (m *model) updateIssueView(issue linear.Issue) error {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(m.issueView.Width),
	)
	if err != nil {
		return err
	}

	str, err := renderer.Render(issue.Description)
	if err != nil {
		return err
	}

	m.issueView.SetContent(str)
	m.issueView.GotoTop()
	return nil
}

func (m *model) refresh() {
	issuesAsync := make(chan []linear.Issue, 1)
	go func() {
		i, err := linear.GetIssues(m.gqlClient)
		if err != nil {
			fmt.Printf("Error retrieving issues: %v", err)
		}
		issuesAsync <- i
	}()
	issues := <-issuesAsync
	m.updateList(issues)
	m.updateIssueView(issues[0])
}

func (m model) View() string {
	help := m.help.ShortHelpView(m.keys.ShortHelp())

	return lipgloss.JoinHorizontal(
		0.4,
		listStyle.Render(m.list.View()),
		m.issueView.View(),
	) + "\n" + helpStyle.Render(help)
}

var rootCmd = &cobra.Command{
	Use:   "lin",
	Short: "lin is a CLI tool to interact with Linear",
	Run: func(cmd *cobra.Command, args []string) {
		if config.GetConfig().APIKey == "" {
			fmt.Println("Please run the 'auth' subcommand to add your Linear API key.")
			return
		}

		/*
			issues, err := store.ReadObjectFromFile[linear.Issue]("./cache")
			if err != nil {
				log.Fatalf("Failed to open cache file: %v", err)
			}
		*/

		delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
			Foreground(linearPurple).BorderLeftForeground(linearPurple)

		delegate.Styles.SelectedDesc = delegate.Styles.SelectedTitle

		selectedItemStyle = delegate.Styles.SelectedTitle

		/*
			delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
				Foreground(linearPurple).BorderLeftForeground(linearPurple)
		*/

		m := model{
			list:      list.New([]list.Item{}, delegate, 0, 0),
			keys:      tui.Keys,
			issueView: viewport.New(issueViewWidth, 50),
			gqlClient: linear.GetClient(),
			help:      help.New(),
		}
		m.help.ShortSeparator = " • "

		m.issueView.Style = contentStyle

		m.list.AdditionalShortHelpKeys = func() []key.Binding {
			return m.keys.ShortHelp()
		}
		m.list.Title = "Assigned Issues"
		m.list.Styles.Title = m.list.Styles.Title.Background(linearPurple)
		m.list.SetShowHelp(false)

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
