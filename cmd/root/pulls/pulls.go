package pulls

import (
	"fmt"
	"io"
	"lin_cli/internal/linear"
	"lin_cli/internal/tui"
	"lin_cli/internal/tui/styles"
	"lin_cli/internal/util"
	"sort"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PullsModel struct {
	list      list.Model
	Selecting bool

	keys tui.KeyMap
}

type Attachment struct {
	title string
	data  *linear.Attachment
}

func (a Attachment) Title() string       { return a.title }
func (a Attachment) Description() string { return "" }
func (a Attachment) FilterValue() string { return a.Title() + a.Description() }

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 1 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Attachment)
	if !ok {
		return
	}

	getText := func(item Attachment, selected bool) string {
		str := fmt.Sprintf("%d. %s ", index+1, item.Title())
		if selected {
			return fmt.Sprintf("- %s", str)
		} else {
			return fmt.Sprintf("  %s", str)
		}
	}

	pageItems := m.VisibleItems()

	longestLine := 0
	for _, item := range pageItems {
		longestLine = max(longestLine, len(getText(item.(Attachment), false)))
	}

	str := getText(i, index == m.Index())

	padding := longestLine - len(str)
	for i := 0; i < padding; i++ {
		str += " "
	}

	if index == m.Index() {
		style := lipgloss.NewStyle().Foreground(styles.LinearPurpleDark)
		fmt.Fprint(w, style.Render(str))
	} else {
		fmt.Fprint(w, str)
	}
}

func (p *PullsModel) Init() tea.Cmd {
	p.list = list.New([]list.Item{}, itemDelegate{}, 0, 0)
	p.keys = tui.Keys
	p.list.Title = "Pull Requests"
	p.list.SetShowHelp(false)
	p.list.SetShowStatusBar(false)
	p.list.SetStatusBarItemName("pull request", "pull requests")
	p.list.SetFilteringEnabled(false)
	p.list.SetShowTitle(false)
	p.list.SetShowPagination(false)
	p.list.Styles.NoItems = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).PaddingBottom(1)

	return nil
}

func (p *PullsModel) UpdateList(attachments []*linear.Attachment) {
	pulls := []list.Item{}

	sort.Slice(attachments, func(i, j int) bool {
		return attachments[i].UpdatedAt.After(attachments[j].UpdatedAt)
	})
	for _, attachment := range attachments {
		pulls = append(pulls, Attachment{
			title: attachment.Title,
			data:  attachment,
		})
	}

	p.list.SetItems(pulls)
	p.list.ResetSelected()

	delegate := itemDelegate{}
	height := len(pulls)*delegate.Height() + (len(pulls)-1)*delegate.Spacing()
	p.list.SetHeight(height + 1)
}

func (p *PullsModel) GetSelectedItem() *linear.Attachment {
	selectedItem := p.list.SelectedItem()
	if selectedItem == nil {
		return &linear.Attachment{}
	}

	return selectedItem.(Attachment).data
}

func (p PullsModel) Update(msg tea.Msg) (PullsModel, tea.Cmd) {
	var cmd tea.Cmd
	pull := p.GetSelectedItem()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, p.keys.P), key.Matches(msg, p.keys.Esc):
			p.Selecting = false
			return p, nil
		case key.Matches(msg, p.keys.Enter):
			util.OpenURL(pull.Url)
			p.Selecting = false
			break
		default:
			break
		}
	}

	p.list, cmd = p.list.Update(msg)
	return p, cmd
}

func (p *PullsModel) View() string {
	style := lipgloss.NewStyle() //.Background(styles.LinearPurpleDarker)
	return style.Render(p.list.View())
}
