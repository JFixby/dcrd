package memwallet

import (
	"github.com/decred/dcrd/chaincfg"
	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrtest"
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/hdkeychain"
	"github.com/decred/dcrd/rpctest/testharness"
	"github.com/decred/dcrd/wire"
)

type MemWalletFactory struct {
}

func (f *MemWalletFactory) NewWallet(cfg *testharness.DcrdWalletConfig) testharness.DcrWallet {
	dcrtest.AssertNotNil("ActiveNet", cfg.ActiveNet)
	w, e := newMemWallet(cfg.ActiveNet, cfg.Seed)
	dcrtest.CheckTestSetupMalfunction(e)
	return w
}

// newMemWallet creates and returns a fully initialized instance of the
// InMemoryWallet given a particular blockchain's parameters.
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
