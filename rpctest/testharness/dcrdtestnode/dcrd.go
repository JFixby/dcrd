// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package dcrdtestnode

import (
	"fmt"
	"path/filepath"
	"io/ioutil"

	"github.com/decred/dcrwallet/errors"
	"github.com/decred/dcrd/rpcclient"
	"github.com/decred/dcrd/rpctest"
	"github.com/decred/dcrd/rpctest/commandline"
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/chaincfg"
	"github.com/decred/dcrd/rpctest/testharness"
)

type DcrdTestServer struct {
	rpcUser    string
	rpcPass    string
	p2pAddress string
	rpcListen  string
	rpcConnect string
	profile    string
	debugLevel string
	appDir     string
	endpoint   string

	externalProcess *commandline.ExternalProcess

	DcrdExecutablePathProvider commandline.ExecutablePathProvider

	rPCClient *testharness.RPCConnection

	miningAddress dcrutil.Address

	network *chaincfg.Params
}

// RPCConnectionConfig creates new connection config for RPC client
func (n *DcrdTestServer) RPCConnectionConfig() *rpcclient.ConnConfig {
	file := n.CertFile()
	fmt.Println("reading: " + file)
	cert, err := ioutil.ReadFile(file)
	rpctest.CheckTestSetupMalfunction(err)

	return &rpcclient.ConnConfig{
		Host:                 n.rpcListen,
		Endpoint:             n.endpoint,
		User:                 n.rpcUser,
		Pass:                 n.rpcPass,
		Certificates:         cert,
		DisableAutoReconnect: true,
		HTTPPostMode:         false,
	}
}

func (server *DcrdTestServer) CertFile() string {
	return filepath.Join(server.appDir, "rpc.cert")
}

func (server *DcrdTestServer) KeyFile() string {
	return filepath.Join(server.appDir, "rpc.key")
}

func (server *DcrdTestServer) Network() *chaincfg.Params {
	return server.network
}

func (h *DcrdTestServer) Dispose() error {
	if h.rPCClient.IsConnected() {
		h.rPCClient.Disconnect()
	}
	if h.IsRunning() {
		h.Stop()
	}
	return nil
}

func (server *DcrdTestServer) IsRunning() bool {
	return server.externalProcess.IsRunning()
}

// Stop interrupts the running dcrd process.
func (n *DcrdTestServer) Stop() {
	if !n.IsRunning() {
		rpctest.ReportTestSetupMalfunction(errors.Errorf("DcrdTestServer is not running"))
	}
	fmt.Println("Stop DCRD process...")
	err := n.externalProcess.Stop()
	rpctest.CheckTestSetupMalfunction(err)
}

func (n *DcrdTestServer) cookArguments(extraArguments map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	result["txindex"] = commandline.NoArgumentValue
	result["addrindex"] = commandline.NoArgumentValue
	result["rpcuser"] = n.rpcUser
	result["rpcpass"] = n.rpcPass
	result["rpcconnect"] = n.rpcConnect
	result["rpclisten"] = n.rpcListen
	result["listen"] = n.p2pAddress
	result["appdata"] = n.appDir
	result["debuglevel"] = n.debugLevel
	result["profile"] = n.profile
	result["rpccert"] = n.CertFile()
	result["rpckey"] = n.KeyFile()
	if n.miningAddress != nil {
		result["miningaddr"] = n.miningAddress.String()
	}
	result[networkFor(n.network)] = commandline.NoArgumentValue

	commandline.ArgumentsCopyTo(extraArguments, result)
	return result
}

// networkFor resolves network argument for dcrd and wallet console commands
func networkFor(net *chaincfg.Params) string {
	if net == &chaincfg.SimNetParams {
		return "simnet"
	}
	if net == &chaincfg.TestNet3Params {
		return "testnet"
	}
	if net == &chaincfg.RegNetParams {
		return "regnet"
	}
	if net == &chaincfg.MainNetParams {
		// no argument needed for the MainNet
		return commandline.NoArgument
	}

	// should never reach this line, report violation
	rpctest.ReportTestSetupMalfunction(fmt.Errorf("unknown network: %v ", net))
	return ""
}

func (server *DcrdTestServer) FullConsoleCommand() string {
	return server.externalProcess.FullConsoleCommand()
}

func (server *DcrdTestServer) P2PAddress() string {
	return server.p2pAddress
}

func (server *DcrdTestServer) RPCClient() *testharness.RPCConnection {
	return server.rPCClient
}

func (server *DcrdTestServer) Shutdown() {
	fmt.Println("Disconnect from DCRD RPC...")
	server.rPCClient.Disconnect()

	server.Stop()

	// Delete files, RPC servers will recreate them on the next launch sequence
	rpctest.DeleteFile(server.CertFile())
	rpctest.DeleteFile(server.KeyFile())
}

func (n *DcrdTestServer) Launch(args *testharness.DcrdLaunchArgs) {
	if n.IsRunning() {
		rpctest.ReportTestSetupMalfunction(errors.Errorf("DcrdTestServer is already running"))
	}
	fmt.Println("Start DCRD process...")
	rpctest.MakeDirs(n.appDir)

	n.miningAddress = args.MiningAddress

	dcrdExe := n.DcrdExecutablePathProvider.Executable()
	n.externalProcess.CommandName = dcrdExe
	n.externalProcess.Arguments = commandline.ArgumentsToStringArray(
		n.cookArguments(args.ExtraArguments),
	)
	n.externalProcess.Launch(args.DebugOutput)
	// DCRD RPC instance will create a cert file when it is ready for incoming calls
	rpctest.WaitForFile(n.CertFile(), 7)

	fmt.Println("Connect to DCRD RPC...")
	cfg := n.RPCConnectionConfig()
	n.rPCClient.Connect(cfg, nil)
	fmt.Println("DCRD RPC client connected.")
}
