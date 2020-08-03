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

	commits := os.Args[1]
	command := os.Args[2]

	commitHashes := revList(commits)

	pool := workerpool.New(5)

	for _, hash := range commitHashes {
		// This line is necessary because otherwise `hash`'s contents change
		hash := hash
		pool.Submit(func() {
			err := runTest(command, hash)
			if err != nil {
				log.Fatal(errors.Wrap(err, "failure in workerpool task"))
			}
		})
	}

	pool.StopWait()

	showResults(commitHashes)
}

func revList(commits string) []string {
	cmd := exec.Command("git", "rev-list", commits)
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(errors.Wrap(err, "revList"))
	}

	commitHashes := []string{}
	for _, s := range strings.Split(string(output), "\n") {
		if s == "" {
			continue
		}
		commitHashes = append(commitHashes, s)
	}

	// Put hashes in graph-order from "oldest" ancestor to "youngest" descendent
	// i.e. Reverse the slice
	commitHashesReversed := []string{}
	for i := len(commitHashes) - 1; i >= 0; i-- {
		commitHashesReversed = append(commitHashesReversed, commitHashes[i])
	}

	return commitHashesReversed
}

func runTest(command, hash string) error {
	root := getRootDir()

	err := os.MkdirAll(root, 0755)
	if err != nil {
		return errors.Wrap(err, "runTest: failed to create build directory root")
	}

	setTestStatus(hash, testResultRunning)

	commitDir := path.Join(root, hash)

	err = os.RemoveAll(commitDir)
	if err != nil {
		return errors.Wrap(err, "runTest: failed removing commitDir")
	}

	err = runExclusively(func() error {
		cmd := exec.Command("git", "worktree", "add", "--force", "--detach", commitDir, hash)
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
		setTestStatus(hash, testResultPass)
	} else {
		setTestStatus(hash, testResultFail)
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

func showResults(hashes []string) {
	for _, hash := range hashes {
		outputHash := gitGetOutput("log", "-1", "--format=%h", hash)
		outputResult := getTestStatus(hash)
		outputSubject := gitGetOutput("log", "-1", "--format=%s", hash)

		fmt.Printf("%s [%s] %s\n", outputHash, outputResult, outputSubject)
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
	cmd := exec.Command("git", "rev-parse", "--absolute-git-dir")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(errors.Wrap(err, "getRootDir"))
	}

	dir := strings.TrimSpace(string(output))

	return path.Join(dir, "git-test-branch")
}
