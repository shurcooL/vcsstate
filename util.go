package vcsstate

import (
	"bytes"
	"os"
	"os/exec"
	"time"
)

// timeout for running commands. It helps if some remote server stalls and doesn't hang up, etc.
const timeout = 20 * time.Second

// outputTimeout runs the command and returns its standard output,
// with a timeout.
func outputTimeout(cmd *exec.Cmd) ([]byte, error) {
	var buf bytes.Buffer
	cmd.Stdout = &buf
	err := runTimeout(cmd)
	return buf.Bytes(), err
}

// dividedOutputTimeout runs the command and returns its standard output and standard error,
// with a timeout.
func dividedOutputTimeout(cmd *exec.Cmd) (stdout []byte, stderr []byte, err error) {
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = runTimeout(cmd)
	return outBuf.Bytes(), errBuf.Bytes(), err
}

// runTimeout starts the specified command and waits for it to complete,
// up to a timeout.
func runTimeout(cmd *exec.Cmd) error {
	err := cmd.Start()
	if err != nil {
		return err
	}
	t := time.AfterFunc(timeout, func() {
		cmd.Process.Signal(os.Interrupt)
	})
	defer t.Stop()
	return cmd.Wait()
}
