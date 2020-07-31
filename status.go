package main

var testStatus map[string]string

func setTestStatus(hash, message string) {
	if testStatus == nil {
		testStatus = map[string]string{}
	}

	testStatus[hash] = message
}

func getTestStatus(hash string) string {
	return testStatus[hash]
}
