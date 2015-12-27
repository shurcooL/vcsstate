package vcsstate

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/shurcooL/go/osutil"
	"github.com/shurcooL/go/trim"
)

type git struct{}

func (v git) DefaultBranch() string {
	return v.defaultBranch()
}

func (git) Status(dir string) (string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (git) Stash(dir string) (string, error) {
	cmd := exec.Command("git", "stash", "list")
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (git) RemoteURL(dir string) (string, error) {
	/*
		Not specifying "origin" has a problem with rego repo:

		rego $ git-branches -remote
		| Branch                         | Remote        | Behind | Ahead |
		|--------------------------------|---------------|-------:|:------|
		| master                         | origin/master |      0 | 0     |
		| **remove-obsolete-workaround** |               |        |       |
		rego $ gostatus -v
		b #  sourcegraph.com/sqs/rego/...
		rego $ git ls-remote --get-url origin
		https://github.com/sqs/rego
		rego $ git ls-remote --get-url
		https://github.com/shurcooL/rego

		It's likely a rare edge case because the checked out branch *used to* have another remote, but still.

		I forgot what my motivation for trying to remove it was... It helped in some other situation,
		but I can't remember which. :/ So revert this for now until I can recall, then document it!
	*/
	cmd := exec.Command("git", "ls-remote", "--get-url", "origin")
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return trim.LastNewline(string(out)), nil
}

func (git) LocalBranch(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	// Since rev-parse is considered porcelain and may change, need to error-check its output.
	return trim.LastNewline(string(out)), nil
}

// gitRevisionLength is the length of a git revision hash.
const gitRevisionLength = 40

func (v git) LocalRevision(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", v.defaultBranch())
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	if len(out) < gitRevisionLength {
		return "", fmt.Errorf("output length %v is shorter than %v", len(out), gitRevisionLength)
	}
	return string(out[:gitRevisionLength]), nil
}

func (v git) RemoteRevision(dir string) (string, error) {
	// true here is not a boolean value, but a command /bin/true that will make git think it asked for a password,
	// and prevent potential interactive password prompts (opting to return failure exit code instead).
	cmd := exec.Command("git", "-c", "core.askpass=true", "ls-remote", "--heads", "origin", v.defaultBranch())
	cmd.Dir = dir
	env := osutil.Environ(os.Environ())
	env.Set("GIT_SSH_COMMAND", "ssh -o StrictHostKeyChecking=yes") // Default for StrictHostKeyChecking is "ask", which we don't want since v is non-interactive and we prefer to fail than block asking for user input.
	cmd.Env = env

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	if len(out) < gitRevisionLength {
		return "", fmt.Errorf("output length %v is shorter than %v", len(out), gitRevisionLength)
	}
	return string(out[:gitRevisionLength]), nil
}

func (v git) IsContained(dir string, revision string) (bool, error) {
	cmd := exec.Command("git", "branch", "--list", "--contains", revision, v.defaultBranch())
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(out) >= 2 && trim.LastNewline(string(out[2:])) == v.defaultBranch(), nil
}

func (git) defaultBranch() string {
	return "master"
}
