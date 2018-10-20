// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dcrtest

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
)

// LeakyResource is a handler for disposable resources
// This includes removing all temporary directories,
// and shutting down any created processes.
type LeakyResource interface {
	Dispose()
}

// leaksList keeps track of all leaky assets and resources
// created by test setup execution.
//
var leaksList map[LeakyResource]bool

// RegisterDisposableAsset registers disposable asset
// Is abs strict method, does not tolerate multiple appends of the same element
func RegisterDisposableAsset(a LeakyResource) {
	if leaksList[a] == true {
		ReportTestSetupMalfunction(
			fmt.Errorf("LeakyResource is already registered: %v ",
				a,
			),
		)
	}
	if leaksList == nil {
		leaksList = make(map[LeakyResource]bool)
	}
	leaksList[a] = true
}

// DeRegisterDisposableAsset removes disposable asset from list
// Is abs strict method, does not tolerate multiple removals of the same element
func DeRegisterDisposableAsset(a LeakyResource) {
	if leaksList[a] == false {
		ReportTestSetupMalfunction(
			fmt.Errorf("LeakyResource is not registered: %v ",
				a,
			),
		)
	}
	delete(leaksList, a)
}

// VerifyNoResourcesLeaked is a abs strict method, checks all leaky
// resources were properly disposed. Should be called before test setup exit.
func VerifyNoResourcesLeaked() {
	if len(leaksList) != 0 {
		ReportTestSetupMalfunction(
			fmt.Errorf(
				"incorrect state: resources leak detected: %v ",
				spew.Sdump(leaksList),
			),
		)
	}
}
