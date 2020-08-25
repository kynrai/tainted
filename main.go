package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/build"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func usage() {
	fmt.Println(`Usage: <list of packages via stdin> | tainted

Example:
	go list ./... | tainted

This program takes a list of packages from stdin and returns a list of packages
which have beend tained and need to be rebuilt. A package is tained when one or
more of its dependacies have been modified`)
	fmt.Println()
	flag.PrintDefaults()
}

var (
	packages         map[string]struct{}       // the packages to check for taint
	changedDirs      map[string]struct{}       // the directories which contain modified files
	cache            map[string]*build.Package // a map[>package name>]<build.Package> to skip repeat lookups
	gitDirPtr        *string                   // the git directory to check for changes
	commitFromPtr    *string                   // the earliest commit to diff
	commitToPtr      *string                   // the latest commit to diff
	includeTestFiles *bool                     // this will include test files for evaluation
)

func init() {
	cache = make(map[string]*build.Package)
	changedDirs = make(map[string]struct{})
	packages = make(map[string]struct{})
}

func main() {
	gitDirPtr = flag.String("dir", ".", "the git directory to check")
	commitFromPtr = flag.String("from", "HEAD~1", "commit to take changes from")
	commitToPtr = flag.String("to", "HEAD", "commit to take changes to")
	includeTestFiles = flag.Bool("test", false, "include test files")

	flag.Usage = usage
	flag.Parse()

	// check if we have anything coming from stdin
	stat, err := os.Stdin.Stat()
	if err != nil {
		fmt.Printf("failed to read from stdin: %s", err)
		os.Exit(1)
	}
	if (stat.Mode() & os.ModeNamedPipe) == 0 {
		flag.Usage()
		os.Exit(0)
	}

	// populate the changed directories slice
	modified()

	// read packages from stdin
	readPackages()

	// for each package we want to get its full deps tree and see if it
	// contains any elements from the changedDirs
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	output := make(map[string]struct{})
	for k := range packages {
		// get all the deps
		deps, err := findDeps(k, cwd)
		if err != nil {
			log.Fatal(err)
		}
		if hasChanges(deps) {
			output[k] = struct{}{}
		}
	}
	// finally to make it all pretty, sort it in a slice
	prettyOutput := make([]string, 0, len(output))
	for k := range output {
		prettyOutput = append(prettyOutput, k)
	}
	if len(prettyOutput) == 0 {
		return
	}
	sort.Strings(prettyOutput)
	fmt.Println(strings.Join(prettyOutput, "\n"))
}

// checks to see if any of the deps have the same suffix as anything in the changedDirs
func hasChanges(deps []string) bool {
	for _, v := range deps {
		for k := range changedDirs {
			if strings.HasSuffix(v, k) {
				return true
			}
		}
	}
	return false
}

// read all the packages from stdin into the global packages var
func readPackages() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		packages[scanner.Text()] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

// modified will use git to find out which folders have been changed
func modified() {
	cmdArgs := []string{
		"--no-pager",
		"-C",
		*gitDirPtr,
		"diff",
		"--name-only",
		*commitFromPtr,
		*commitToPtr,
	}

	cmd := exec.Command("git", cmdArgs...)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			scanned := scanner.Text()
			if !*includeTestFiles && strings.Contains(scanned, "_test.go") {
				continue
			}

			if dir := filepath.Dir(scanned); dir != "." {
				changedDirs[dir] = struct{}{}
			}
		}
	}()

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}
