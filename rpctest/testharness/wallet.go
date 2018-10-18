package testharness

import (
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/wire"
	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/rpcclient"
	"github.com/decred/dcrd/chaincfg"
)

type DcrWallet interface {
	GetNewAddress(account string) (dcrutil.Address, error)
	Dispose() error
	Launch(args *DcrWalletLaunchArgs) error
	Shutdown()
	Sync()
	ConfirmedBalance() dcrutil.Amount
	CreateTransaction(outputs []*wire.TxOut, feeRate dcrutil.Amount) (*wire.MsgTx, error)
	UnlockOutputs(inputs []*wire.TxIn)
	SendOutputs(outputs []*wire.TxOut, feeRate dcrutil.Amount) (*chainhash.Hash, error)
}

type DcrdWalletConfig struct {
	Seed          [chainhash.HashSize + 4]byte
	WalletRPCHost string
	WalletRPCPort int
	ActiveNet     *chaincfg.Params
}

type DcrWalletFactory interface {
	NewWallet(cfg *DcrdWalletConfig) DcrWallet
}

type DcrWalletLaunchArgs struct {
	DcrdCertFile             string
	WalletExtraArguments     map[string]interface{}
	DebugWalletOutput        bool
	MaxSecondsToWaitOnLaunch int
	DcrdRPCConfig            *rpcclient.ConnConfig
}
