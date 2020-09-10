package main

import (
	"fmt"
	"log"
	"os"

	"os/exec"
	"path"
	"strings"
	"sync"

	"github.com/dxw/git-test-branch/screenwriter"
	"github.com/gammazero/workerpool"
	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
)

var mutex sync.Mutex

func main() {
	flags := flag.NewFlagSet("git test-branch", flag.ContinueOnError)
	flags.Usage = func() {
		fmt.Printf("Usage: git test-branch main..@ 'command-to-run'\n\n")
		fmt.Printf("Options:\n")
		flags.PrintDefaults()
	}

	err := flags.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		// Do nothing
	} else if err != nil {
		log.Fatal(err)
	}

	if len(flags.Args()) < 2 {
		flags.Usage()
		os.Exit(1)
	}

	commitSpecification := flags.Args()[0]
	command := flags.Args()[1]
	rootDir := getRootDir()

	commits := getCommits(commitSpecification)
	// This only needs a small buffer because showResults will be reading it constantly
	results := make(chan testResult, 10)

	pool := workerpool.New(5)

	// Process failures
	failures := make(chan bool, 10)
	aCommandFailed := false
	go func() {
		for <-failures {
			aCommandFailed = true
		}
	}()

	// Show the results before we do anything, and update them regularly
	go showResults(commits, results)

	for _, commit := range commits {
		// This line is necessary because otherwise `commit`'s contents change
		commit := commit
		pool.Submit(func() {
			err := runTest(rootDir, command, commit, results, failures)
			if err != nil {
				log.Fatal(errors.Wrap(err, "failure in workerpool task"))
			}
		})
	}

	pool.StopWait()

	close(results)
	close(failures)

	if aCommandFailed {
		os.Exit(1)
	}
}

func runTest(root string, command string, commit commit, results chan<- testResult, failures chan<- bool) error {
	err := os.MkdirAll(root, 0755)
	if err != nil {
		return errors.Wrap(err, "runTest: failed to create build directory root")
	}

	results <- testResult{commit: commit, status: testStatusRunning}

	commitDir := path.Join(root, commit.hash)

	err = os.RemoveAll(commitDir)
	if err != nil {
		return errors.Wrap(err, "runTest: failed removing commitDir")
	}

	err = runExclusively(func() error {
		cmd := exec.Command("git", "worktree", "add", "--force", "--detach", commitDir, commit.hash)
		err := cmd.Run()
		if err != nil {
			return errors.Wrap(err, "runTest: failed running worktree add command")
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "runTest: failed running exclusively")
	}

	cmd := exec.Command("sh", "-c", fmt.Sprintf("cd %s && %s", commitDir, command))
	err = cmd.Run()

	if err == nil {
		results <- testResult{commit: commit, status: testStatusPass}
	} else {
		results <- testResult{commit: commit, status: testStatusFail}
		failures <- true
	}

	err = runExclusively(func() error {
		cmd = exec.Command("git", "worktree", "remove", "--force", commitDir)
		err = cmd.Run()
		if err != nil {
			return errors.Wrap(err, "runTest: failed running worktree remove command")
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "runTest: failed running exclusively")
	}

	return nil
}

func runExclusively(f func() error) error {
	defer mutex.Unlock()
	mutex.Lock()
	err := f()
	if err != nil {
		return errors.Wrap(err, "runExclusively: f() returned non-nil error")
	}
	return nil
}

func showResults(commits []commit, results <-chan testResult) {
	finalResults := map[string]testResult{}

	// Populate finalResults with some placeholder values
	for _, commit := range commits {
		finalResults[commit.hash] = testResult{commit: commit, status: testStatusWaiting}
	}

	// Update finalResults with the correct values as they come in from the channel
	// These get updated continuously
	update := make(chan bool, 1)
	go func() {
		for thisResult := range results {
			finalResults[thisResult.commit.hash] = thisResult
			update <- true
		}
		// When results is closed, close update as well
		close(update)
	}()

	w := screenwriter.New(os.Stdout)

	// Display finalResults immediately, then whenever they're updated
	showResultsOnce(w, commits, finalResults)
	for range update {
		showResultsOnce(w, commits, finalResults)
	}
}

func showResultsOnce(w *screenwriter.ScreenWriter, commits []commit, finalResults map[string]testResult) {
	text := ""
	for _, commit := range commits {
		result := finalResults[commit.hash]
		text += fmt.Sprintf("%s [%s] %s\n", result.commit.shortHash, result.status, result.commit.subject)
	}
	err := w.Display(text)
	if err != nil {
		log.Fatal(err)
	}
}

func gitGetOutput(command ...string) string {
	cmd := exec.Command("git", command...)
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(errors.Wrap(err, "gitGetOutput"))
	}

	return strings.TrimSpace(string(output))
}

func getRootDir() string {
	dir := gitGetOutput("rev-parse", "--absolute-git-dir")

	return path.Join(dir, "git-test-branch")
}
