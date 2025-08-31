package main

import (
	"log"
	"os"
	"os/exec"
)

func main() {
	// For convenience in local dev, just run the real server binary if built; otherwise instruct how to run.
	if _, err := os.Stat("cmd/server"); err == nil {
		cmd := exec.Command("go", "run", "./cmd/server")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
		return
	}
	log.Println("Use: go run ./cmd/server to start the API server")
}