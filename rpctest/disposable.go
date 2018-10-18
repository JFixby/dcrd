// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpctest

func RegisterDisposableAsset(a DisposableAsset) {
	DisposableAssetsList = append(DisposableAssetsList, a)
}

var DisposableAssetsList []DisposableAsset

type DisposableAsset interface {
	Dispose()
}
