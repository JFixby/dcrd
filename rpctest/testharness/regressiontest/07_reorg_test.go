package regressiontest

import (
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/rpctest/testharness"
	"testing"
)

func TestMemWalletReorg(t *testing.T) {
	// Skip tests when running with -short
	if testing.Short() {
		t.Skip("Skipping RPC harness tests in short mode")
	}
	if skipTest(t) {
		t.Skip("Skipping test")
	}
	r := ObtainHarness("TestMemWalletReorg.0")

	// Create a fresh test harness.
	// Deploy harness spawner with empty test chain
	blocks := 5
	spawner := &ChainWithMatureOutputsSpawner{
		WorkingDir:        WorkingDir,
		DebugDCRDOutput:   false,
		DebugWalletOutput: false,
		NumMatureOutputs:  uint32(blocks + 1),
		BasePort:          35000,
		WalletFactory:     WalletFactory,
		DcrdFactory:       DcrdFactory,
		ActiveNet:         Network,
	}
	harness := spawner.NewInstance("TestMemWalletReorg.1").(*testharness.Harness)
	defer spawner.Dispose(harness)

	// Create a fresh harness, we'll be using the main harness to force a
	// re-org on this local harness.
	//harness := harnessWithZeroMOSpawner.NewInstance("TestMemWalletReorg.1").(*testharness.Harness)
	//defer harnessWithZeroMOSpawner.Dispose(harness)

	// Ensure the internal wallet has the expected balance.
	expectedBalance := dcrutil.Amount(blocks * 300 * dcrutil.AtomsPerCoin)
	harness.Wallet.Sync()
	walletBalance := harness.Wallet.ConfirmedBalance()
	if expectedBalance != walletBalance {
		t.Fatalf("wallet balance incorrect: expected %v, got %v",
			expectedBalance, walletBalance)
	}

	// Now connect this local harness to the main harness then wait for
	// their chains to synchronize.
	if err := ConnectNode(harness, r); err != nil {
		t.Fatalf("unable to connect harnesses: %v", err)
	}
	nodeSlice := []*testharness.Harness{r, harness}
	if err := JoinNodes(nodeSlice, Blocks); err != nil {
		t.Fatalf("unable to join node on blocks: %v", err)
	}

	// The original wallet should now have a balance of 0 Coin as its entire
	// chain should have been decimated in favor of the main harness'
	// chain.
	expectedBalance = dcrutil.Amount(0)
	walletBalance = harness.Wallet.ConfirmedBalance()
	if expectedBalance != walletBalance {
		t.Fatalf("wallet balance incorrect: expected %v, got %v",
			expectedBalance, walletBalance)
	}
}
