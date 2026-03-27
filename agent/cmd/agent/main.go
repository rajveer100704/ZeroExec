package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rajveer100704/ZeroExec/agent"
)

func main() {
	sm := agent.NewSessionManager()
	defer sm.Cleanup()

	// Set up signal handling for graceful exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Println("ZeroExec Windows Agent Starting...")
	fmt.Println("------------------------------------")

	// Start a default PowerShell session for testing
	session, err := sm.StartSession("powershell.exe", 24, 80)
	if err != nil {
		log.Fatalf("Failed to start session: %v", err)
	}
	fmt.Printf("Started Session: %s (PID: %d)\n", session.ID, session.PTY.PID())
	fmt.Println("Type commands below (Ctrl+C to exit):")

	// PTY Output -> Local Stdout
	go func() {
		io.Copy(os.Stdout, session.PTY)
	}()

	// Local Stdin -> PTY Input
	go func() {
		io.Copy(session.PTY, os.Stdin)
	}()

	<-sigChan
	fmt.Println("\nCleaning up and exiting...")
}
