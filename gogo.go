package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
)

func isDir(dir string) (bool, error) {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if info.IsDir() {
		return true, nil
	}
	return false, nil
}

func findWorkspace(start, godir string) (string, error) {
	parts := strings.Split(start, string(os.PathSeparator))
	for l := len(parts); l > 0; l-- {
		dir := path.Join(string(os.PathSeparator), path.Join(parts[:l]...), godir)
		found, err := isDir(dir)
		if err != nil {
			return "", err
		}
		if found {
			return dir, nil
		}
	}

	return "", fmt.Errorf("workspace not found")
}

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

	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get current directory: %s\n", err)
		os.Exit(3)
	}
	_, err = findWorkspace(wd, ".go")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(5)
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
