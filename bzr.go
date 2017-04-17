package vcsstate

import (
	"fmt"
	"os/exec"

	"github.com/shurcooL/go/trim"
)

type bzr struct{}

func (v bzr) DefaultBranch() string {
	return v.defaultBranch()
}

func (bzr) Status(dir string) (string, error) {
	cmd := exec.Command("bzr", "status")
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (bzr) Branch(dir string) (string, error) {
	cmd := exec.Command("bzr", "branch")
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return trim.LastNewline(string(out)), nil
}

// Bazaar uses UUID instead of SHA-1 hashes and is roughtly 60 characters in length
const bzrRevisionLength = 60

func (v bzr) LocalRevision(dir string) (string, error) {
	// Alternative: bzr version-info.
	cmd := exec.Command("bzr", "version-info")
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	if len(out) < bzrRevisionLength {
		return "", fmt.Errorf("output length %v is shorter than %v", len(out), bzrRevisionLength)
	}
	return string(out[:bzrRevisionLength]), nil
}

func (bzr) Stash(dir string) (string, error) {
	// TODO: Does bzr have stashes? Figure it out, add support, etc.
	return "", fmt.Errorf("Stash is not implemented for bzr")
}

func (bzr) Contains(dir string, revision string) (bool, error) {
	// TODO: Implement this.
	return false, fmt.Errorf("Contains is not implemented for bzr")
}

func (bzr) RemoteURL(dir string) (string, error) {
	// TODO: Implement this.
	return "", fmt.Errorf("RemoteURL is not implemented for bzr")
}

func (v bzr) RemoteRevision(dir string) (string, error) {
	// TODO: Implement this.
	return "", fmt.Errorf("RemoteRevision is not implemented for bzr")
}

func (bzr) defaultBranch() string {
	return "default"
}
