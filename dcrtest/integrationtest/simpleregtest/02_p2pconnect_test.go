package simpleregtest

import (
	"testing"
	"github.com/decred/dcrd/dcrtest/integrationtest"
)

func TestP2PConnect(t *testing.T) {
	// Skip tests when running with -short
	if testing.Short() {
		t.Skip("Skipping RPC harness tests in short mode")
	}
	if skipTest(t) {
		t.Skip("Skipping test")
	}
	r := ObtainHarness(MainHarnessName)

	// Create a fresh test harness.
	harness := harnessWithZeroMOSpawner.NewInstance("TestP2PConnect").(*integrationtest.Harness)
	defer harnessWithZeroMOSpawner.Dispose(harness)

	// Establish a p2p connection from our new local harness to the main
	// harness.
	if err := ConnectNode(harness, r); err != nil {
		t.Fatalf("unable to connect local to main harness: %v", err)
	}

	// The main harness should show up in our local harness' peer's list,
	// and vice verse.
	assertConnectedTo(t, harness, r)
}
