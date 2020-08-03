package main

import (
	"errors"

	"github.com/fatih/color"
)

type testResult int

const (
	testResultPass = iota
	testResultFail
	testResultRunning
	testResultWaiting
)

func (r testResult) String() string {
	switch r {
	case testResultPass:
		return color.New(color.FgBlack, color.BgGreen).Sprint("PASS")
	case testResultFail:
		return color.New(color.FgRed, color.BgBlack).Sprint("FAIL")
	case testResultRunning:
		return color.New(color.FgBlue, color.BgBlack).Sprint("RUNNING")
	case testResultWaiting:
		return color.New(color.FgBlue, color.BgBlack).Sprint("WAITING")
	}
	panic(errors.New("unknown result type"))
}

var testStatus map[string]testResult

func setTestStatus(hash string, message testResult) {
	if testStatus == nil {
		testStatus = map[string]testResult{}
	}

	testStatus[hash] = message
}

func getTestStatus(hash string) testResult {
	result, ok := testStatus[hash]
	if ok {
		return result
	}
	return testResultWaiting
}
