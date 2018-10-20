// Copyright (c) 2016-2017 The btcsuite developers
// Copyright (c) 2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package memwallet

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/decred/dcrd/blockchain"
	"github.com/decred/dcrd/chaincfg"
	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrec/secp256k1"
	"github.com/decred/dcrd/dcrtest"
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/hdkeychain"
	"github.com/decred/dcrd/rpcclient"
	"github.com/decred/dcrd/rpctest/testharness"
	"github.com/decred/dcrd/txscript"
	"github.com/decred/dcrd/wire"
	"time"
)

const chainUpdateSignal = "chainUpdateSignal"
const stopSignal = "stopSignal"

// chainUpdate encapsulates an update to the current main chain. This struct is
// used to sync up the InMemoryWallet each time a new block is connected to the main
// chain.
type chainUpdate struct {
	blockHeight  int64
	filteredTxns []*dcrutil.Tx
}

// undoEntry is functionally the opposite of a chainUpdate. An undoEntry is
// created for each new block received, then stored in a log in order to
// properly handle block re-orgs.
type undoEntry struct {
	utxosDestroyed map[wire.OutPoint]*utxo
	utxosCreated   []wire.OutPoint
}

// InMemoryWallet is a simple in-memory wallet whose purpose is to provide basic
// wallet functionality to the harness. The wallet uses a hard-coded HD key
// hierarchy which promotes reproducibility between harness test runs.
type InMemoryWallet struct {
	coinbaseKey  *secp256k1.PrivateKey
	coinbaseAddr dcrutil.Address

	// hdRoot is the root master private key for the wallet.
	hdRoot *hdkeychain.ExtendedKey

	// hdIndex is the next available key index offset from the hdRoot.
	hdIndex uint32

	// currentHeight is the latest height the wallet is known to be synced
	// to.
	currentHeight int64

	// addrs tracks all addresses belonging to the wallet. The addresses
	// are indexed by their keypath from the hdRoot.
	addrs map[uint32]dcrutil.Address

	// utxos is the set of utxos spendable by the wallet.
	utxos map[wire.OutPoint]*utxo

	// reorgJournal is a map storing an undo entry for each new block
	// received. Once a block is disconnected, the undo entry for the
	// particular height is evaluated, thereby rewinding the effect of the
	// disconnected block on the wallet's set of spendable utxos.
	reorgJournal map[int64]*undoEntry

	chainUpdates      []*chainUpdate
	chainUpdateSignal chan string
	chainMtx          sync.Mutex

	net *chaincfg.Params

	dcrdRpc *rpcclient.Client

	sync.RWMutex
}

func (w *InMemoryWallet) GetNewAddress(account string) (dcrutil.Address, error) {
	return w.NewAddress()
	//return w.rpc.GetNewAddress(account)
}

func (wallet *InMemoryWallet) Launch(args *testharness.DcrWalletLaunchArgs) error {
	handlers := &rpcclient.NotificationHandlers{}

	// If a handler for the OnBlockConnected/OnBlockDisconnected callback
	// has already been set, then we create a wrapper callback which
	// executes both the currently registered callback, and the mem
	// wallet's callback.
	if handlers.OnBlockConnected != nil {
		obc := handlers.OnBlockConnected
		handlers.OnBlockConnected = func(header []byte, filteredTxns [][]byte) {
			wallet.IngestBlock(header, filteredTxns)
			obc(header, filteredTxns)
		}
	} else {
		// Otherwise, we can claim the callback ourselves.
		handlers.OnBlockConnected = wallet.IngestBlock
	}
	if handlers.OnBlockDisconnected != nil {
		obd := handlers.OnBlockDisconnected
		handlers.OnBlockDisconnected = func(header []byte) {
			wallet.UnwindBlock(header)
			obd(header)
		}
	} else {
		handlers.OnBlockDisconnected = wallet.UnwindBlock
	}

	//handlers.OnClientConnected = wallet.onDcrdConnect

	wallet.dcrdRpc = testharness.NewRPCConnection(args.DcrdRPCConfig, 5, handlers)
	dcrtest.AssertNotNil("dcrdRpc", wallet.dcrdRpc)

	// Filter transactions that pay to the coinbase associated with the
	// wallet.
	wallet.updateTxFilter()

	// Ensure dcrd properly dispatches our registered call-back for each new
	// block. Otherwise, the InMemoryWallet won't function properly.
	err := wallet.dcrdRpc.NotifyBlocks()
	dcrtest.CheckTestSetupMalfunction(err)

	go wallet.chainSyncer()
	return nil
}

func (wallet *InMemoryWallet) updateTxFilter() {
	filterAddrs := []dcrutil.Address{}
	for _, v := range wallet.addrs {
		filterAddrs = append(filterAddrs, v)
	}
	//pin.D("filterAddrs", filterAddrs)
	err := wallet.dcrdRpc.LoadTxFilter(true, filterAddrs, nil)
	dcrtest.CheckTestSetupMalfunction(err)
}

func (w *InMemoryWallet) Shutdown() {
	go func() {
		w.chainUpdateSignal <- stopSignal
	}()
	w.dcrdRpc.Disconnect()
	w.dcrdRpc = nil
}

func (w *InMemoryWallet) Sync() {
	// Block until the wallet has fully synced up to the tip of the main
	// chain.
	_, height, err := w.dcrdRpc.GetBestBlock()
	dcrtest.CheckTestSetupMalfunction(err)
	ticker := time.NewTicker(time.Millisecond * 100)
	for range ticker.C {
		walletHeight := w.SyncedHeight()
		if walletHeight == height {
			break
		}
	}
}

func (m *InMemoryWallet) Dispose() error {
	return nil
}

// SyncedHeight returns the height the wallet is known to be synced to.
//
// This function is safe for concurrent access.
func (m *InMemoryWallet) SyncedHeight() int64 {
	m.RLock()
	defer m.RUnlock()
	return m.currentHeight
}

// IngestBlock is a call-back which is to be triggered each time a new block is
// connected to the main chain. Ingesting a block updates the wallet's internal
// utxo state based on the outputs created and destroyed within each block.
func (m *InMemoryWallet) IngestBlock(header []byte, filteredTxns [][]byte) {
	var hdr wire.BlockHeader
	if err := hdr.FromBytes(header); err != nil {
		panic(err)
	}
	height := int64(hdr.Height)

	txns := make([]*dcrutil.Tx, 0, len(filteredTxns))
	for _, txBytes := range filteredTxns {
		tx, err := dcrutil.NewTxFromBytes(txBytes)
		if err != nil {
			panic(err)
		}
		txns = append(txns, tx)
	}

	// Append this new chain update to the end of the queue of new chain
	// updates.
	m.chainMtx.Lock()
	m.chainUpdates = append(m.chainUpdates, &chainUpdate{height, txns})
	m.chainMtx.Unlock()

	// Launch a goroutine to signal the chainSyncer that a new update is
	// available. We do this in a new goroutine in order to avoid blocking
	// the main loop of the rpc client.
	go func() {
		m.chainUpdateSignal <- chainUpdateSignal
	}()
}

// chainSyncer is a goroutine dedicated to processing new blocks in order to
// keep the wallet's utxo state up to date.
//
// NOTE: This MUST be run as a goroutine.
func (m *InMemoryWallet) chainSyncer() {
	var update *chainUpdate

	for s := range m.chainUpdateSignal {
		if s == stopSignal {
			break
		}
		// A new update is available, so pop the new chain update from
		// the front of the update queue.
		m.chainMtx.Lock()
		update = m.chainUpdates[0]
		m.chainUpdates[0] = nil // Set to nil to prevent GC leak.
		m.chainUpdates = m.chainUpdates[1:]
		m.chainMtx.Unlock()

		// Update the latest synced height, then process each filtered
		// transaction in the block creating and destroying utxos within
		// the wallet as a result.
		m.Lock()
		m.currentHeight = update.blockHeight
		undo := &undoEntry{
			utxosDestroyed: make(map[wire.OutPoint]*utxo),
		}
		for _, tx := range update.filteredTxns {
			mtx := tx.MsgTx()
			isCoinbase := blockchain.IsCoinBaseTx(mtx)
			txHash := mtx.TxHash()
			m.evalOutputs(mtx.TxOut, &txHash, isCoinbase, undo)
			m.evalInputs(mtx.TxIn, undo)
		}

		// Finally, record the undo entry for this block so we can
		// properly update our internal state in response to the block
		// being re-org'd from the main chain.
		m.reorgJournal[update.blockHeight] = undo
		m.Unlock()
	}
}

// evalOutputs evaluates each of the passed outputs, creating a new matching
// utxo within the wallet if we're able to spend the output.
func (m *InMemoryWallet) evalOutputs(outputs []*wire.TxOut, txHash *chainhash.Hash, isCoinbase bool, undo *undoEntry) {
	for i, output := range outputs {
		pkScript := output.PkScript

		// Scan all the addresses we currently control to see if the
		// output is paying to us.
		for keyIndex, addr := range m.addrs {
			pkHash := addr.ScriptAddress()
			if !bytes.Contains(pkScript, pkHash) {
				continue
			}

			// If this is a coinbase output, then we mark the
			// maturity height at the proper block height in the
			// future.
			var maturityHeight int64
			if isCoinbase {
				maturityHeight = m.currentHeight + int64(m.net.CoinbaseMaturity)
			}

			op := wire.OutPoint{Hash: *txHash, Index: uint32(i)}
			m.utxos[op] = &utxo{
				value:          dcrutil.Amount(output.Value),
				keyIndex:       keyIndex,
				maturityHeight: maturityHeight,
				pkScript:       pkScript,
			}
			undo.utxosCreated = append(undo.utxosCreated, op)
		}
	}
}

// evalInputs scans all the passed inputs, destroying any utxos within the
// wallet which are spent by an input.
func (m *InMemoryWallet) evalInputs(inputs []*wire.TxIn, undo *undoEntry) {
	for _, txIn := range inputs {
		op := txIn.PreviousOutPoint
		oldUtxo, ok := m.utxos[op]
		if !ok {
			continue
		}

		undo.utxosDestroyed[op] = oldUtxo
		delete(m.utxos, op)
	}
}

// UnwindBlock is a call-back which is to be executed each time a block is
// disconnected from the main chain. Unwinding a block undoes the effect that a
// particular block had on the wallet's internal utxo state.
func (m *InMemoryWallet) UnwindBlock(header []byte) {
	var hdr wire.BlockHeader
	if err := hdr.FromBytes(header); err != nil {
		panic(err)
	}
	height := int64(hdr.Height)

	m.Lock()
	defer m.Unlock()

	undo := m.reorgJournal[height]

	for _, utxo := range undo.utxosCreated {
		delete(m.utxos, utxo)
	}

	for outPoint, utxo := range undo.utxosDestroyed {
		m.utxos[outPoint] = utxo
	}

	delete(m.reorgJournal, height)
}

// newAddress returns a new address from the wallet's hd key chain.  It also
// loads the address into the RPC client's transaction filter to ensure any
// transactions that involve it are delivered via the notifications.
func (m *InMemoryWallet) newAddress() (dcrutil.Address, error) {
	index := m.hdIndex

	childKey, err := m.hdRoot.Child(index)
	if err != nil {
		return nil, err
	}
	privKey, err := childKey.ECPrivKey()
	if err != nil {
		return nil, err
	}
	addr, err := keyToAddr(privKey, m.net)
	if err != nil {
		return nil, err
	}
	m.updateTxFilter()
	m.addrs[index] = addr
	m.hdIndex++
	return addr, nil
}

// NewAddress returns a fresh address spendable by the wallet.
//
// This function is safe for concurrent access.
func (m *InMemoryWallet) NewAddress() (dcrutil.Address, error) {
	m.Lock()
	defer m.Unlock()

	return m.newAddress()
}

// fundTx attempts to fund a transaction sending amt coins.  The coins are
// selected such that the final amount spent pays enough fees as dictated by
// the passed fee rate.  The passed fee rate should be expressed in
// atoms-per-byte.
//
// NOTE: The InMemoryWallet's mutex must be held when this function is called.
func (m *InMemoryWallet) fundTx(tx *wire.MsgTx, amt dcrutil.Amount, feeRate dcrutil.Amount) error {
	const (
		// spendSize is the largest number of bytes of a sigScript
		// which spends a p2pkh output: OP_DATA_73 <sig> OP_DATA_33 <pubkey>
		spendSize = 1 + 73 + 1 + 33
	)

	var (
		amtSelected dcrutil.Amount
		txSize      int
	)

	for outPoint, utxo := range m.utxos {
		// Skip any outputs that are still currently immature or are
		// currently locked.
		if !utxo.isMature(m.currentHeight) || utxo.isLocked {
			continue
		}

		amtSelected += utxo.value

		// Add the selected output to the transaction, updating the
		// current tx size while accounting for the size of the future
		// sigScript.
		tx.AddTxIn(wire.NewTxIn(&outPoint, int64(utxo.value), nil))
		txSize = tx.SerializeSize() + spendSize*len(tx.TxIn)

		// Calculate the fee required for the txn at this point
		// observing the specified fee rate. If we don't have enough
		// coins from he current amount selected to pay the fee, then
		// continue to grab more coins.
		reqFee := dcrutil.Amount(txSize * int(feeRate))
		if amtSelected-reqFee < amt {
			continue
		}

		// If we have any change left over, then add an additional
		// output to the transaction reserved for change.
		changeVal := amtSelected - amt - reqFee
		if changeVal > 0 {
			addr, err := m.newAddress()
			if err != nil {
				return err
			}
			pkScript, err := txscript.PayToAddrScript(addr)
			if err != nil {
				return err
			}
			changeOutput := &wire.TxOut{
				Value:    int64(changeVal),
				PkScript: pkScript,
			}
			tx.AddTxOut(changeOutput)
		}

		return nil
	}

	// If we've reached this point, then coin selection failed due to an
	// insufficient amount of coins.
	return fmt.Errorf("not enough funds for coin selection")
}

// SendOutputs creates, then sends a transaction paying to the specified output
// while observing the passed fee rate. The passed fee rate should be expressed
// in atoms-per-byte.
func (m *InMemoryWallet) SendOutputs(outputs []*wire.TxOut, feeRate dcrutil.Amount) (*chainhash.Hash, error) {
	tx, err := m.CreateTransaction(outputs, feeRate)
	if err != nil {
		return nil, err
	}

	return m.dcrdRpc.SendRawTransaction(tx, true)
}

// CreateTransaction returns a fully signed transaction paying to the specified
// outputs while observing the desired fee rate. The passed fee rate should be
// expressed in atoms-per-byte.
//
// This function is safe for concurrent access.
func (m *InMemoryWallet) CreateTransaction(outputs []*wire.TxOut, feeRate dcrutil.Amount) (*wire.MsgTx, error) {
	m.Lock()
	defer m.Unlock()

	tx := wire.NewMsgTx()

	// Tally up the total amount to be sent in order to perform coin
	// selection shortly below.
	var outputAmt dcrutil.Amount
	for _, output := range outputs {
		outputAmt += dcrutil.Amount(output.Value)
		tx.AddTxOut(output)
	}

	// Attempt to fund the transaction with spendable utxos.
	if err := m.fundTx(tx, outputAmt, feeRate); err != nil {
		return nil, err
	}

	// Populate all the selected inputs with valid sigScript for spending.
	// Along the way record all outputs being spent in order to avoid a
	// potential double spend.
	spentOutputs := make([]*utxo, 0, len(tx.TxIn))
	for i, txIn := range tx.TxIn {
		outPoint := txIn.PreviousOutPoint
		utxo := m.utxos[outPoint]

		extendedKey, err := m.hdRoot.Child(utxo.keyIndex)
		if err != nil {
			return nil, err
		}

		privKey, err := extendedKey.ECPrivKey()
		if err != nil {
			return nil, err
		}

		sigScript, err := txscript.SignatureScript(tx, i, utxo.pkScript,
			txscript.SigHashAll, privKey, true)
		if err != nil {
			return nil, err
		}

		txIn.SignatureScript = sigScript

		spentOutputs = append(spentOutputs, utxo)
	}

	// As these outputs are now being spent by this newly created
	// transaction, mark the outputs are "locked". This action ensures
	// these outputs won't be double spent by any subsequent transactions.
	// These locked outputs can be freed via a call to UnlockOutputs.
	for _, utxo := range spentOutputs {
		utxo.isLocked = true
	}

	return tx, nil
}

// UnlockOutputs unlocks any outputs which were previously locked due to
// being selected to fund a transaction via the CreateTransaction method.
//
// This function is safe for concurrent access.
func (m *InMemoryWallet) UnlockOutputs(inputs []*wire.TxIn) {
	m.Lock()
	defer m.Unlock()

	for _, input := range inputs {
		utxo, ok := m.utxos[input.PreviousOutPoint]
		if !ok {
			continue
		}

		utxo.isLocked = false
	}
}

// ConfirmedBalance returns the confirmed balance of the wallet.
//
// This function is safe for concurrent access.
func (m *InMemoryWallet) ConfirmedBalance() dcrutil.Amount {
	m.RLock()
	defer m.RUnlock()

	var balance dcrutil.Amount
	for _, utxo := range m.utxos {
		// Prevent any immature or locked outputs from contributing to
		// the wallet's total confirmed balance.
		if !utxo.isMature(m.currentHeight) || utxo.isLocked {
			continue
		}

		balance += utxo.value
	}

	return balance
}
