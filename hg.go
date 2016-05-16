package vcsstate

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/shurcooL/go/trim"
)

type hg struct{}

func (hg) Status(dir string) (string, error) {
	cmd := exec.Command("hg", "status")
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (hg) Branch(dir string) (string, error) {
	cmd := exec.Command("hg", "branch")
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return trim.LastNewline(string(out)), nil
}

// hgRevisionLength is the length of a Mercurial revision hash.
const hgRevisionLength = 40

func (hg) LocalRevision(dir string, defaultBranch string) (string, error) {
	cmd := exec.Command("hg", "--debug", "identify", "-i", "--rev", defaultBranch)
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	if len(out) < hgRevisionLength {
		return "", fmt.Errorf("output length %v is shorter than %v", len(out), hgRevisionLength)
	}
	return string(out[:hgRevisionLength]), nil
}

func (hg) Stash(dir string) (string, error) {
	cmd := exec.Command("hg", "shelve", "--list")
	cmd.Dir = dir

	stdout, stderr, err := dividedOutput(cmd)
	switch {
	case err == nil && len(stdout) != 0:
		return string(stdout), nil
	case err == nil && len(stdout) == 0:
		return "", nil
	case err != nil && string(stderr) == "hg: unknown command 'shelve'\n":
		return "", nil
	default:
		return "", err
	}
}

func (hg) Contains(dir string, revision string, defaultBranch string) (bool, error) {
	cmd := exec.Command("hg", "log", "--branch", defaultBranch, "--rev", revision)
	cmd.Dir = dir

	stdout, stderr, err := dividedOutput(cmd)
	switch {
	case err == nil && len(stdout) != 0:
		return true, nil // Non-zero output means this commit is indeed contained.
	case err == nil && len(stdout) == 0:
		return false, nil // Zero output means this commit is not contained.
	case err != nil && string(stderr) == fmt.Sprintf("abort: unknown revision '%s'!\n", revision):
		return false, nil // Unknown revision error means this commit is not contained.
	default:
		return false, err
	}
}

func (hg) RemoteURL(dir string) (string, error) {
	cmd := exec.Command("hg", "paths", "default")
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return trim.LastNewline(string(out)), nil
}

func (hg) RemoteBranchAndRevision(dir string) (branch string, revision string, err error) {
	// TODO: Query remote branch from actual remote; it's currently hardcoded to "default".
	const defaultBranch = "default"

	cmd := exec.Command("hg", "--debug", "identify", "-i", "--rev", defaultBranch, "default")
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return "", "", err
	}
	// Get the last line of output.
	lines := strings.Split(trim.LastNewline(string(out)), "\n") // lines will always contain at least one element.
	return defaultBranch, lines[len(lines)-1], nil
}

func (hg) NoRemoteDefaultBranch() string {
	return "default"
}

type remoteHg struct{}

func (remoteHg) RemoteBranchAndRevision(remoteURL string) (branch string, revision string, err error) {
	// TODO: Query remote branch from actual remote; it's currently hardcoded to "default".
	const defaultBranch = "default"

	cmd := exec.Command("hg", "--debug", "identify", "-i", "--rev", defaultBranch, remoteURL)

	out, err := cmd.Output()
	if err != nil {
		return "", "", err
	}
	// Get the last line of output.
	lines := strings.Split(trim.LastNewline(string(out)), "\n") // lines will always contain at least one element.
	return defaultBranch, lines[len(lines)-1], nil
}
