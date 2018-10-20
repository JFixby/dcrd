package dcrdtestnode

import (
	"github.com/decred/dcrd/dcrtest"
	"github.com/decred/dcrd/dcrtest/commandline"
	"net"
	"path/filepath"
	"strconv"
	"github.com/decred/dcrd/dcrtest/integrationtest"
)

type DcrdTestServerFactory struct {
	DcrdExecutablePathProvider commandline.ExecutablePathProvider
}

func (factory *DcrdTestServerFactory) NewNode(config *integrationtest.DcrdNodeConfig) integrationtest.DcrdNode {
	exec := factory.DcrdExecutablePathProvider

	dcrtest.AssertNotNil("DcrdExecutablePathProvider", exec)
	dcrtest.AssertNotNil("WorkingDir", config.WorkingDir)
	dcrtest.AssertNotEmpty("WorkingDir", config.WorkingDir)

	dcrd := &DcrdTestServer{
		p2pAddress:                 net.JoinHostPort(config.P2PHost, strconv.Itoa(config.P2PPort)),
		rpcListen:                  net.JoinHostPort(config.DcrdRPCHost, strconv.Itoa(config.DcrdRPCPort)),
		rpcUser:                    "user",
		rpcPass:                    "pass",
		appDir:                     filepath.Join(config.WorkingDir, "dcrd"),
		endpoint:                   "ws",
		externalProcess:            &commandline.ExternalProcess{CommandName: "dcrd"},
		rPCClient:                  &integrationtest.RPCConnection{MaxConnRetries: 20},
		DcrdExecutablePathProvider: exec,
		network:                    config.ActiveNet,
	}
	return dcrd
}
