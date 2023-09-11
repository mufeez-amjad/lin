package grpc

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"lin_cli/internal/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	client *proto.LinearClient

	serverPid      int
	initServerOnce sync.Once
)

func GetClient() *proto.LinearClient {
	return client
}

// Spawns the Typescript server
func spawnServer() (int, error) {
	cmd := exec.Command("node", "index.js")
	cmd.Dir = "server/dist"

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting TypeScript server: %v\n", err)
		return 0, err
	}

	serverPid = cmd.Process.Pid
	return serverPid, nil
}

// Connects to the Typescript server through grpc
func connectToServer(pid chan int, conn chan *grpc.ClientConn) {
	var err error
	spawnedPid := serverPid

	initServerOnce.Do(func() {
		spawnedPid, err = spawnServer()
		if err != nil {
			fmt.Println("Error spawning child process:", err)
			pid <- -1
		}
	})
	pid <- spawnedPid

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
		conn <- connection
		break // Connection successful, exit the loop
	}

	fmt.Println("Connected to server")
}
