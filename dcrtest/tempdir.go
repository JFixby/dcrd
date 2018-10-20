package dcrtest

import (
	"path/filepath"
	"fmt"
)

// TempDirHandler implements LeakyResource
type TempDirHandler struct {
	target string
}

func (t *TempDirHandler) Dispose() {
	if !t.Exists() {
		ReportTestSetupMalfunction(
			fmt.Errorf("folder does not exists: %v",
				t.target,
			),
		)
	}
	DeleteFile(t.target)
	DeRegisterDisposableAsset(t)
}

func (t *TempDirHandler) MakeDir() {
	if t.Exists() {
		ReportTestSetupMalfunction(
			fmt.Errorf("folder does already exists: %v",
				t.target,
			),
		)
	}
	MakeDirs(t.target)
	RegisterDisposableAsset(t)
}

func (t *TempDirHandler) Exists() bool {
	return FileExists(t.target)
}
func (t *TempDirHandler) Path() string {
	return t.target
}

func NewTempDir(targetParent string, targetName string) *TempDirHandler {
	return &TempDirHandler{
		target: filepath.Join(targetParent, targetName),
	}
}
