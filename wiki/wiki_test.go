package wiki

import "testing"

func TestIsExists(t *testing.T) {
	actualFilePath := isExist("/usr/bin/env")
	if !actualFilePath {
		t.Error("/usr/bin/env should exist.")
	}

	dummyFilePath := isExist("/usr/bin/no-such-file")
	if dummyFilePath {
		t.Error("/usr/bin/no-such-file should not exist.")
	}
}
