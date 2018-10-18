// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package gobuilder

import (
	"os"
	"github.com/decred/dcrd/rpctest"
	"go/build"
	"sync"
	"path/filepath"
	"runtime"
	"github.com/decred/dcrd/rpctest/commandline"
)

type GoBuiderConfig struct {
	GoProjectPath    string
	OutputFolderPath string
	BuidFileName     string
}

func NewGoBuider(config *GoBuiderConfig) *GoBuider {
	buider := &GoBuider{
		cfg: config,
	}
	return buider
}

type GoBuider struct {
	cfg        *GoBuiderConfig
	compileMtx sync.Mutex
}

func (builder *GoBuider) Executable() string {
	outputPath := filepath.Join(builder.cfg.OutputFolderPath, builder.cfg.BuidFileName)
	if runtime.GOOS == "windows" {
		outputPath += ".exe"
	}
	return outputPath
}

func (buider *GoBuider) Build() {
	buider.compileMtx.Lock()
	defer buider.compileMtx.Unlock()

	goProjectPath := buider.cfg.GoProjectPath
	outputFolderPath := buider.cfg.OutputFolderPath
	rpctest.MakeDirs(outputFolderPath)
	clearOutput(buider)

	// check project path
	pkg, err := build.ImportDir(goProjectPath, build.FindOnly)
	rpctest.CheckTestSetupMalfunction(err)
	goProjectPath = pkg.ImportPath

	runBuildCommand(buider, goProjectPath, outputFolderPath)
}

func runBuildCommand(builder *GoBuider, goProjectPath string, outputFolderPath string) {
	// Build and output an executable in a static temp path.
	proc := &commandline.ExternalProcess{
		CommandName: "go",
		WaitForExit: true,
	}
	proc.Arguments = append(proc.Arguments, "build")
	proc.Arguments = append(proc.Arguments, "-v")
	//proc.Arguments = append(proc.Arguments, "-x")
	proc.Arguments = append(proc.Arguments, "-o")
	proc.Arguments = append(proc.Arguments, builder.Executable())
	proc.Arguments = append(proc.Arguments, goProjectPath)

	proc.Launch(true)
}

func clearOutput(buider *GoBuider) {
	targetFolder := buider.cfg.OutputFolderPath
	rpctest.FileExists(targetFolder)
	rpctest.CheckTestSetupMalfunction(os.Remove(targetFolder))
}
