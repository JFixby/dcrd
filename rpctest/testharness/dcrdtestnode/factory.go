package dcrdtestnode

import (
	"github.com/decred/dcrd/rpctest"
	"net"
	"strconv"
	"path/filepath"
	"github.com/decred/dcrd/rpctest/commandline"
	"github.com/decred/dcrd/rpctest/testharness"
)

type DcrdTestServerFactory struct {
	DcrdExecutablePathProvider commandline.ExecutablePathProvider
}

func (factory *DcrdTestServerFactory) NewNode(config *testharness.DcrdNodeConfig) testharness.DcrdNode {
	exec := factory.DcrdExecutablePathProvider

	rpctest.AssertNotNil("DcrdExecutablePathProvider", exec)
	rpctest.AssertNotNil("WorkingDir", config.WorkingDir)
	rpctest.AssertNotEmpty("WorkingDir", config.WorkingDir)

	dcrd := &DcrdTestServer{
		p2pAddress:                 net.JoinHostPort(config.P2PHost, strconv.Itoa(config.P2PPort)),
		rpcListen:                  net.JoinHostPort(config.DcrdRPCHost, strconv.Itoa(config.DcrdRPCPort)),
		rpcUser:                    "user",
		rpcPass:                    "pass",
		appDir:                     filepath.Join(config.WorkingDir, "dcrd"),
		endpoint:                   "ws",
		externalProcess:            &commandline.ExternalProcess{CommandName: "dcrd"},
		rPCClient:                  &testharness.RPCConnection{MaxConnRetries: 20},
		DcrdExecutablePathProvider: exec,
		network:                    config.ActiveNet,
	}
	return dcrd
}
