package main

import (
	"log"
	"os/exec"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

func gitGetOutput(command ...string) string {
	cmd := exec.Command("git", command...)
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(errors.Wrap(err, "gitGetOutput"))
	}

	return strings.TrimSpace(string(output))
}

var gitGetOutputCache sync.Map

func gitGetOutputCached(command ...string) string {
	joined := strings.Join(command, " ")
	output, ok := gitGetOutputCache.Load(joined)
	if !ok {
		output = gitGetOutput(command...)
		gitGetOutputCache.Store(joined, output)
	}

	return output.(string)
}
