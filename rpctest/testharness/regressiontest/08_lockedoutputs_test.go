package regressiontest

import (
	"github.com/decred/dcrd/dcrutil"
	"testing"
	"github.com/decred/dcrd/txscript"
	"github.com/decred/dcrd/wire"
)

func TestMemWalletLockedOutputs(t *testing.T) {
	// Skip tests when running with -short
	if testing.Short() {
		t.Skip("Skipping RPC harness tests in short mode")
	}
	if skipTest(t) {
		t.Skip("Skipping test")
	}
	r := ObtainHarness(MainHarnessName)

	// Obtain the initial balance of the wallet at this point.
	startingBalance := r.Wallet.ConfirmedBalance()

	// First, create a signed transaction spending some outputs.
	addr, err := r.Wallet.GetNewAddress("default")
	if err != nil {
		t.Fatalf("unable to generate new address: %v", err)
	}
	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		t.Fatalf("unable to create script: %v", err)
	}
	outputAmt := dcrutil.Amount(50 * dcrutil.AtomsPerCoin)
	output := wire.NewTxOut(int64(outputAmt), pkScript)
	tx, err := r.Wallet.CreateTransaction([]*wire.TxOut{output}, 10)
	if err != nil {
		t.Fatalf("unable to create transaction: %v", err)
	}

	// The current wallet balance should now be at least 50 Coin less
	// (accounting for fees) than the period balance
	currentBalance := r.Wallet.ConfirmedBalance()
	if !(currentBalance <= startingBalance-outputAmt) {
		t.Fatalf("spent outputs not locked: previous balance %v, "+
			"current balance %v", startingBalance, currentBalance)
	}

	// Now unlocked all the spent inputs within the unbroadcast signed
	// transaction. The current balance should now be exactly that of the
	// starting balance.
	r.Wallet.UnlockOutputs(tx.TxIn)
	currentBalance = r.Wallet.ConfirmedBalance()
	if currentBalance != startingBalance {
		t.Fatalf("current and starting balance should now match: "+
			"expected %v, got %v", startingBalance, currentBalance)
	}
}
