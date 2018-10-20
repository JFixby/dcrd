// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dcrtest

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
)

// disposableAssetsList keeps track of all assets and resources
// created by test setup execution for the following disposal.
//
// This includes removing all temporary directories,
// and shutting down any created processes.
var disposableAssetsList map[DisposableAsset]bool

// RegisterDisposableAsset registers disposable asset
func RegisterDisposableAsset(a DisposableAsset) {
	if disposableAssetsList == nil {
		disposableAssetsList = make(map[DisposableAsset]bool)
	}
	disposableAssetsList[a] = true
}

func DeRegisterDisposableAsset(a DisposableAsset) {
	delete(disposableAssetsList, a)
}

// DisposableAsset is a handler for disposable resources
type DisposableAsset interface {
	// Dispose called by test setup before exit
	Dispose()
}

func VerifyNoResourcesLeaked() {
	if len(disposableAssetsList) != 0 {
		ReportTestSetupMalfunction(
			fmt.Errorf(
				"incorrect state: resources leak detected: %v ",
				spew.Sdump(disposableAssetsList),
			),
		)
	}
}
