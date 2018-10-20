// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package dcrtest

// disposableAssetsList keeps track of all assets and resources
// created by test setup execution for the following disposal.
//
// This includes removing all temporary directories,
// and shutting down any created processes.
var disposableAssetsList []DisposableAsset

// RegisterDisposableAsset registers disposable asset
func RegisterDisposableAsset(a DisposableAsset) {
	disposableAssetsList = append(disposableAssetsList, a)
}

// DisposableAsset is a handler for disposable resources
type DisposableAsset interface {
	// Dispose called by test setup before exit
	Dispose()
}
