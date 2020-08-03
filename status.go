package main

import "errors"

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
		return "PASS"
	case testResultFail:
		return "FAIL"
	case testResultRunning:
		return "RUNNING"
	case testResultWaiting:
		return "WAITING"
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
