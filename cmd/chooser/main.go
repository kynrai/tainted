package main

import (
	"flag"
	"log"
	"os"
	"os/exec"

	"github.com/kynrai/chooser"
)

const (
	gitCommand = "git"
	goCommand  = "go"
)

var (
	rootDir *string
	from    *string
	to      *string
)

func main() {
	// make sure we have the root directory
	rootDir = flag.String("dir", "", "Root directory to scan, must be a git repo and contain go packages")
	from = flag.String("from", "", "git commit to scan from")
	to = flag.String("to", "", "git commit to scan to")
	flag.Parse()

	if *rootDir == "" || *from == "" || *to == "" {
		flag.Usage()
		os.Exit(0)
	}

	// check commands we need are present
	if !cmdExists(gitCommand) {
		log.Fatalf("%s is missing", gitCommand)
	}
	if !cmdExists(goCommand) {
		log.Fatalf("%s is missing", goCommand)
	}
	e, err := chooser.New(*rootDir, *from, *to)
	if err != nil {
		log.Fatal(err)
	}
	if err := e.Run(); err != nil {
		log.Fatal(err)
	}
}

// this uses the common "which" command to determin if a command exists
func cmdExists(cmd string) bool {
	c := exec.Command("which", "-s", cmd)
	if err := c.Start(); err != nil {
		return false
	}
	if err := c.Wait(); err != nil {
		return false
	}
	return true
}
