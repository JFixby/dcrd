// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package integrationtest

import (
	"fmt"
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/rpcclient"
	"os"
)

// Harness provides a unified platform for creating RPC-driven integration
// tests involving dcrd and dcrwallet.
// The active DcrdTestServer and WalletServer will typically be run in simnet
// mode to allow for easy generation of test blockchains.
type Harness struct {
	name string

	DcrdServer DcrdNode
	Wallet     DcrWallet

	WorkingDir string

	MiningAddress dcrutil.Address
}

// DcrdRPCClient manages access to the RPCClient,
// test cases suppose to use it when the need access to the Dcrd RPC
func (h *Harness) DcrdRPCClient() *rpcclient.Client {
	return h.DcrdServer.RPCClient().rpcClient
}

func (harness *Harness) DeleteWorkingDir() error {
	dir := harness.WorkingDir
	fmt.Println("delete: " + dir)
	err := os.RemoveAll(dir)
	return err
}

func (h *Harness) P2PAddress() string {
	return h.DcrdServer.P2PAddress()
}
func (harness *Harness) Name() string {
	return harness.name
}
