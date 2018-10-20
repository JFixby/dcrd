// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package gobuilder

import (
	"testing"
	"github.com/decred/dcrd/dcrtest"
	"io/ioutil"
	"fmt"
	"os"
)

// TestGoBuider builds current project executable
func TestGoBuider(t *testing.T) {
	defer dcrtest.VerifyNoResourcesLeaked()
	runExample()
}

func runExample() {
	testWorkingDir, err := ioutil.TempDir("", "gobuid-test-")
	if err != nil {
		fmt.Println("Unable to create working dir: ", err)
		os.Exit(-1)
	}
	defer dcrtest.DeleteFile(testWorkingDir)

	builder := &GoBuider{
		GoProjectPath:    DetermineProjectPackagePath("dcrd"),
		OutputFolderPath: testWorkingDir,
		BuildFileName:    "dcrd",
	}

	builder.Build()
	builder.Dispose()
}
