package regressiontest

import (
	"fmt"
	"github.com/decred/dcrd/dcrtest"
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/rpctest/testharness"
	"strconv"
	"strings"
)

func DeploySimpleChain(testSetup *ChainWithMatureOutputsSpawner, harness *testharness.Harness) {
	fmt.Println("Deploying Harness[" + harness.Name() + "]")

	// launch a fresh harness (assumes harness working dir is empty)
	{
		args := &launchArguments{
			DebugDCRDOutput:   testSetup.DebugDCRDOutput,
			DebugWalletOutput: testSetup.DebugWalletOutput,
		}
		launchHarnessSequence(harness, args)
	}

	// Get a new address from the WalletTestServer
	// to be set with dcrd --miningaddr
	{
		address, err := harness.Wallet.GetNewAddress("default")
		dcrtest.CheckTestSetupMalfunction(err)
		harness.MiningAddress = address

		dcrtest.AssertNotNil("MiningAddress", harness.MiningAddress)
		dcrtest.AssertNotEmpty("MiningAddress", harness.MiningAddress.String())

		fmt.Println("Mining address: " + harness.MiningAddress.String())
	}

	// restart the harness with the new argument
	{
		shutdownHarnessSequence(harness)

		args := &launchArguments{
			DebugDCRDOutput:   testSetup.DebugDCRDOutput,
			DebugWalletOutput: testSetup.DebugWalletOutput,
		}
		{
			// set create test chain with numMatureOutputs
			args.CreateTestChain = true
			args.NumMatureOutputs = testSetup.NumMatureOutputs
			args.CoinbaseMaturity = testSetup.ActiveNet.CoinbaseMaturity
		}
		launchHarnessSequence(harness, args)
	}

	// wait for the WalletTestServer to sync up to the current height
	{
		//desiredHeight := int64(
		//	testSetup.NumMatureOutputs +
		//		uint32(cfg.ActiveNet.CoinbaseMaturity))
		harness.Wallet.Sync()
	}

	fmt.Println("Harness[" + harness.Name() + "] is ready")

}

// local struct to bundle launchHarnessSequence function arguments
type launchArguments struct {
	DebugDCRDOutput   bool
	DebugWalletOutput bool
	CreateTestChain   bool
	NumMatureOutputs  uint32
	CoinbaseMaturity  uint16
	MiningAddress     *dcrutil.Address
}

// launchHarnessSequence
// 1. launch Dcrd node
// 2. connect to the node via RPC client
// 3. launch wallet and connects it to the Dcrd node
// 4. connect to the wallet via RPC client
func launchHarnessSequence(harness *testharness.Harness, args *launchArguments) {
	dcrd := harness.DcrdServer
	wallet := harness.Wallet

	dcrdLaunchArguments := &testharness.DcrdLaunchArgs{
		DebugOutput:   args.DebugDCRDOutput,
		MiningAddress: harness.MiningAddress,
	}
	dcrd.Launch(dcrdLaunchArguments)

	rpcConfig := dcrd.RPCConnectionConfig()

	walletLaunchArguments := &testharness.DcrWalletLaunchArgs{
		DcrdCertFile:             dcrd.CertFile(),
		DebugWalletOutput:        args.DebugWalletOutput,
		MaxSecondsToWaitOnLaunch: 90,
		DcrdRPCConfig:            rpcConfig,
	}
	wallet.Launch(walletLaunchArguments)

	if args.CreateTestChain {
		numToGenerate := uint32(args.CoinbaseMaturity) + args.NumMatureOutputs
		err := generateTestChain(numToGenerate, harness.DcrdRPCClient())
		dcrtest.CheckTestSetupMalfunction(err)
	}
}

// shutdownHarnessSequence reverses the launchHarnessSequence
// 4. disconnect from the wallet RPC
// 3. stop wallet
// 2. disconnect from the Dcrd RPC
// 1. stop Dcrd node
func shutdownHarnessSequence(harness *testharness.Harness) {
	harness.Wallet.Shutdown()
	harness.DcrdServer.Shutdown()
}

// generateListeningPorts returns 3 subsequent network ports starting from base
func generateListeningPorts(index, base int) (int, int, int) {
	x := base + index*3 + 0
	y := base + index*3 + 1
	z := base + index*3 + 2
	return x, y, z
}

func extractSeedIndexFromHarnessName(harnessName string) uint32 {
	parts := strings.Split(harnessName, ".")
	if len(parts) != 2 {
		return 0
	}
	seedString := parts[1]
	tmp, err := strconv.Atoi(seedString)
	seedIndex := uint32(tmp)
	dcrtest.CheckTestSetupMalfunction(err)
	return seedIndex
}
