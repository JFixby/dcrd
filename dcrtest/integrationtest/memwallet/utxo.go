package memwallet

import "github.com/decred/dcrd/dcrutil"

// utxo represents an unspent output spendable by the InMemoryWallet. The maturity
// height of the transaction is recorded in order to properly observe the
// maturity period of direct coinbase outputs.
type utxo struct {
	pkScript       []byte
	value          dcrutil.Amount
	maturityHeight int64
	keyIndex       uint32
	isLocked       bool
}

// isMature returns true if the target utxo is considered "mature" at the
// passed block height. Otherwise, false is returned.
func (u *utxo) isMature(height int64) bool {
	return height >= u.maturityHeight
}
