package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"lin_cli/cmd/root/pulls"
	"lin_cli/internal/config"
	"lin_cli/internal/git"
	"lin_cli/internal/linear"
	"lin_cli/internal/tui"
	"lin_cli/internal/tui/styles"
	"lin_cli/internal/util"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/spf13/cobra"
)

var (
	issueViewWidth = 65
	listStyle      = lipgloss.NewStyle().
			Border(lipgloss.HiddenBorder()).
			Margin(2, 1).
			Width(100 - issueViewWidth)

	contentStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("0")).
			PaddingRight(2)

	spinnerStyle = lipgloss.NewStyle().Foreground(styles.Primary)

	selectedItemStyle lipgloss.Style

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).MarginBottom(1).MarginLeft(1)

	delegate = list.NewDefaultDelegate()
)

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

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 3 }
func (d itemDelegate) Spacing() int                              { return 1 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	baseStyle := lipgloss.NewStyle().
		PaddingLeft(1).
		BorderLeft(true).
		BorderStyle(lipgloss.HiddenBorder())

	titleStyle := baseStyle.Copy().
		Foreground(styles.Secondary).UnsetBorderLeftForeground()
	descriptionStyle := baseStyle.Copy().
		Foreground(styles.Tertiary).UnsetBorderLeftForeground()

	i, ok := listItem.(Issue)
	if !ok {
		return
	}

	var title, description string
	description = i.data.Title

	selected := m.Cursor() == index%m.Paginator.PerPage

	borderStyle := lipgloss.NormalBorder()

	if selected {
		title = titleStyle.
			Copy().
			Foreground(styles.Primary).
			BorderLeftForeground(styles.Primary).
			BorderStyle(borderStyle).
			Render(i.data.Identifier)
	} else {
		title = titleStyle.Render(i.data.Identifier)
	}

	fmt.Fprintf(w, title)

	chunks := util.SplitIntoChunks(description, 30)
	for i, chunk := range chunks {
		if selected {
			chunk = descriptionStyle.
				Copy().
				Foreground(styles.Primary).
				BorderLeftForeground(styles.Primary).
				BorderStyle(borderStyle).
				Render(chunk)
		} else {
			chunk = descriptionStyle.Render(chunk)
		}

		fmt.Fprintf(w, "\n%s", chunk)

		if i >= 1 && len(chunks) > 2 {
			if selected {
				style := lipgloss.NewStyle().Foreground(styles.Primary)
				fmt.Fprintf(w, style.Render("…"))
			} else {
				style := lipgloss.NewStyle().Foreground(styles.Tertiary)
				fmt.Fprintf(w, style.Render("…"))
			}
			break
		}
	}
}

type model struct {
	// Models
	list      list.Model
	issueView viewport.Model
	help      help.Model

	pulls pulls.PullsModel

	// Helpers
	keys tui.KeyMap

	activePane pane

	gqlClient linear.GqlClient
	loading   bool
	spinner   spinner.Model
	updateNum int

	// Data
	org *linear.Organization
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *model) updatePane() {
	if m.activePane == contentPane {
		m.activePane = listPane
		m.issueView.Style = contentStyle

		delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
			Foreground(styles.LinearPurple).BorderLeftForeground(styles.LinearPurple)
	} else {
		m.activePane = contentPane

		m.issueView.Style = m.issueView.Style.Copy().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(styles.Primary)

		delegate.Styles.SelectedTitle = selectedItemStyle.Copy().
			Foreground(styles.Grey).
			BorderLeftForeground(styles.Grey)
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

func (m *model) UpdateKeyBindings() {
	keys := []*key.Binding{
		&m.keys.P,
		&m.keys.Tab,
		&m.keys.C,
	}

	filtering := m.list.SettingFilter()
	for _, key := range keys {
		key.SetEnabled(!filtering)
	}
}

func (m *model) HandleMsg(msg tea.Msg) (*model, tea.Cmd) {
	issue := m.GetSelectedIssue()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.P):
			attachments := issue.Attachments

			if len(attachments) == 1 {
				util.OpenURL(attachments[0].Url)
			} else {
				m.pulls.UpdateList(attachments)
				m.pulls.Selecting = true
			}
			return m, nil
		case key.Matches(msg, m.keys.Tab):
			m.updatePane()
		case key.Matches(msg, m.keys.Up):
			if m.activePane == contentPane {
				var cmd tea.Cmd
				m.issueView, cmd = m.issueView.Update(msg)
				return m, cmd
			}

			idx := m.list.Index()
			items := m.list.VisibleItems()

			if idx == 0 || len(items) == 0 {
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
			items := m.list.VisibleItems()

			if idx == len(items)-1 || len(items) == 0 {
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
			return m, m.refreshIssues()
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
		m.issueView.Height = msg.Height - 10
	}

	var cmd tea.Cmd
	if m.activePane == contentPane {
		m.issueView.Update(msg)
	} else {
		m.list, cmd = m.list.Update(msg)
	}

	return m, cmd
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.loading {
		switch msg := msg.(type) {
		case spinner.TickMsg:
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	m.UpdateKeyBindings()

	if m.pulls.Selecting {
		m.pulls, cmd = m.pulls.Update(msg)
		return m, cmd
	}

	m, cmd = m.HandleMsg(msg)

	// Update the content pane
	issue := m.GetSelectedIssue()
	m.updateIssueView(issue)

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

func (m *model) refreshIssues() (cmd tea.Cmd) {
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

func (m *model) refreshOrg() {
	org, err := linear.GetOrganization(m.gqlClient)
	if err != nil {
		fmt.Printf("Error retrieving issues: %v", err)
	}
	m.org = org
}

func (m model) View() string {
	if m.loading {
		style := lipgloss.NewStyle().PaddingBottom(1)
		return style.Render(m.spinner.View() + " Loading...")
	}

	help := m.help.ShortHelpView(m.keys.ShortHelp())

	render := lipgloss.JoinHorizontal(
		0.4,
		listStyle.Render(m.list.View()),
		m.issueView.View(),
	) + "\n" + helpStyle.Render(help)

	if m.pulls.Selecting {
		style := lipgloss.NewStyle().
			Background(styles.OverlayBG).Padding(1).PaddingBottom(0)
		render = tui.PlaceOverlay(0, 0, style.Render(m.pulls.View()), render, false)
	}

	return render
}

var rootCmd = &cobra.Command{
	Use:   "lin",
	Short: "lin is a CLI tool to interact with Linear",
	Run: func(cmd *cobra.Command, args []string) {
		org, needRefreshOrg, err := linear.LoadOrg()
		if err != nil {
			log.Fatalf("Failed to open cache file: %v", err)
		}

		issues, needRefreshIssues, err := linear.LoadIssues(linear.GetClient())
		if err != nil {
			log.Fatalf("Failed to open cache file: %v", err)
		}

		delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Copy().
			Foreground(styles.LinearPurple).BorderLeftForeground(styles.LinearPurple)

		delegate.Styles.SelectedDesc = delegate.Styles.SelectedTitle

		selectedItemStyle = delegate.Styles.SelectedTitle

		delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.MaxWidth(30)

		pulls := pulls.PullsModel{}
		pulls.Init()

		m := model{
			list:      list.New([]list.Item{}, itemDelegate{}, 0, 0),
			pulls:     pulls,
			keys:      tui.Keys,
			issueView: viewport.New(issueViewWidth, 80),
			gqlClient: linear.GetClient(),
			help:      help.New(),
			loading:   needRefreshIssues || needRefreshOrg,
			spinner:   spinner.New(),
			org:       org,
		}
		m.help.ShortSeparator = " • "
		m.spinner.Spinner = spinner.MiniDot
		m.spinner.Style = spinnerStyle

		m.issueView.Style = contentStyle

		m.list.AdditionalShortHelpKeys = func() []key.Binding {
			return m.keys.ShortHelp()
		}
		m.list.Title = "Assigned Issues"
		m.list.Styles.Title = m.list.Styles.Title.
			Background(styles.Primary)
		m.list.SetShowHelp(false)
		m.list.SetShowStatusBar(false)

		if len(issues) > 0 {
			m.updateList(issues)
			m.updateIssueView(issues[0])
		}

		var p *tea.Program
		if needRefreshIssues || needRefreshOrg {
			p = tea.NewProgram(&m)
			go func() {
				startTime := time.Now()

				// TODO: make this non-blocking
				if needRefreshIssues {
					m.refreshIssues()
				}
				if needRefreshOrg {
					m.refreshOrg()
				}
				loadingFor := time.Now().Sub(startTime)
				// Show spinner for at least 1 second
				if loadingFor < time.Second {
					time.Sleep(time.Second - loadingFor)
				}
				m.loading = false
				// TODO: clear Loading message
				p.Send(tea.EnterAltScreen())
			}()
		} else {
			p = tea.NewProgram(&m, tea.WithAltScreen())
		}

		if _, err := p.Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}

	},
}

func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	if config.GetConfig().APIKey == "" {
		authCmd.Run(nil, nil)
		if config.GetConfig().APIKey == "" {
			// Exited without entering API key
			return
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
