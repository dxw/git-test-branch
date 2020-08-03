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

var testResults map[string]testResult

func setTestResult(hash string, message testResult) {
	if testResults == nil {
		testResults = map[string]testResult{}
	}

	testResults[hash] = message
}

func getTestResult(hash string) testResult {
	result, ok := testResults[hash]
	if ok {
		return result
	}
	return testResultWaiting
}
