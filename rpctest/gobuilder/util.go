// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package gobuilder

import (
	"runtime"
	"github.com/decred/dcrd/rpctest"
	"fmt"
	"strings"
)

func DetermineProjectPackagePath(projectName string) string {
	// Determine import path of this package.
	_, launchDir, _, ok := runtime.Caller(1)
	if !ok {
		rpctest.CheckTestSetupMalfunction(fmt.Errorf("Cannot get project path, launch dir is: %v ", launchDir))
	}
	sep := "/"
	steps := strings.Split(launchDir, sep)
	for i, s := range steps {
		if s == projectName {
			pkgPath := strings.Join(steps[:i+1], "/")
			return pkgPath
		}
	}
	rpctest.CheckTestSetupMalfunction(fmt.Errorf("Cannot get project path, launch dir is: %v ", launchDir))
	return ""
}
