package vcsstate

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/shurcooL/go/osutil"
	"github.com/shurcooL/go/trim"
)

type git struct{}

func (git) Status(dir string) (string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (git) Branch(dir string) (string, error) {
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

func (git) LocalRevision(dir string, defaultBranch string) (string, error) {
	cmd := exec.Command("git", "rev-parse", defaultBranch)
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

func (git) Stash(dir string) (string, error) {
	cmd := exec.Command("git", "stash", "list")
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (git) Contains(dir string, revision string, defaultBranch string) (bool, error) {
	cmd := exec.Command("git", "branch", "--list", "--contains", revision, defaultBranch)
	cmd.Dir = dir

	stdout, stderr, err := dividedOutput(cmd)
	switch {
	case err == nil:
		// If this commit is contained, the expected output is exactly "* master\n" or "  master\n" if we're on another branch.
		return string(stdout) == fmt.Sprintf("* %s\n", defaultBranch) ||
			string(stdout) == fmt.Sprintf("  %s\n", defaultBranch), nil
	case err != nil && strings.HasPrefix(string(stderr), fmt.Sprintf("error: no such commit %s\n", revision)):
		return false, nil // No such commit error means this commit is not contained.
	default:
		return false, err
	}
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

		Okay, it might've been for when master branch is tracking some non-origin remote.

		Also, not specifying "origin" allows to more easily determine that there's no remote, because exit status is non-zero.

		Ok, it's really needed in the following situation. Imagine you're currently on non-default branch, and that branch
		happens to have a different remote set. Then `git ls-remote --get-url` will get the remote of current branch, instead
		of that of origin. So, in order to get any kind of sane results, we must assume default remote is "origin" and explicitly
		specify it here.
	*/
	cmd := exec.Command("git", "ls-remote", "--get-url", "origin")
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return trim.LastNewline(string(out)), nil
}

func (git) RemoteBranchAndRevision(dir string) (branch string, revision string, err error) {
	// true here is not a boolean value, but a command /bin/true that will make git think it asked for a password,
	// and prevent potential interactive password prompts (opting to return failure exit code instead).
	cmd := exec.Command("git", "-c", "core.askpass=true", "ls-remote", "origin", "HEAD", "refs/heads/*")
	cmd.Dir = dir
	env := osutil.Environ(os.Environ())
	env.Set("GIT_SSH_COMMAND", "ssh -o StrictHostKeyChecking=yes") // Default for StrictHostKeyChecking is "ask", which we don't want since this is non-interactive and we prefer to fail than block asking for user input.
	cmd.Env = env

	stdout, stderr, err := dividedOutput(cmd)
	switch {
	case err != nil && strings.HasPrefix(string(stderr), "fatal: 'origin' does not appear to be a git repository\n"):
		return "", "", ErrNoRemote
	case err != nil:
		return "", "", fmt.Errorf("%v: %s", err, trim.LastNewline(string(stderr)))
	default:
		return parseGitLsRemote(stdout)
	}
}

func (git) NoRemoteDefaultBranch() string {
	return "master"
}

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

func parseGitLsRemote(out []byte) (branch string, revision string, err error) {
	if len(out) == 0 {
		return "", "", errors.New("empty ls-remote output")
	}
	lines := strings.Split(string(out[:len(out)-1]), "\n")
	for _, line := range lines {
		// E.g., "7cafcd837844e784b526369c9bce262804aebc60	refs/heads/main".
		revisionReference := strings.Split(line, "\t")
		rev, ref := revisionReference[0], revisionReference[1]

		// This assumes HEAD comes first, before all other references.
		if ref == "HEAD" {
			revision = rev
			continue
		}

		// HACK: There may be more than one branch that matches; prefer "master" over all
		//       others, but otherwise no choice but to pick a random one, since there does
		//       not seem to be a way of finding it exactly (I'm happy to be proven wrong though).
		// TODO: Once git 2.8 becomes available, use ls-remote --symref to fix this.
		if rev == revision && branch != "master" {
			branch = ref[len("refs/heads/"):]
		}
	}
	if branch == "" || revision == "" {
		return "", "", errors.New("HEAD revision not found in ls-remote output")
	}
	return branch, revision, nil
}
