package root

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

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
	BorderForeground(lipgloss.Color("0")).
	PaddingRight(2)

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

// list.Item type
type Issue struct {
	data *linear.Issue
}

func (i Issue) Title() string       { return i.data.Identifier }
func (i Issue) Description() string { return i.data.Title }
func (i Issue) FilterValue() string { return i.Title() + i.Description() }

func splitIntoChunks(inputString string, chunkSize int) []string {
	words := strings.Fields(inputString) // Split the input into words
	chunks := []string{}
	currentChunk := ""

	for _, word := range words {
		// If adding the current word to the current chunk would exceed the chunk size,
		// add the current chunk to the list of chunks and start a new chunk.
		if len(currentChunk)+len(word)+1 > chunkSize {
			chunks = append(chunks, strings.TrimSpace(currentChunk))
			currentChunk = ""
		}

		// Add the word to the current chunk (with a space if not empty).
		if currentChunk != "" {
			currentChunk += " "
		}
		currentChunk += word
	}

	// Add the last chunk (if any).
	if currentChunk != "" {
		chunks = append(chunks, strings.TrimSpace(currentChunk))
	}

	return chunks
}

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 3 }
func (d itemDelegate) Spacing() int                              { return 1 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	borderStyle := lipgloss.NormalBorder()

	baseStyle := lipgloss.NewStyle().
		MarginLeft(2).
		PaddingLeft(1).
		BorderLeft(true).
		BorderStyle(borderStyle)

	titleStyle := baseStyle.Copy().Foreground(lipgloss.Color("#ffffff"))
	descriptionStyle := baseStyle.Copy().Foreground(lipgloss.Color("#808080"))

	i, ok := listItem.(Issue)
	if !ok {
		return
	}

	var title, description string
	description = i.data.Title

	selected := m.Cursor() == index%m.Paginator.PerPage

	if selected {
		title = titleStyle.
			Foreground(linearPurple).
			BorderLeftForeground(linearPurple).
			Render(i.data.Identifier)
	} else {
		title = titleStyle.Render(i.data.Identifier)
	}

	fmt.Fprintf(w, title)

	chunks := splitIntoChunks(description, 30)
	for i, chunk := range chunks {
		if selected {
			chunk = descriptionStyle.
				Foreground(linearPurple).
				BorderLeftForeground(linearPurple).
				Render(chunk)
		} else {
			chunk = descriptionStyle.Render(chunk)
		}

		fmt.Fprintf(w, "\n%s", chunk)

		if i >= 1 && len(chunks) > 2 {
			fmt.Fprintf(w, lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")).Render("…"))
			break
		}
	}
}

type model struct {
	// Models
	list      list.Model
	issueView viewport.Model
	help      help.Model

	pulls pulls

	// Helpers
	keys tui.KeyMap

	activePane pane

	gqlClient linear.GqlClient
	loading   bool
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

func (m model) GetSelectedIssue() *linear.Issue {
	selectedItem := m.list.SelectedItem()
	if selectedItem == nil {
		return &linear.Issue{}
	}

	return selectedItem.(Issue).data
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	issue := m.GetSelectedIssue()

	if m.pulls.selecting {
		m.pulls, cmd = m.pulls.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.P):
			attachments := issue.Attachments

			if len(attachments) == 1 {
				util.OpenURL(attachments[0].Url)
			} else {
				m.pulls.UpdateList(attachments)
				m.pulls.selecting = true
			}
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
			// Ignore if user is filtering
			if m.list.SettingFilter() {
				break
			}

			// TODO: handle multiple branches (based on issue attachments)
			err := git.CheckoutBranch(issue.BranchName)
			if err != nil {
				fmt.Printf("%s", err)
			}
			return m, tea.Quit
		case key.Matches(msg, m.keys.CtrlR):
			return m, m.refresh()
		case key.Matches(msg, m.keys.Enter):
			// Ignore if user is filtering
			if m.list.SettingFilter() {
				break
			}

			util.OpenURL(issue.Url)
			break
		}
	case tea.WindowSizeMsg:
		h, v := listStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-2)
	}

	m.list, cmd = m.list.Update(msg)
	m.issueView.Update(msg)
	return m, cmd
}

func issuesToItems(issues []*linear.Issue) []list.Item {
	items := []list.Item{}

	for _, issue := range issues {
		items = append(items, Issue{
			data: issue,
		})
	}

	return items
}

func (m *model) updateList(issues []*linear.Issue) tea.Cmd {
	return m.list.SetItems(issuesToItems(issues))
}

func (m *model) updateIssueView(issue *linear.Issue) error {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(glamour.DraculaStyleConfig),
		glamour.WithWordWrap(m.issueView.Width-2),
	)
	if err != nil {
		return err
	}

	content := fmt.Sprintf("# %s\n\n%s", issue.Title, issue.Description)

	str, err := renderer.Render(content)
	if err != nil {
		return err
	}

	m.issueView.SetContent(str)
	m.issueView.GotoTop()
	return nil
}

func (m *model) refresh() (cmd tea.Cmd) {
	issuesAsync := make(chan []*linear.Issue, 1)
	go func() {
		i, err := linear.GetIssues(m.gqlClient)
		if err != nil {
			fmt.Printf("Error retrieving issues: %v", err)
		}
		issuesAsync <- i
	}()
	issues := <-issuesAsync

	cmd = m.updateList(issues)
	if len(issues) > 0 {
		m.updateIssueView(issues[0])
	}

	return cmd
}

func (m model) View() string {
	help := m.help.ShortHelpView(m.keys.ShortHelp())

	render := lipgloss.JoinHorizontal(
		0.4,
		listStyle.Render(m.list.View()),
		m.issueView.View(),
	) + "\n" + helpStyle.Render(help)

	if m.pulls.selecting {
		style := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(linearPurple)
		render = tui.PlaceOverlay(20, 15, style.Render(m.pulls.View()), render, false)
	}

	return render
}

var rootCmd = &cobra.Command{
	Use:   "lin",
	Short: "lin is a CLI tool to interact with Linear",
	Run: func(cmd *cobra.Command, args []string) {
		if config.GetConfig().APIKey == "" {
			fmt.Println("Please run the 'auth' subcommand to add your Linear API key.")
			return
		}

		issues, needRefresh, err := linear.LoadIssues(linear.GetClient())
		if err != nil {
			log.Fatalf("Failed to open cache file: %v", err)
		}

		delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
			Foreground(linearPurple).BorderLeftForeground(linearPurple)

		delegate.Styles.SelectedDesc = delegate.Styles.SelectedTitle

		selectedItemStyle = delegate.Styles.SelectedTitle

		delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.MaxWidth(30)

		pulls := pulls{}
		pulls.Init()

		m := model{
			list:      list.New([]list.Item{}, itemDelegate{}, 0, 0),
			pulls:     pulls,
			keys:      tui.Keys,
			issueView: viewport.New(issueViewWidth, 50),
			gqlClient: linear.GetClient(),
			help:      help.New(),
			loading:   needRefresh,
		}
		m.help.ShortSeparator = " • "

		m.issueView.Style = contentStyle

		m.list.AdditionalShortHelpKeys = func() []key.Binding {
			return m.keys.ShortHelp()
		}
		m.list.Title = "Assigned Issues"
		m.list.Styles.Title = m.list.Styles.Title.Background(linearPurple)
		m.list.SetShowHelp(false)
		m.list.SetShowStatusBar(false)

		if len(issues) > 0 {
			m.updateList(issues)
			m.updateIssueView(issues[0])
		}

		// TODO: make this non-blocking
		if needRefresh || len(issues) == 0 {
			m.refresh()
		}

		p := tea.NewProgram(m, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
