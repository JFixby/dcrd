// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package commandline

import (
	"testing"
	"github.com/decred/dcrd/dcrtest"
)

func TestGoExample(t *testing.T) {
	defer dcrtest.VerifyNoResourcesLeaked()
	proc := &ExternalProcess{
		CommandName: "go",
		WaitForExit: true,
	}
	proc.Arguments = append(proc.Arguments, "help")

	debugOutput := true
	proc.Launch(debugOutput)
}
