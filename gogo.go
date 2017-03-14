package main

import (
	"encoding/json"
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

const MetadataFilename = "gogo.json"

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
	fmt.Println(`Usage: gogo [<go-command>|boostrap] [argument]...

Example:
	- calling a go command: 'gogo build app.go'
	- bootstrapping the local GOPATH: 'gogo boostrap <import-path>'
		The import path is usually something like 'github.com/user/project'.
		Note that the bootstrap command should be run from the projects root directory!
	- print this message: 'gogo'
`)
}

type metaData struct {
	Project    string `json:"projectName"`
	Namespace  string `json:"namespace"`
	ImportPath string `json:"importPath"`
}

func boostrap(wd, workspace string, args ...string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "import path is missing\n")
		printUsage()
		os.Exit(ErrMissingArguments)
	}

	loc, _ := findWorkspace(wd, workspace)
	if loc != "" {
		fmt.Printf("project is already boostrapped in %q\n", loc)
		return
	}

	srcPath := path.Join(wd, workspace, "src")
	err := os.MkdirAll(srcPath, 0755)
	exitIfError(err, ErrWorkspaceProblem, "could not create workspace directory\n")

	// e.g. let importPath="github.com/klingtnet/gogo" then namespace="github.com/klingtnet" and project="gogo"
	// or let importPath="gogo" then namespace="" and project="gogo"
	importPath := args[1]
	parts := strings.Split(importPath, string(os.PathSeparator))
	l := len(parts)
	project := parts[l-1]
	namespace := ""
	if l > 1 {
		namespace = path.Join(parts[:l-1]...)
		err := os.MkdirAll(path.Join(srcPath, namespace), 0755)
		exitIfError(err, ErrWorkspaceProblem, "could not create project directory in workspace\n")
	}

	fullPath := path.Clean(path.Join(srcPath, namespace, project))
	err = os.Symlink(wd, fullPath)
	exitIfError(err, ErrWorkspaceProblem, "could not create symbolic link from %q to %q\n", wd, fullPath)

	fpath := path.Join(wd, workspace, MetadataFilename)
	jsonFile, err := os.Create(fpath)
	exitIfError(err, ErrWorkspaceProblem, "could not write metadata into: %q\n", fpath)
	defer jsonFile.Close()

	enc := json.NewEncoder(jsonFile)
	err = enc.Encode(metaData{
		Project:    project,
		Namespace:  namespace,
		ImportPath: importPath,
	})
	exitIfError(err, ErrWorkspaceProblem, "could not marshal metadata to JSON\n")
}

func goCommand(wd, goCmd, workspace string, args ...string) {
	loc, err := findWorkspace(wd, workspace)
	exitIfError(err, ErrWorkspaceProblem, "did you forget to boostrap the local gopath?\n")

	fpath := path.Join(loc, MetadataFilename)
	jsonFile, err := os.Open(fpath)
	exitIfError(err, ErrWorkspaceProblem, "metadata file not found: %q\n", fpath)
	defer jsonFile.Close()
	decoder := json.NewDecoder(jsonFile)
	meta := metaData{}
	err = decoder.Decode(&meta)
	exitIfError(err, ErrWorkspaceProblem, "decoding %q failed\n", fpath)

	cmd := exec.Command(goCmd, args...)
	cmd.Env = []string{"GOPATH=" + loc}
	cmd.Dir = path.Join(loc, "src", meta.Namespace, meta.Project)
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

	wd, err := os.Getwd()
	exitIfError(err, ErrWorkspaceProblem, "could not get working directory")

	workspace := ".gogo"
	switch os.Args[1] {
	case "bootstrap":
		boostrap(wd, workspace, os.Args[1:]...)
	default:
		goCommand(wd, goCmd, workspace, os.Args[1:]...)
	}
}
