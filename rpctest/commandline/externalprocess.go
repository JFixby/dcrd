// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package commandline

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/decred/dcrwallet/errors"
	"github.com/decred/dcrd/rpctest"
)

type ExternalProcess struct {
	CommandName string
	Arguments   []string

	isRunning bool

	WaitForExit bool

	runningCommand *exec.Cmd
}

func (p *ExternalProcess) FullConsoleCommand() string {
	cmd := p.runningCommand
	args := strings.Join(cmd.Args[1:], " ")
	return cmd.Path + " " + args
}

func (p *ExternalProcess) ClearArguments() {
	p.Arguments = []string{}
}

func (process *ExternalProcess) Launch(debugOutput bool) {
	if process.isRunning {
		rpctest.ReportTestSetupMalfunction(errors.Errorf("Process is already running: %v", process.runningCommand))
	}
	process.isRunning = true

	process.runningCommand = exec.Command(process.CommandName, process.Arguments...)
	cmd := process.runningCommand
	fmt.Println("run command # " + cmd.Path)
	fmt.Println(strings.Join(cmd.Args[0:], "\n    "))
	fmt.Println()
	if debugOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err := cmd.Start()
	rpctest.CheckTestSetupMalfunction(err)

	if process.WaitForExit {
		process.waitForExit()
		return
	}

	ExternalProcesses.add(process)
}

// Stop interrupts the running process, and waits until it exits properly.
// It is important that the process be stopped via Stop(), otherwise,
// it will persist unless explicitly killed.
func (process *ExternalProcess) Stop() error {
	if !process.isRunning {
		rpctest.ReportTestSetupMalfunction(errors.Errorf("Process is not running: %v", process.runningCommand))
	}
	process.isRunning = false

	ExternalProcesses.remove(process)

	return killProcess(process, os.Stdout)
}

func (process *ExternalProcess) waitForExit() {
	err := process.runningCommand.Wait()
	rpctest.CheckTestSetupMalfunction(err)
	process.isRunning = false
	ExternalProcesses.remove(process)
}

func (process *ExternalProcess) IsRunning() bool {
	return process.isRunning
}

// On windows, interrupt is not supported, so a kill signal is used instead.
func killProcess(process *ExternalProcess, logStream *os.File) error {
	cmd := process.runningCommand
	defer cmd.Wait()

	fmt.Fprintln(
		logStream,
		fmt.Sprintf(
			"Killing process: %v",
			process.FullConsoleCommand(),
		))

	osProcess := cmd.Process
	if runtime.GOOS == "windows" {
		err := osProcess.Signal(os.Kill)
		return err
	} else {
		err := osProcess.Signal(os.Interrupt)
		return err
	}
}
