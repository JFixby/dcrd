// Copyright (c) 2018 The btcsuite developers
// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package memwallet

import (
	"github.com/decred/dcrd/chaincfg"
	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/hdkeychain"
	"github.com/decred/dcrd/integration"
	"github.com/decred/dcrd/integration/harness"
	"github.com/decred/dcrd/wire"
)

// WalletFactory produces a new InMemoryWallet-instance upon request
type WalletFactory struct {
}

// NewWallet creates and returns a fully initialized instance of the InMemoryWallet.
func (f *WalletFactory) NewWallet(cfg *harness.TestWalletConfig) harness.TestWallet {
	integration.AssertNotNil("ActiveNet", cfg.ActiveNet)
	w, e := newMemWallet(cfg.ActiveNet, cfg.Seed)
	integration.CheckTestSetupMalfunction(e)
	return w
}

func newMemWallet(net *chaincfg.Params, harnessHDSeed [chainhash.HashSize + 4]byte) (*InMemoryWallet, error) {
	hdRoot, err := hdkeychain.NewMaster(harnessHDSeed[:], net)
	if err != nil {
		return nil, nil
	}

	// The first child key from the hd root is reserved as the coinbase
	// generation address.
	coinbaseChild, err := hdRoot.Child(0)
	if err != nil {
		return nil, err
	}
	coinbaseKey, err := coinbaseChild.ECPrivKey()
	if err != nil {
		return nil, err
	}
	coinbaseAddr, err := keyToAddr(coinbaseKey, net)
	if err != nil {
		return nil, err
	}

	// Track the coinbase generation address to ensure we properly track
	// newly generated coins we can spend.
	addrs := make(map[uint32]dcrutil.Address)
	addrs[0] = coinbaseAddr

	return &InMemoryWallet{
		net:               net,
		coinbaseKey:       coinbaseKey,
		coinbaseAddr:      coinbaseAddr,
		hdIndex:           1,
		hdRoot:            hdRoot,
		addrs:             addrs,
		utxos:             make(map[wire.OutPoint]*utxo),
		chainUpdateSignal: make(chan string),
		reorgJournal:      make(map[int64]*undoEntry),
	}, nil
}
