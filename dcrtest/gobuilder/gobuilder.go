// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package gobuilder

import (
	"path/filepath"
	"runtime"
	"sync"

	"github.com/decred/dcrd/dcrtest"
	"github.com/decred/dcrd/dcrtest/commandline"
	"go/build"
)

// GoBuider is a handler helping to build a target Go project
type GoBuider struct {
	GoProjectPath    string
	OutputFolderPath string
	BuildFileName    string

	compileMtx sync.Mutex
}

// Executable returns full path to an executable target file
func (builder *GoBuider) Executable() string {
	outputPath := filepath.Join(
		builder.OutputFolderPath, builder.BuildFileName)
	if runtime.GOOS == "windows" {
		outputPath += ".exe"
	}
	return outputPath
}

// Build compiles target project and writes output to the target output folder
func (builder *GoBuider) Build() {
	builder.compileMtx.Lock()
	defer builder.compileMtx.Unlock()

	goProjectPath := builder.GoProjectPath
	outputFolderPath := builder.OutputFolderPath
	dcrtest.MakeDirs(outputFolderPath)

	target := builder.Executable()
	if dcrtest.FileExists(target) {
		deleteOutputExecutable(builder)
		dcrtest.DeRegisterDisposableAsset(builder)
	}

	// check project path
	pkg, err := build.ImportDir(goProjectPath, build.FindOnly)
	dcrtest.CheckTestSetupMalfunction(err)
	goProjectPath = pkg.ImportPath

	runBuildCommand(builder, goProjectPath)
	dcrtest.RegisterDisposableAsset(builder)
}

// runBuildCommand calls `go build`
func runBuildCommand(builder *GoBuider, goProjectPath string) {
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

func (builder *GoBuider) Dispose() {
	deleteOutputExecutable(builder)
	dcrtest.DeRegisterDisposableAsset(builder)
}

func deleteOutputExecutable(builder *GoBuider) {
	dcrtest.DeleteFile(builder.Executable())
}
