package chooser

import (
	"bufio"
	"fmt"
	"go/build"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

// Chooser will determine which packages contain imports which have been altered between two commits.
type Chooser struct {
	FullPkgPath         string              // The full path on the file system to the package
	RootDir             string              // The highest level directory to be used for comparing (usually the top of the repo e.g github.com/golang/go)
	FromCommit          string              // The earliest commit to take changes from
	ToCommit            string              // The latest commit to take changes from
	Changes             map[string]struct{} // A map containing the directories of all files that have changed between the specifiec commits
	Packages            []string            // All the packages to be checked for changes (all the packages under the RootDir)
	PackagesWithImports []Package           // Packages wih all their imports
	Chosen              []string            // List of all packages which contain imports that have been alatered
}

func New(rootDir, from, to string) (*Chooser, error) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	fullPath := filepath.Join(gopath, "src", rootDir)
	fi, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}
	if mode := fi.Mode(); !mode.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", fullPath)
	}
	return &Chooser{
		FullPkgPath: fullPath,
		RootDir:     rootDir,
		FromCommit:  from,
		ToCommit:    to,
		Changes:     make(map[string]struct{}),
	}, nil
}

// Run will execute the chooser and populate the list of Changes, Packages and Chosen with the results of the run.
func (c *Chooser) Run() error {
	if err := c.modified(); err != nil {
		return err
	}
	if err := c.packages(); err != nil {
		return err
	}

	lock := sync.Mutex{}

	wg := sync.WaitGroup{}
	throttle := make(chan struct{}, runtime.NumCPU())
	for _, p := range c.Packages {
		go func(p string) {
			wg.Add(1)
			throttle <- struct{}{}
			defer func() {
				<-throttle
				wg.Done()
			}()
			imports, err := c.imports(p)
			if err != nil {
				log.Fatal(err)
			}
			lock.Lock()
			defer lock.Unlock()
			c.PackagesWithImports = append(c.PackagesWithImports, Package{
				RootDir:    c.RootDir,
				ImportPath: p,
				Imports:    imports,
			})
		}(p)
	}
	wg.Wait()

	if err := c.choose(); err != nil {
		return err
	}
	for _, v := range c.Chosen {
		fmt.Println(v)
	}
	return nil
}

// modified will use git diff to determin which directories have been modified between two commits.
// It will get a list of all files, then dedupe the directories with a map.
func (c *Chooser) modified() error {
	if c.FullPkgPath == "" {
		c.FullPkgPath = "."
	}
	cmdArgs := []string{
		"--no-pager",
		"-C",
		c.FullPkgPath,
		"diff",
		"--name-only",
		c.FromCommit,
		c.ToCommit,
	}

	cmd := exec.Command(gitCommand, cmdArgs...)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			if dir := filepath.Dir(scanner.Text()); dir != "." {
				c.Changes[dir] = struct{}{}
			}
		}
	}()

	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

// packages uses the go to list all go packages under the given root path
func (c *Chooser) packages() error {
	cmdArgs := []string{
		"list",
		"./...",
	}
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = c.FullPkgPath
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			c.Packages = append(c.Packages, scanner.Text())
		}
	}()

	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

// imports returns a list of non standard imports used by a given pkg
func (c *Chooser) imports(pkg string) ([]string, error) {
	importsArgs := []string{
		"list",
		"-f",
		`{{join .Deps "\n"}}`,
		pkg,
	}
	importsCmd := exec.Command("go", importsArgs...)
	importsCmd.Dir = c.FullPkgPath

	filterArgs := []string{
		"go",
		"list",
		"-f",
		`{{if not .Standard}}{{.ImportPath}}{{end}}`,
	}
	filterCmd := exec.Command("xargs", filterArgs...)
	filterCmd.Dir = c.FullPkgPath

	filterCmd.Stdin, _ = importsCmd.StdoutPipe()

	filterReader, err := filterCmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	var files []string
	scanner := bufio.NewScanner(filterReader)
	go func() {
		for scanner.Scan() {
			files = append(files, strings.TrimPrefix(scanner.Text(), c.RootDir))
		}
	}()

	if err := filterCmd.Start(); err != nil {
		return nil, err
	}
	if err := importsCmd.Run(); err != nil {
		return nil, err
	}
	if err := filterCmd.Wait(); err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func (c *Chooser) choose() error {
	dedupe := make(map[string]struct{})
	wg := sync.WaitGroup{}
	for _, v := range c.PackagesWithImports {
		go func(p Package) {
			wg.Add(1)
			defer wg.Done()
			for _, r := range p.Imports {
				// ignore anything outside of the cmd directory
				if !strings.HasPrefix(p.ImportPath, filepath.Join(c.RootDir, "cmd")) {
					continue
				}
				if _, ok := c.Changes[r]; ok {
					// if one of the imports is in the changes list then this package needs to be rebuilt
					dedupe[p.ImportPath] = struct{}{}
					break
				}
			}
		}(v)
	}
	wg.Wait()
	for p := range dedupe {
		c.Chosen = append(c.Chosen, p)
	}
	sort.Strings(c.Chosen)
	return nil
}
