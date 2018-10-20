package integrationtest

import (
	"github.com/decred/dcrd/chaincfg"
	"github.com/decred/dcrd/dcrutil"
	"github.com/decred/dcrd/rpcclient"
)

type DcrdNode interface {
	Dispose() error
	Shutdown()
	Launch(args *DcrdLaunchArgs)
	CertFile() string
	RPCConnectionConfig() *rpcclient.ConnConfig
	RPCClient() *RPCConnection
	P2PAddress() string
	Network() *chaincfg.Params
}

type DcrdNodeConfig struct {
	ActiveNet *chaincfg.Params

	WorkingDir string

	P2PHost string
	P2PPort int

	DcrdRPCHost string
	DcrdRPCPort int
}

type DcrdNodeFactory interface {
	NewNode(cfg *DcrdNodeConfig) DcrdNode
}

type DcrdLaunchArgs struct {
	ExtraArguments map[string]interface{}
	DebugOutput    bool
	MiningAddress  dcrutil.Address
}
