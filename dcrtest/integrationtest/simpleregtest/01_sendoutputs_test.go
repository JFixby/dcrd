package simpleregtest

import (
	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/txscript"
	"github.com/decred/dcrd/wire"
	"github.com/jfixby/pin"
	"testing"
	"github.com/decred/dcrd/dcrtest/integrationtest"
)

func genSpend(t *testing.T, r *integrationtest.Harness, amt dcrutil.Amount) *chainhash.Hash {
	// Grab a fresh address from the wallet.
	addr, err := r.Wallet.GetNewAddress("default")
	if err != nil {
		t.Fatalf("unable to get new address: %v", err)
	}

	// Next, send amt to this address, spending from one of our
	// mature coinbase outputs.
	addrScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		t.Fatalf("unable to generate pkscript to addr: %v", err)
	}
	output := wire.NewTxOut(int64(amt), addrScript)
	txid, err := r.Wallet.SendOutputs([]*wire.TxOut{output}, 10)
	if err != nil {
		t.Fatalf("coinbase spend failed: %v", err)
	}
	return txid
}

func assertTxMined(t *testing.T, r *integrationtest.Harness, txid *chainhash.Hash, blockHash *chainhash.Hash) {
	block, err := r.DcrdRPCClient().GetBlock(blockHash)
	if err != nil {
		t.Fatalf("unable to get block: %v", err)
	}

	numBlockTxns := len(block.Transactions)
	if numBlockTxns < 2 {
		t.Fatalf("crafted transaction wasn't mined, block should have "+
			"at least %v transactions instead has %v", 2, numBlockTxns)
	}

	minedTx := block.Transactions[1]
	txHash := minedTx.TxHash()
	if txHash != *txid {
		t.Fatalf("txid's don't match, %v vs %v", txHash, txid)
	}
}

func TestSendOutputs(t *testing.T) {
	// Skip tests when running with -short
	if testing.Short() {
		t.Skip("Skipping RPC harness tests in short mode")
	}
	if skipTest(t) {
		t.Skip("Skipping test")
	}
	r := ObtainHarness(MainHarnessName)
	r.Wallet.Sync()
	balance := r.Wallet.ConfirmedBalance()
	pin.D("balance", balance)
	pin.D("reward", r.DcrdServer.Network().BaseSubsidy/dcrutil.AtomsPerCoin)
	// First, generate a small spend which will require only a single
	// input.
	txid := genSpend(t, r, dcrutil.Amount(5*dcrutil.AtomsPerCoin))

	// Generate a single block, the transaction the wallet created should
	// be found in this block.
	blockHashes, err := r.DcrdRPCClient().Generate(1)
	if err != nil {
		t.Fatalf("unable to generate single block: %v", err)
	}
	assertTxMined(t, r, txid, blockHashes[0])

	// Next, generate a spend much greater than the block reward. This
	// transaction should also have been mined properly.
	txid = genSpend(t, r, dcrutil.Amount(1000*dcrutil.AtomsPerCoin))
	blockHashes, err = r.DcrdRPCClient().Generate(1)
	if err != nil {
		t.Fatalf("unable to generate single block: %v", err)
	}
	assertTxMined(t, r, txid, blockHashes[0])

	// Generate another block to ensure the transaction is removed from the
	// mempool.
	if _, err := r.DcrdRPCClient().Generate(1); err != nil {
		t.Fatalf("unable to generate block: %v", err)
	}
}
