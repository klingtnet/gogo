package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
)

func findWorkspace(start, godir string) (string, error) {
	parts := strings.Split(start, string(os.PathSeparator))
	for l := len(parts); l > 0; l-- {
		dir := path.Join(string(os.PathSeparator), path.Join(parts[:l]...), godir)
		info, err := os.Stat(dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		if info.IsDir() {
			return dir, nil
		}
	}

	return "", fmt.Errorf("workspace not found")
}

const (
	ErrToolchainMissing = iota
	ErrMissingArguments
	ErrWorkspaceProblem
	ErrGoFailed
)

func exitIfError(err error, code int, format string, args ...interface{}) {
	if err == nil {
		return
	}
	if format != "" {
		fmt.Fprintf(os.Stderr, format, args...)
	}
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(ErrToolchainMissing)
}

func printUsage() {
	name := os.Args[0]
	fmt.Printf(`Usage: %s <go-command> [argument]...

	Example: %s build app.go
`, name, name)
}

func main() {
	path, err := exec.LookPath("go")
	exitIfError(err, ErrToolchainMissing, "go is not installed")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(ErrMissingArguments)
	}

	wd, err := os.Getwd()
	exitIfError(err, ErrWorkspaceProblem, "could not get working directory")
	_, err = findWorkspace(wd, ".go")
	exitIfError(err, ErrWorkspaceProblem, "")

	cmd := exec.Command(path, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		exitCode := ErrGoFailed
		if err, ok := err.(*exec.ExitError); ok {
			if exitStatus, ok := err.Sys().(syscall.WaitStatus); ok {
				exitCode = exitStatus.ExitStatus()
			}
		}
		os.Exit(exitCode)
	}
}
