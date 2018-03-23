package main

import (
	"flag"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

var flagGitDir = flag.String("git-dir", "~/GIT", "Sets the root directory for all git repositories")

func main() {
	flag.Parse()
	for _, repo := range flag.Args() {
		parentDir := projectParentDir(repo)
		if err := createLocalDir(parentDir); err != nil {
			log.Fatalf("Could not create local dir: %s. Got error: %s\n", parentDir, err.Error())
		}
		runGitClone(repo, parentDir)
		log.Printf("Cloned %s into %s\n", repo, parentDir)
	}
}

func runGitClone(repo string, projectParentDir string) {
	cmd := exec.Command("git", "clone", repo)
	cmd.Dir = projectParentDir
	stderr, stdErrErr := cmd.StderrPipe()
	stdout, stdOutErr := cmd.StdoutPipe()
	if stdErrErr != nil || stdOutErr != nil {
		log.Fatalf("Could not get pipes from subcommand. Error: %s / %s", stdErrErr.Error(), stdOutErr.Error())
	}
	if err := cmd.Start(); err != nil {
		log.Fatalf("Could not clone repository %s into directory %s. Error: %s\n", repo, projectParentDir, err.Error())
	}
	io.Copy(os.Stderr, stderr)
	io.Copy(os.Stdout, stdout)
	if err := cmd.Wait(); err != nil {
		log.Fatalf("Could not clone repository %s into directory %s. Error: %s\n", repo, projectParentDir, err.Error())
	}
}

func createLocalDir(projectParentDir string) error {
	return os.MkdirAll(projectParentDir, os.ModePerm)
}

func projectPath(u *url.URL) string {
	return filepath.Join(u.Host, filepath.Dir(u.Path))
}

func projectParentDir(repo string) string {
	repo = strings.Replace(repo, "github.com:", "github.com/", 1)
	u, err := url.Parse(repo)
	if err != nil {
		log.Fatalf("Could not parse repo URL: %s. Got error: %s\n", repo, err.Error())
	}
	return filepath.Join(rootDir(), projectPath(u))
}

func rootDir() string {
	if strings.Contains(*flagGitDir, "~") {
		u, err := user.Current()
		if err != nil {
			log.Fatal("Could not determine current user to resolve '~' .")
		}
		return strings.Replace(*flagGitDir, "~", u.HomeDir, 1)
	} else {
		return *flagGitDir
	}
}
