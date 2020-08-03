package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"

	"github.com/gammazero/workerpool"
	"github.com/pkg/errors"
)

var mutex sync.Mutex

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: git test-branch main..@ 'command-to-run'")
	}

	commitSpecification := os.Args[1]
	command := os.Args[2]

	commits := getCommits(commitSpecification)
	// This only needs a small buffer because showResults will be reading it constantly
	results := make(chan testResult, 10)

	pool := workerpool.New(5)

	// Show the results before we do anything, and update them regularly
	go showResults(commits, results)

	for _, commit := range commits {
		// This line is necessary because otherwise `commit`'s contents change
		commit := commit
		pool.Submit(func() {
			err := runTest(command, commit, results)
			if err != nil {
				log.Fatal(errors.Wrap(err, "failure in workerpool task"))
			}
		})
	}

	pool.StopWait()

	close(results)
}

func runTest(command string, commit commit, results chan<- testResult) error {
	root := getRootDir()

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
	}

	cmd = exec.Command("git", "worktree", "remove", "--force", commitDir)
	err = cmd.Run()
	if err != nil {
		return errors.Wrap(err, "runTest: failed running worktree remove command")
	}

	return nil
}

func runExclusively(f func() error) error {
	mutex.Lock()
	err := f()
	if err != nil {
		return errors.Wrap(err, "runExclusively: f() returned non-nil error")
	}
	mutex.Unlock()
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

	// Display finalResults immediately, then whenever they're updated
	showResultsOnce(commits, finalResults)
	for range update {
		showResultsOnce(commits, finalResults)
	}
}

func showResultsOnce(commits []commit, finalResults map[string]testResult) {
	for _, commit := range commits {
		result := finalResults[commit.hash]
		fmt.Printf("%s [%s] %s\n", result.commit.shortHash, result.status, result.commit.subject)
	}
	fmt.Println()
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
