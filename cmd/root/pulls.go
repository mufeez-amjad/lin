package root

import (
	"lin_cli/internal/linear"
	"lin_cli/internal/tui"
	"lin_cli/internal/util"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type pulls struct {
	list      list.Model
	selecting bool

	keys tui.KeyMap
}

type Attachment struct {
	title string
	data  *linear.Attachment
}

func (a Attachment) Title() string       { return a.title }
func (a Attachment) Description() string { return "" }
func (a Attachment) FilterValue() string { return a.Title() + a.Description() }

func (p *pulls) Init() tea.Cmd {
	p.list = list.New([]list.Item{}, list.NewDefaultDelegate(), 45, 15)
	p.keys = tui.Keys
	p.list.Title = "Pull Requests"
	p.list.SetShowHelp(false)
	p.list.SetShowStatusBar(false)
	p.list.SetStatusBarItemName("pull request", "pull requests")
	p.list.SetFilteringEnabled(false)

	return nil
}

func (p *pulls) UpdateList(attachments []*linear.Attachment) {
	pulls := []list.Item{}
	for _, attachment := range attachments {
		pulls = append(pulls, Attachment{
			title: attachment.Title,
			data:  attachment,
		})
	}

	p.list.SetItems(pulls)
}

func (p pulls) GetSelectedItem() *linear.Attachment {
	selectedItem := p.list.SelectedItem()
	if selectedItem == nil {
		return &linear.Attachment{}
	}

	return selectedItem.(Attachment).data
}

func (p pulls) Update(msg tea.Msg) (pulls, tea.Cmd) {
	var cmd tea.Cmd
	pull := p.GetSelectedItem()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, p.keys.P), key.Matches(msg, p.keys.Esc):
			p.selecting = false
			return p, nil
		case key.Matches(msg, p.keys.Enter):
			util.OpenURL(pull.Url)
			break
		default:
			break
		}
	}

	p.list, cmd = p.list.Update(msg)
	return p, cmd
}

func (p pulls) View() string {
	return p.list.View()
}
