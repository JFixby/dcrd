// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package gobuilder

import (
	"testing"
	"io/ioutil"
	"fmt"
	"os"
	"github.com/decred/dcrd/rpctest"
)

// TestGoBuider builds current project executable
func TestGoBuider(t *testing.T) {
	testWorkingDir, err := ioutil.TempDir("", "gobuid-test-")
	if err != nil {
		fmt.Println("Unable to create working dir: ", err)
		os.Exit(-1)
	}

	cfg := &GoBuiderConfig{
		GoProjectPath:    DetermineProjectPackagePath("dcrd"),
		OutputFolderPath: testWorkingDir,
		BuidFileName:     "dcrd",
	}

	builder := NewGoBuider(cfg)

	builder.Build()
	
	rpctest.DeleteFile(testWorkingDir)
}
