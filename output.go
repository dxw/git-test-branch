package main

import (
	"log"
	"os/exec"
	"strings"

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

var gitGetOutputCache map[string]string

func gitGetOutputCached(command ...string) string {
	if gitGetOutputCache == nil {
		gitGetOutputCache = map[string]string{}
	}

	joined := strings.Join(command, " ")
	output, ok := gitGetOutputCache[joined]
	if ok {
		return output
	}

	output = gitGetOutput(command...)
	gitGetOutputCache[joined] = output
	return output
}
