package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"log"
	"context"
	"time"
	
	"google.golang.org/grpc"
	"lin_cli/proto"

	"github.com/charmbracelet/bubbles/list"
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

type item struct {
	title string
	desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.desc }

type model struct {
	list list.Model
	state   sessionState
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			if m.state == listView {
				m.state = issueView
			} else {
				m.state = listView
			}
		case "c":
			// branchName := m.list.SelectedItem().BranchName
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
    cmd.Dir = "server/src"

	// TODO: Remove after development
    // Redirect standard output and error streams to the Go process's standard output and error streams
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
	// Spawn the TypeScript server
	childPid, err := spawnServer()
	if err != nil {
		fmt.Println("Error spawning child process:", err)
		os.Exit(1)
	}

	// Defer the termination of the child process
	defer func() {
		err := syscall.Kill(childPid, syscall.SIGKILL)
		if err != nil {
			fmt.Println("Error killing child process:", err)
		}
	}()

	time.Sleep(2 * time.Second)

	// Connect to the TypeScript server
	var conn *grpc.ClientConn
    for {
        var err error
        conn, err = grpc.Dial("0.0.0.0:50051", grpc.WithInsecure())
        if err != nil {
            log.Printf("Failed to connect: %v. Retrying...", err)
            time.Sleep(2 * time.Second) // Adjust the retry interval as needed
            continue
        }
        break // Connection successful, exit the loop
    }
	defer conn.Close()
	fmt.Println("Connected to server")

	client := proto.NewLinearClient(conn)

	req := &proto.GetIssuesRequest{
		ApiKey: "lin_api_3nQjimQ7hLcGfMIX0pAiQkS3Li9cizR2WfL3wAui",
	}

    resp, err := client.GetIssues(context.Background(), req)
    if err != nil {
        log.Fatalf("GetIssues failed: %v", err)
    }

	// Do stuff in the CLI

	items := []list.Item{}

	for _, issue := range resp.Issues {
		items = append(items, item{title: issue.GetIdentifier(), desc: issue.GetTitle()})
	}

	m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0), state: listView}
	m.list.Title = "Assigned Issues"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	// syscall.Kill(childPid, syscall.SIGKILL)

	// // Wait for the TypeScript server to exit (you can also implement logic to gracefully handle server exit)
    // if err := cmd.Wait(); err != nil {
    //     fmt.Printf("TypeScript server exited with error: %v\n", err)
    // }
}