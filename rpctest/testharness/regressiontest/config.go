package regressiontest

import (
	"fmt"
	"io/ioutil"
	"os"
)

// Test setup working directory
var WorkingDir = SetupWorkingDir()

func SetupWorkingDir() string {
	testWorkingDir, err := ioutil.TempDir("", "testharness")
	if err != nil {
		fmt.Println("Unable to create working dir: ", err)
		os.Exit(-1)
	}
	return testWorkingDir
}

func DeleteWorkingDir() error {
	file := WorkingDir
	y, err := fileExists(file)
	if err != nil {
		return err
	}
	if y {
		fmt.Println("delete: " + file)
		return os.RemoveAll(file)
	}
	return nil
}

func fileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
