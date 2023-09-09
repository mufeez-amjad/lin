package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"

	linproto "lin_cli/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	bList "github.com/charmbracelet/bubbles/list"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

// sessionState is used to track which model is focused
type sessionState uint

const (
	listView sessionState = iota
	issueView
)

type keyMap struct {
	Enter key.Binding
	C     key.Binding
	Quit  key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.C, k.Enter}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.C, k.Enter}, // first column
	}
}

var keys = keyMap{
	C: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "checkout branch"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open issue"),
	),
}

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
	keys  keyMap
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

		_, err = file.Write(data)
		if err != nil {
			return err
		}
	}

	return nil
}

func readProtobufFromFile(filepath string) ([]*linproto.Issue, bool, error) {
	_, err := os.Stat(filepath)

	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		} else {
			return nil, false, nil
		}
	}

	file, err := os.Open(filepath)
	if err != nil {
		return nil, true, err
	}
	defer file.Close()

	var messages []*linproto.Issue
	for {
		msg := new(linproto.Issue)
		data := make([]byte, proto.Size(msg))
		n, err := file.Read(data)
		if err != nil {
			break
		}
		if n == 0 {
			break
		}
		if err := proto.Unmarshal(data, msg); err != nil {
			return nil, true, err
		}
		messages = append(messages, msg)
	}

	return messages, true, nil
}

func checkoutBranch(branchName string) error {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Error getting current working directory: %v\n", err)
	}

	r, err := git.PlainOpen(cwd)
	if err != nil {
		return fmt.Errorf("Error opening repository: %v\n", err)
	}

	refs, err := r.Branches()
	if err != nil {
		return fmt.Errorf("Error getting branches: %v\n", err)
	}

	branchExists := false
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().Short() == branchName {
			branchExists = true
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("Error checking branch existence: %v\n", err)
	}

	// If the branch doesn't exist, create it
	if !branchExists {
		// Create a new branch from the current HEAD
		headRef, err := r.Head()
		if err != nil {
			fmt.Printf("Error getting HEAD reference: %v\n", err)
			os.Exit(1)
		}

		newBranchRef := plumbing.NewBranchReferenceName(branchName)
		branch := plumbing.NewHashReference(newBranchRef, headRef.Hash())

		if err := r.Storer.SetReference(branch); err != nil {
			fmt.Printf("Error creating branch: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Created branch '%s'\n", branchName)
	}

	// Checkout the branch
	w, err := r.Worktree()
	if err != nil {
		fmt.Printf("Error getting worktree: %v\n", err)
		os.Exit(1)
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName)),
		Keep:   true,
	})
	if err != nil {
		fmt.Printf("Error checking out branch: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Checked out branch '%s'\n", branchName)
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	issue := m.list.SelectedItem().(Issue).Data()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.C):
			// TODO: handle multiple branches (based on issue attachments)
			branchName := issue.GetBranchName()
			err := checkoutBranch(branchName)
			if err != nil {
				fmt.Printf("%s", err)
			}
			return m, tea.Quit
		case key.Matches(msg, m.keys.Enter):
			openURL(issue.GetUrl())
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

func openURL(href string) error {
	var cmd *exec.Cmd

	// TODO: open desktop app instead of browser

	// Determine the operating system and open the URL accordingly
	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.Command("open", href)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", href)
	case "linux":
		cmd = exec.Command("xdg-open", href)
	default:
		return fmt.Errorf("unsupported platform")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
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
	_, _, err := readProtobufFromFile("./cache")
	if err != nil {
		log.Fatalf("Failed to open cache file")
	}

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

	issuesResp := make(chan *linproto.GetIssuesResponse, 1)
	go func() {
		req := &linproto.GetIssuesRequest{
			ApiKey: "",
		}

		resp, err := client.GetIssues(context.Background(), req)
		if err != nil {
			log.Fatalf("GetIssues failed: %v", err)
		}
		issuesResp <- resp
	}()

	items := []bList.Item{}

	for _, issue := range (<-issuesResp).Issues {
		items = append(items, Issue{
			identifier: issue.GetIdentifier(),
			title:      issue.GetTitle(),
			data:       issue,
		})
	}

	m := model{
		list:  bList.New(items, bList.NewDefaultDelegate(), 0, 0),
		state: listView,
		keys:  keys, help: help.New(),
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
