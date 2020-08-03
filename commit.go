package main

import (
	"log"
	"os/exec"
	"strings"

	"github.com/gammazero/workerpool"
	"github.com/pkg/errors"
)

type commit struct {
	hash      string
	shortHash string
	subject   string
}

func getCommit(hash string) commit {
	return commit{
		hash:      hash,
		shortHash: gitGetOutput("log", "-1", "--format=%h", hash),
		subject:   gitGetOutput("log", "-1", "--format=%s", hash),
	}
}

func getCommits(commits string) []commit {
	// Put hashes in graph-order from "oldest" ancestor to "youngest" descendent
	// i.e. Reverse the slice
	hashes := reverse(revList(commits))

	pool := workerpool.New(5)

	output := make([]commit, len(hashes))
	for i := range hashes {
		i := i
		pool.Submit(func() {
			output[i] = getCommit(hashes[i])
		})
	}

	pool.StopWait()

	return output
}

func reverse(input []string) []string {
	output := []string{}
	for i := len(input) - 1; i >= 0; i-- {
		output = append(output, input[i])
	}
	return output
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
