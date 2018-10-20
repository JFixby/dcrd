package regressiontest

import (
	"github.com/decred/dcrd/chaincfg"
	"github.com/decred/dcrd/dcrtest"
	"github.com/decred/dcrd/rpctest/testharness"
	"github.com/jfixby/pin"
	"path/filepath"
)

// ChainWithMatureOutputsSpawner initializes the primary mining node
// with a test chain of desired height, providing numMatureOutputs coinbases
// to allow spending from for testing purposes.
type ChainWithMatureOutputsSpawner struct {
	// Each harness will be provided with a dedicated
	// folder inside the WorkingDir
	WorkingDir string

	// DebugDCRDOutput, set true to print out dcrd output to console
	DebugDCRDOutput bool

	// DebugWalletOutput, set true to print out wallet output to console
	DebugWalletOutput bool

	// newHarnessIndex for net port offset
	newHarnessIndex int

	// Harnesses will subsequently reserve
	// network ports starting from the BasePort value
	BasePort int

	// NumMatureOutputs sets requirement for the generated test chain
	NumMatureOutputs uint32

	DcrdFactory   testharness.DcrdNodeFactory
	WalletFactory testharness.DcrWalletFactory

	ActiveNet *chaincfg.Params
}

// NewInstance does the following:
//   1. Starts a new DcrdTestServer process with a fresh SimNet chain.
//   2. Creates a new temporary WalletTestServer connected to the running DcrdTestServer.
//   3. Gets a new address from the WalletTestServer for mining subsidy.
//   4. Restarts the DcrdTestServer with the new mining address.
//   5. Generates a number of blocks so that testing starts with a spendable
//      balance.
func (testSetup *ChainWithMatureOutputsSpawner) NewInstance(harnessName string) dcrtest.Spawnable {
	harnessFolderName := "harness-" + harnessName
	pin.D("ActiveNet", testSetup.ActiveNet)
	dcrtest.AssertNotNil("DcrdFactory", testSetup.DcrdFactory)
	dcrtest.AssertNotNil("ActiveNet", testSetup.ActiveNet)
	dcrtest.AssertNotNil("WalletFactory", testSetup.WalletFactory)

	seedIndex := extractSeedIndexFromHarnessName(harnessName)

	harnessFolder := filepath.Join(testSetup.WorkingDir, harnessFolderName)

	p2p, dcrdRPC, walletRPC := generateListeningPorts(
		testSetup.newHarnessIndex, testSetup.BasePort)
	testSetup.newHarnessIndex++

	localhost := "127.0.0.1"

	dcrdConfig := &testharness.DcrdNodeConfig{
		P2PHost: localhost,
		P2PPort: p2p,

		DcrdRPCHost: localhost,
		DcrdRPCPort: dcrdRPC,

		ActiveNet: testSetup.ActiveNet,

		WorkingDir: harnessFolder,
	}

	walletConfig := &testharness.DcrdWalletConfig{
		Seed:          testharness.NewTestSeed(seedIndex),
		WalletRPCHost: localhost,
		WalletRPCPort: walletRPC,
		ActiveNet:     testSetup.ActiveNet,
	}

	harness := &testharness.Harness{
		DcrdServer: testSetup.DcrdFactory.NewNode(dcrdConfig),
		Wallet:     testSetup.WalletFactory.NewWallet(walletConfig),
		WorkingDir: harnessFolder,
	}

	DeploySimpleChain(testSetup, harness)

	return harness
}

// Dispose harness. This includes removing
// all temporary directories, and shutting down any created processes.
func (testSetup *ChainWithMatureOutputsSpawner) Dispose(s dcrtest.Spawnable) error {
	h := s.(*testharness.Harness)
	if h == nil {
		return nil
	}
	h.Wallet.Dispose()
	h.DcrdServer.Dispose()
	return h.DeleteWorkingDir()
}

// NameForTag defines policy for mapping input tags to harness names
func (testSetup *ChainWithMatureOutputsSpawner) NameForTag(tag string) string {
	harnessName := tag
	return harnessName
}
