package main

import (
	"log"
	"os"
	"os/exec"
)

func main() {
	cmd := "server"
	if len(os.Args) > 1 && os.Args[1] != "" {
		cmd = os.Args[1]
	}
	switch cmd {
	case "server":
		mustExec("/server")
	case "wsserver":
		mustExec("/wsserver")
	default:
		log.Fatalf("unknown command: %s (use 'server' or 'wsserver')", cmd)
	}
}

func mustExec(path string) {
	c := exec.Command(path)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = os.Environ()
	if err := c.Run(); err != nil {
		log.Fatal(err)
	}
}
