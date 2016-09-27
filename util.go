package vcsstate

import (
	"bytes"
	"os"
	"os/exec"

	"github.com/shurcooL/go/osutil"
)

// dividedOutput runs the command and returns its standard output and standard error.
func dividedOutput(cmd *exec.Cmd) (stdout []byte, stderr []byte, err error) {
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	return outb.Bytes(), errb.Bytes(), err
}

func newCmd(dir string, env map[string]string, name string, cmd ...string) *exec.Cmd {
	c := exec.Command(name, cmd...)
	if len(dir) != 0 {
		c.Dir = dir
	}
	dstEnv := osutil.Environ(os.Environ())
	for k, v := range env {
		dstEnv.Set(k, v)
	}
	dstEnv.Set("LANG", "C.UTF8")
	dstEnv.Set("LANGUAGE", "C.UTF8")
	c.Env = dstEnv
	return c
}

func remoteEnv() map[string]string {
	// THINK: Should we use "-c", "credential.helper=true"?
	//        It's higher priority than GIT_ASKPASS, but
	//        maybe stops private repos from working?
	return map[string]string{
		// `true` here is not a boolean value, but a command /bin/true that will
		// make git think it asked for a password, and prevent potential
		// interactive password prompts (opting to return failure exit code
		// instead).
		"GIT_ASKPASS": "true",
		// Default for StrictHostKeyChecking is "ask", which we don't want since
		// this is non-interactive and we prefer to fail than block asking for user
		// input.
		"GIT_SSH_COMMAND": "ssh -o StrictHostKeyChecking=yes",
	}
}
