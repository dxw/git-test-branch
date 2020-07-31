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

	return commitHashes
}

func runTest(command, hash string) error {
	root := getRootDir()

	err := os.MkdirAll(root, 0755)
	if err != nil {
		return errors.Wrap(err, "runTest: failed to create build directory root")
	}

	err = setNote(hash, "RUNNING")
	if err != nil {
		return errors.Wrap(err, "runTest: failed setting note")
	}

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
		err = setNote(hash, "PASS")
		if err != nil {
			return errors.Wrap(err, "runTest: failed setting PASS note")
		}
	} else {
		err = setNote(hash, "FAIL")
		if err != nil {
			return errors.Wrap(err, "runTest: failed setting FAIL note")
		}
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

func setNote(hash, message string) error {
	err := runExclusively(func() error {
		cmd := exec.Command("git", "notes", "--ref=refs/notes/git-test-branch", "add", "--force", fmt.Sprintf("--message=%s", message), hash)
		err := cmd.Run()
		if err != nil {
			return errors.Wrap(err, "setNote: failed git command")
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "setNote: failed running exclusively")
	}
	return nil
}

func showResults(hashes []string) {
	for _, hash := range hashes {
		outputHash := gitGetOutput("log", "-1", "--format=%h", hash)
		outputResult := gitGetOutput("notes", "--ref=refs/notes/git-test-branch", "show", hash)
		if outputResult == "" {
			outputResult = "WAITING"
		}
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
