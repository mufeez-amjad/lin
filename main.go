package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"lin_cli/internal/git"
	"lin_cli/internal/tui"
	"lin_cli/internal/util"
	linproto "lin_cli/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	bList "github.com/charmbracelet/bubbles/list"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

// sessionState is used to track which model is focused
type sessionState uint

const (
	listView sessionState = iota
	issueView
)

type Issue struct {
	identifier string
	title      string
	data       *linproto.Issue
}

func (i Issue) Title() string         { return i.identifier }
func (i Issue) Description() string   { return i.title }
func (i Issue) Data() *linproto.Issue { return i.data }
func (i Issue) FilterValue() string   { return i.identifier + " " + i.title }

type model struct {
	list  bList.Model
	help  help.Model
	keys  tui.KeyMap
	state sessionState

	client  linproto.LinearClient
	loading bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func writeProtobufToFile(filename string, messages []*linproto.Issue) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, msg := range messages {
		data, err := proto.Marshal(msg)
		if err != nil {
			return err
		}

		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(len(data)))

		if _, err := file.Write(buf); err != nil {
			return err
		}

		if _, err := file.Write(data); err != nil {
			return err
		}
	}

	return nil
}

func readProtobufFromFile(filepath string) ([]*linproto.Issue, error) {
	_, err := os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	file, err := os.OpenFile(filepath, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var offset int64
	content := make([][]byte, 0)
	for {
		buf := make([]byte, 4)
		if _, err := file.ReadAt(buf, offset); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		itemSize := binary.LittleEndian.Uint32(buf)
		offset += 4

		item := make([]byte, itemSize)
		if _, err := file.ReadAt(item, offset); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		content = append(content, item)
		offset += int64(itemSize)
	}

	var messages []*linproto.Issue

	for _, item := range content {
		t := new(linproto.Issue)
		err = proto.Unmarshal(item, t)
		if err != nil {
			return nil, err
		}
		messages = append(messages, t)
	}

	return messages, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	issue := m.list.SelectedItem().(Issue).Data()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.C):
			// TODO: handle multiple branches (based on issue attachments)
			branchName := issue.GetBranchName()
			err := git.CheckoutBranch(branchName)
			if err != nil {
				fmt.Printf("%s", err)
			}
			return m, tea.Quit
		case key.Matches(msg, m.keys.Enter):
			util.OpenURL(issue.GetUrl())
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

func spawnServer() (int, error) {
	// Command to run the TypeScript server
	cmd := exec.Command("node", "index.js")

	// Set the current working directory to the directory containing the TypeScript server script
	cmd.Dir = "server/dist"

	// TODO: Remove after development
	// Redirect standard output and error streams to the Go process's standard output
	// and error streams
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start the TypeScript server as a child process
	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting TypeScript server: %v\n", err)
		return 0, err
	}

	// Capture the child process ID
	childPid := cmd.Process.Pid

	return childPid, nil
}

func main() {
	issues, err := readProtobufFromFile("./cache")
	if err != nil {
		log.Fatalf("Failed to open cache file")
	}
	fmt.Printf("cached: %v\n", len(issues))

	// Connect to the gRPC server and fetch issues async.
	childPidChan := make(chan int, 1)
	connChan := make(chan *grpc.ClientConn, 1)
	go func() {
		// Spawn the TypeScript server
		spawnedPid, err := spawnServer()
		if err != nil {
			fmt.Println("Error spawning child process:", err)
			childPidChan <- -1
		}

		childPidChan <- spawnedPid

		time.Sleep(2 * time.Second)

		// Connect to the TypeScript server
		for {
			connection, err := grpc.Dial(
				"0.0.0.0:50051",
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				log.Printf("Failed to connect: %v. Retrying...", err)
				time.Sleep(2 * time.Second) // Adjust the retry interval as needed
				continue
			}
			connChan <- connection
			break // Connection successful, exit the loop
		}

		fmt.Println("Connected to server")
	}()

	childPid := <-childPidChan

	// Defer the termination of the child process
	defer func() {
		if childPid == -1 {
			os.Exit(1)
		}

		err := syscall.Kill(childPid, syscall.SIGKILL)
		if err != nil {
			fmt.Println("Error killing child process:", err)
		}
	}()

	conn := <-connChan
	defer conn.Close()
	client := linproto.NewLinearClient(conn)

	if len(issues) == 0 {
		issuesResp := make(chan *linproto.GetIssuesResponse, 1)
		go func() {
			req := &linproto.GetIssuesRequest{
				ApiKey: "",
			}

			resp, err := client.GetIssues(context.Background(), req)
			if err != nil {
				issuesResp <- &linproto.GetIssuesResponse{}
				log.Fatalf("GetIssues failed: %v", err)
			}

			issuesResp <- resp

			err = writeProtobufToFile("./cache", resp.Issues)
			if err != nil {
				fmt.Printf("Failed to cache issues")
			}
		}()
		issues = (<-issuesResp).Issues
	}

	items := []bList.Item{}

	for _, issue := range issues {
		items = append(items, Issue{
			identifier: issue.GetIdentifier(),
			title:      issue.GetTitle(),
			data:       issue,
		})
	}

	m := model{
		list:   bList.New(items, bList.NewDefaultDelegate(), 0, 0),
		state:  listView,
		keys:   tui.Keys,
		help:   help.New(),
		client: client,
	}

	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		return m.keys.ShortHelp()
	}
	m.list.Title = "Assigned Issues"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	syscall.Kill(childPid, syscall.SIGKILL)
}
