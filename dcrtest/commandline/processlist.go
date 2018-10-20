package commandline

import (
	"os"
	"fmt"
	"github.com/decred/dcrd/rpctest"
)

func init() {
	rpctest.RegisterDisposableAsset(ExternalProcesses)
}

// ExternalProcesses keeps track of all running processes
// to execute emergency killProcess in case of the test setup malfunction
var ExternalProcesses = &ExternalProcessesList{
	set: make(map[*ExternalProcess]bool),
}

type ExternalProcessesList struct {
	set map[*ExternalProcess]bool
}

func (l *ExternalProcessesList) Dispose() {
	l.emergencyKillAll()
}

// VerifyNoExternalProcessLeftBehind sould be called to check if all external
// processes were properly disposed. Will crash if not.
func VerifyNoExternalProcessLeftBehind() {
	N := len(ExternalProcesses.set)
	if N > 0 {
		for k := range ExternalProcesses.set {
			fmt.Fprintln(
				os.Stderr,
				fmt.Sprintf(
					"External process leak, running command: %s",
					k.FullConsoleCommand(),
				))
		}
		rpctest.ReportTestSetupMalfunction(
			fmt.Errorf(
				"incorrect state: %v external processes left running ",
				N,
			))
	}
}

// emergencyKillAll is used to terminate all the external processes
// created within this test setup in case of panic.
// Otherwise, they all will persist unless explicitly killed.
// Should be used only in case of test setup malfunction.
func (list *ExternalProcessesList) emergencyKillAll() {
	for k := range list.set {
		err := killProcess(k, os.Stderr)
		if err != nil {
			fmt.Fprintln(
				os.Stderr,
				fmt.Sprintf(
					"Failed to kill process %v",
					err,
				))
		}
	}
}

func (list *ExternalProcessesList) add(process *ExternalProcess) {
	list.set[process] = true
}

func (list *ExternalProcessesList) remove(process *ExternalProcess) {
	delete(list.set, process)
}
