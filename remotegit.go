package vcsstate

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/shurcooL/go/osutil"
	"github.com/shurcooL/go/trim"
)

type remoteGit struct{}

func (remoteGit) RemoteBranchAndRevision(remoteURL string) (branch string, revision string, err error) {
	// true here is not a boolean value, but a command /bin/true that will make git think it asked for a password,
	// and prevent potential interactive password prompts (opting to return failure exit code instead).
	cmd := exec.Command("git", "-c", "core.askpass=true", "ls-remote", remoteURL, "HEAD", "refs/heads/*")
	env := osutil.Environ(os.Environ())
	env.Set("GIT_SSH_COMMAND", "ssh -o StrictHostKeyChecking=yes") // Default for StrictHostKeyChecking is "ask", which we don't want since this is non-interactive and we prefer to fail than block asking for user input.
	cmd.Env = env

	stdout, stderr, err := dividedOutput(cmd)
	if err != nil {
		return "", "", fmt.Errorf("%v: %s", err, trim.LastNewline(string(stderr)))
	}
	return parseGitLsRemote(stdout)
}
