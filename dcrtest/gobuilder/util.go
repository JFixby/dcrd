// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package gobuilder

import (
	"fmt"
	"github.com/decred/dcrd/dcrtest"
	"runtime"
	"strings"
)

func DetermineProjectPackagePath(projectName string) string {
	// Determine import path of this package.
	_, launchDir, _, ok := runtime.Caller(1)
	if !ok {
		dcrtest.CheckTestSetupMalfunction(fmt.Errorf("Cannot get project path, launch dir is: %v ", launchDir))
	}
	sep := "/"
	steps := strings.Split(launchDir, sep)
	for i, s := range steps {
		if s == projectName {
			pkgPath := strings.Join(steps[:i+1], "/")
			return pkgPath
		}
	}
	dcrtest.CheckTestSetupMalfunction(fmt.Errorf("Cannot get project path, launch dir is: %v ", launchDir))
	return ""
}
