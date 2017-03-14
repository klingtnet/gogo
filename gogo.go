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
	fmt.Println(`Usage: gogo [<go-command>|boostrap|help|usage] [argument]...

Example:
	- calling a go command: 'gogo build app.go'
	- bootstrapping the local GOPATH: 'gogo boostrap'
		Note that the bootstrap command should be run from the projects root directory!
	- print this message: 'gogo', 'gogo help' or 'gogo usage'
`)
}

func boostrap() {
}

func goCommand(goCmd string, args ...string) {
	wd, err := os.Getwd()
	exitIfError(err, ErrWorkspaceProblem, "could not get working directory")
	loc, err := findWorkspace(wd, ".go")
	exitIfError(err, ErrWorkspaceProblem, "did you forget to boostrap the local gopath?\n")
	fmt.Println(loc)

	cmd := exec.Command(goCmd, args...)
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

func main() {
	goCmd, err := exec.LookPath("go")
	exitIfError(err, ErrToolchainMissing, "go is not installed")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(ErrMissingArguments)
	}

	switch os.Args[1] {
	case "boostrap":
		boostrap()
	case "help", "usage":
		printUsage()
	default:
		goCommand(goCmd, os.Args[1:]...)
	}
}
