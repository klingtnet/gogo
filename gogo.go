package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	path, err := exec.LookPath("go")
	if err != nil {
		fmt.Printf("go is not installed: %s\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s [build|run|test...] ...\n", os.Args[0])
		os.Exit(2)
	}

	cmd := exec.Command(path, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		exitCode := 3
		if err, ok := err.(*exec.ExitError); ok {
			if exitStatus, ok := err.Sys().(syscall.WaitStatus); ok {
				exitCode = exitStatus.ExitStatus()
			}
		}
		os.Exit(exitCode)
	}
}
