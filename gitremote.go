package vcs

import (
	"os"
	"os/exec"

	"github.com/shurcooL/go/osutil"
)

type gitRemote struct {
	commonRemote
}

func (gr *gitRemote) Type() Type { return Git }

func (gr *gitRemote) GetDefaultBranch() string {
	return "master"
}

func (gr *gitRemote) GetRemoteRev() string {
	// true here is not a boolean value, but a command /bin/true that will make git think it asked for a password,
	// and prevent potential interactive password prompts (opting to return failure exit code instead).
	cmd := exec.Command("git", "-c", "core.askpass=true", "ls-remote", "--heads", string(gr.remote), gr.GetDefaultBranch())
	env := osutil.Environ(os.Environ())
	env.Set("GIT_SSH_COMMAND", "ssh -o StrictHostKeyChecking=yes") // Default for StrictHostKeyChecking is "ask", which we don't want since this is non-interactive and we prefer to fail than block asking for user input.
	cmd.Env = env

	if out, err := cmd.Output(); err == nil && len(out) >= gitRevisionLength {
		return string(out[:gitRevisionLength])
	} else {
		return ""
	}
}
