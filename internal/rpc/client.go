package rpc

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"lin_cli/internal/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	client *Client

	serverPid      int
	initServerOnce sync.Once
	initClientOnce sync.Once
)

type Client struct {
	inner proto.LinearClient

	conn      *grpc.ClientConn
	serverPid int

	err error
}

func (c *Client) Get() proto.LinearClient {
	return c.inner
}

func InitClient() *Client {
	initClientOnce.Do(func() {
		pid, conn, err := connectToServer()
		client = &Client{
			inner:     proto.NewLinearClient(conn),
			conn:      conn,
			serverPid: pid,
			err:       err,
		}
	})

	return client
}

func (c *Client) GetErr() error {
	return c.err
}

func (c *Client) Cleanup() {
	c.conn.Close()
	syscall.Kill(c.serverPid, syscall.SIGKILL)
}

// Spawns the Typescript server
func spawnServer() (int, error) {
	cmd := exec.Command("node", "index.js")
	cmd.Dir = "../../server/dist"

	// cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting TypeScript server: %v\n", err)
		return 0, err
	}

	serverPid = cmd.Process.Pid
	return serverPid, nil
}

// Connects to the Typescript server through grpc
func connectToServer() (pid int, conn *grpc.ClientConn, err error) {
	pid = serverPid
	initServerOnce.Do(func() {
		pid, err = spawnServer()
		if err != nil {
			fmt.Println("Error spawning child process:", err)
			pid, conn = -1, nil
			return
		}
	})

	time.Sleep(1 * time.Second)

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
		conn = connection
		break // Connection successful, exit the loop
	}

	return
}
