package main

import (
	"errors"

	"github.com/fatih/color"
)

type testStatus int

const (
	testStatusPass = iota
	testStatusFail
	testStatusRunning
	testStatusWaiting
)

func (r testStatus) String() string {
	switch r {
	case testStatusPass:
		return color.New(color.FgBlack, color.BgGreen).Sprint("PASS")
	case testStatusFail:
		return color.New(color.FgRed, color.BgBlack).Sprint("FAIL")
	case testStatusRunning:
		return color.New(color.FgBlue, color.BgBlack).Sprint("RUNNING")
	case testStatusWaiting:
		return color.New(color.FgBlue, color.BgBlack).Sprint("WAITING")
	}
	panic(errors.New("unknown result type"))
}
