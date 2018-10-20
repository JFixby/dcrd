// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package commandline

import "testing"

func TestGoExample(t *testing.T) {
	proc := &ExternalProcess{
		CommandName: "go",
		WaitForExit: true,
	}
	proc.Arguments = append(proc.Arguments, "-v")
	proc.Arguments = append(proc.Arguments, "help")

	debugOutput := true
	proc.Launch(debugOutput)
}
