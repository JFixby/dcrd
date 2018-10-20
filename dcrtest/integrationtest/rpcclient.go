// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package integrationtest

import (
	"fmt"
	"math"
	"time"

	"github.com/decred/dcrd/dcrtest"
	"github.com/decred/dcrd/rpcclient"
	"github.com/decred/dcrwallet/errors"
)

type RPCConnection struct {
	rpcClient      *rpcclient.Client
	MaxConnRetries int
	isConnected    bool
}

func (client *RPCConnection) Connect(rpcConf *rpcclient.ConnConfig, ntfnHandlers *rpcclient.NotificationHandlers) {
	if client.isConnected {
		dcrtest.ReportTestSetupMalfunction(errors.Errorf("%v is already connected", client.rpcClient))
	}
	client.isConnected = true
	rpcClient := NewRPCConnection(rpcConf, client.MaxConnRetries, ntfnHandlers)
	err := rpcClient.NotifyBlocks()
	dcrtest.CheckTestSetupMalfunction(err)
	client.rpcClient = rpcClient
}

func (client *RPCConnection) Disconnect() {
	if !client.isConnected {
		dcrtest.ReportTestSetupMalfunction(errors.Errorf("%v is already disconnected", client))
	}
	client.isConnected = false
	client.rpcClient.Disconnect()
	client.rpcClient.Shutdown()
}

func NewRPCConnection(config *rpcclient.ConnConfig, maxConnRetries int, ntfnHandlers *rpcclient.NotificationHandlers) *rpcclient.Client {
	var client *rpcclient.Client
	var err error = nil

	for i := 0; i < maxConnRetries; i++ {
		client, err = rpcclient.New(config, ntfnHandlers)
		if err != nil {
			fmt.Println("err: " + err.Error())
			time.Sleep(time.Duration(math.Log(float64(i+3))) * 50 * time.Millisecond)
			continue
		}
		break
	}
	if client == nil {
		dcrtest.ReportTestSetupMalfunction(errors.Errorf("client connection timedout"))
	}
	return client
}

func (client *RPCConnection) IsConnected() bool {
	return client.isConnected
}
func (client *RPCConnection) Connection() *rpcclient.Client {
	return client.rpcClient
}
