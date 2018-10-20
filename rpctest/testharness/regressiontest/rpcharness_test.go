// Copyright (c) 2016 The btcsuite developers
// Copyright (c) 2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package regressiontest

import (
	"flag"
	"fmt"
	"github.com/decred/dcrd/chaincfg"
	"github.com/decred/dcrd/dcrtest"
	"github.com/decred/dcrd/dcrtest/commandline"
	"github.com/decred/dcrd/dcrtest/gobuilder"
	"github.com/decred/dcrd/rpctest/testharness"
	"github.com/decred/dcrd/rpctest/testharness/dcrdtestnode"
	"github.com/decred/dcrd/rpctest/testharness/memwallet"
	"github.com/jfixby/pin"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"testing"
)

type dcrtestCase func(t *testing.T)

// Default harness name
const MainHarnessName = "main"

/*
skipTest function will trigger when test name is present in the skipTestsList

To use this function add the following code in your test:

    if skipTest(t) {
		t.Skip("Skipping test")
	}

*/
func skipTest(t *testing.T) bool {
	return dcrtest.ListContainsString(skipTestsList, t.Name())
}

// skipTestsList contains names of the tests mentioned in the testCasesToSkip
var skipTestsList []string

// testCasesToSkip, use it to mark tests for being skipped
var testCasesToSkip = []dcrtestCase{}

// Get function name from module name
var funcInModulePath = regexp.MustCompile(`^.*\.(.*)$`)

// Get the name of a function type
func functionName(tc dcrtestCase) string {
	fncName := runtime.FuncForPC(reflect.ValueOf(tc).Pointer()).Name()
	return funcInModulePath.ReplaceAllString(fncName, "$1")
}

// harnessPool stores and manages harnesses
// multiple harness instances may be run concurrently, to allow for testing
// complex scenarios involving multiple nodes.
var harnessPool *dcrtest.Pool

// harnessWithZeroMOSpawner creates a local test harness
// with only the genesis block.
var harnessWithZeroMOSpawner *ChainWithMatureOutputsSpawner

// ObtainHarness manages access to the Pool for test cases
func ObtainHarness(tag string) *testharness.Harness {
	s := harnessPool.ObtainSpawnableConcurrentSafe(tag)
	return s.(*testharness.Harness)
}

var DcrdFactory testharness.DcrdNodeFactory
var WalletFactory testharness.DcrWalletFactory

var Network = &chaincfg.RegNetParams

// TestMain, is executed by go-test, and is
// responsible for setting up and disposing test harnesses.
func TestMain(m *testing.M) {
	flag.Parse()

	{ // Build list of all ignored tests
		for _, testCase := range testCasesToSkip {
			caseName := functionName(testCase)
			skipTestsList = append(skipTestsList, caseName)
		}
	}

	// Deploy test setup
	setupDcrNodeFactory()
	setupWalletFactory()
	setupHarnessPool()
	// Deploy harness spawner with empty test chain
	harnessWithZeroMOSpawner = &ChainWithMatureOutputsSpawner{
		WorkingDir:        WorkingDir,
		DebugDCRDOutput:   false,
		DebugWalletOutput: false,
		NumMatureOutputs:  0,
		BasePort:          30000, // 30001, 30002, ...
		WalletFactory:     WalletFactory,
		DcrdFactory:       DcrdFactory,
		ActiveNet:         Network,
	}

	// Run tests
	exitCode := m.Run()

	// TearDown all harnesses in test Pool.
	// This includes removing all temporary directories,
	// and shutting down any created processes.
	harnessPool.TearDownAll()
	err := DeleteWorkingDir()

	if err != nil {
		pin.E("DeleteWorkingDir", err)
	}

	verifyCorrectExit()
	os.Exit(exitCode)
}
func setupHarnessPool() {
	// Deploy harness spawner with generated
	// test chain of 25 mature outputs
	harnessWith25MOSpawner := &ChainWithMatureOutputsSpawner{
		WorkingDir:        WorkingDir,
		DebugDCRDOutput:   true,
		DebugWalletOutput: true,
		NumMatureOutputs:  25,
		BasePort:          20000, // 20001, 20002, ...
		WalletFactory:     WalletFactory,
		DcrdFactory:       DcrdFactory,
		ActiveNet:         Network,
	}

	harnessPool = dcrtest.NewPool(harnessWith25MOSpawner)

	if !testing.Short() {
		// Initialize harnesses
		// 18 seconds to init each
		// uncomment to init harness before running test
		// otherwise it will be inited on request
		tagsList := []string{
			MainHarnessName,
		}
		harnessPool.InitTags(tagsList)
	}
}
func setupWalletFactory() {
	WalletFactory = &memwallet.MemWalletFactory{}
}
func setupDcrNodeFactory() {
	dcrdProjectGoPath := gobuilder.DetermineProjectPackagePath("dcrd")
	tempBinDir := filepath.Join(WorkingDir, "bin")
	dcrtest.MakeDirs(tempBinDir)

	cfg := &gobuilder.GoBuiderConfig{
		GoProjectPath:    dcrdProjectGoPath,
		OutputFolderPath: tempBinDir,
		BuidFileName:     "dcrd",
	}
	DcrdGoBuilder := gobuilder.NewGoBuider(cfg)
	DcrdGoBuilder.Build()

	DcrdFactory = &dcrdtestnode.DcrdTestServerFactory{
		DcrdExecutablePathProvider: DcrdGoBuilder,
	}
}

// verifyCorrectExit is an additional safety check to ensure required
// teardown routines were properly performed.
func verifyCorrectExit() {
	if harnessPool.Size() != 0 {
		dcrtest.ReportTestSetupMalfunction(
			fmt.Errorf(
				"incorrect state: " +
					"Pool should be disposed before exit. " +
					"Call Pool.TearDownAll() ",
			))
	}

	commandline.VerifyNoExternalProcessLeftBehind()

	file := WorkingDir
	if dcrtest.FileExists(file) {
		dcrtest.ReportTestSetupMalfunction(
			fmt.Errorf(
				" incorrect state: "+
					"working dir should be deleted before exit. %v ",
				file,
			))
	}
}
