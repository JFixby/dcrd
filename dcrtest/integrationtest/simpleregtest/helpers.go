package simpleregtest

import (
	"reflect"
	"time"

	"fmt"
	"github.com/decred/dcrd/dcrjson"
	"github.com/decred/dcrd/rpcclient"
	"strconv"
	"testing"
	"github.com/decred/dcrd/dcrtest/integrationtest"
)

// JoinType is an enum representing a particular type of "node join". A node
// join is a synchronization tool used to wait until a subset of nodes have a
// consistent state with respect to an attribute.
type JoinType uint8

const (
	// Blocks is a JoinType which waits until all nodes share the same
	// block height.
	Blocks JoinType = iota

	// Mempools is a JoinType which blocks until all nodes have identical
	// mempool.
	Mempools
)

// JoinNodes is a synchronization tool used to block until all passed nodes are
// fully synced with respect to an attribute. This function will block for a
// period of time, finally returning once all nodes are synced according to the
// passed JoinType. This function be used to to ensure all active test
// harnesses are at a consistent state before proceeding to an assertion or
// check within rpc tests.
func JoinNodes(nodes []*integrationtest.Harness, joinType JoinType) error {
	switch joinType {
	case Blocks:
		return syncBlocks(nodes)
	case Mempools:
		return syncMempools(nodes)
	}
	return nil
}

// syncMempools blocks until all nodes have identical mempools.
func syncMempools(nodes []*integrationtest.Harness) error {
	poolsMatch := false

	for !poolsMatch {
	retry:
		firstPool, err := nodes[0].DcrdRPCClient().GetRawMempool(dcrjson.GRMAll)
		if err != nil {
			return err
		}

		// If all nodes have an identical mempool with respect to the
		// first node, then we're done. Otherwise, drop back to the top
		// of the loop and retry after a short wait period.
		for _, node := range nodes[1:] {
			nodePool, err := node.DcrdRPCClient().GetRawMempool(dcrjson.GRMAll)
			if err != nil {
				return err
			}

			if !reflect.DeepEqual(firstPool, nodePool) {
				time.Sleep(time.Millisecond * 100)
				goto retry
			}
		}

		poolsMatch = true
	}

	return nil
}

// syncBlocks blocks until all nodes report the same block height.
func syncBlocks(nodes []*integrationtest.Harness) error {
	blocksMatch := false

	for !blocksMatch {
	retry:
		blockHeights := make(map[int64]struct{})

		for _, node := range nodes {
			blockHeight, err := node.DcrdRPCClient().GetBlockCount()
			if err != nil {
				return err
			}

			blockHeights[blockHeight] = struct{}{}
			if len(blockHeights) > 1 {
				time.Sleep(time.Millisecond * 100)
				goto retry
			}
		}

		blocksMatch = true
	}

	return nil
}

// ConnectNode establishes a new peer-to-peer connection between the "from"
// harness and the "to" harness.  The connection made is flagged as persistent,
// therefore in the case of disconnects, "from" will attempt to reestablish a
// connection to the "to" harness.
func ConnectNode(from *integrationtest.Harness, to *integrationtest.Harness) error {
	peerInfo, err := from.DcrdRPCClient().GetPeerInfo()
	if err != nil {
		return err
	}
	numPeers := len(peerInfo)

	targetAddr := to.P2PAddress()
	if err := from.DcrdRPCClient().AddNode(targetAddr, rpcclient.ANAdd); err != nil {
		return err
	}

	// Block until a new connection has been established.
	peerInfo, err = from.DcrdRPCClient().GetPeerInfo()
	if err != nil {
		return err
	}
	for len(peerInfo) <= numPeers {
		peerInfo, err = from.DcrdRPCClient().GetPeerInfo()
		if err != nil {
			return err
		}
	}

	return nil
}

// Create a test chain with the desired number of mature coinbase outputs
func generateTestChain(numToGenerate uint32, node *rpcclient.Client) error {
	fmt.Printf("Generating %v blocks...\n", numToGenerate)
	_, err := node.Generate(numToGenerate)
	if err != nil {
		return err
	}
	fmt.Println("Block generation complete.")
	return nil
}

func assertConnectedTo(t *testing.T, nodeA *integrationtest.Harness, nodeB *integrationtest.Harness) {
	nodeAPeers, err := nodeA.DcrdRPCClient().GetPeerInfo()
	if err != nil {
		t.Fatalf("unable to get nodeA's peer info")
	}

	nodeAddr := nodeB.P2PAddress()
	addrFound := false
	for _, peerInfo := range nodeAPeers {
		if peerInfo.Addr == nodeAddr {
			addrFound = true
			break
		}
	}

	if !addrFound {
		t.Fatal("nodeA not connected to nodeB")
	}
}

// Waits for wallet to sync to the target height
func syncWalletTo(rpcClient *rpcclient.Client, desiredHeight int64) (int64, error) {
	var count int64 = 0
	var err error = nil
	for count != desiredHeight {
		//rpctest.Sleep(100)
		count, err = rpcClient.GetBlockCount()
		if err != nil {
			return -1, err
		}
		fmt.Println("   sync to: " + strconv.FormatInt(count, 10))
	}
	return count, nil
}
