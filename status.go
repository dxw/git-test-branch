package main

import (
	"errors"
	"sync"

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

var testResults sync.Map

func setTestResult(hash string, message testResult) {
	testResults.Store(hash, message)
}

func getTestResult(hash string) testResult {
	result, ok := testResults.Load(hash)
	if ok {
		return result.(testResult)
	}
	return testResultWaiting
}
