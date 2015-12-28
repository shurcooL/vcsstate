package vcsstate

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/shurcooL/go/trim"
)

type hg struct{}

func (v hg) DefaultBranch() string {
	return v.defaultBranch()
}

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

func (v hg) LocalRevision(dir string) (string, error) {
	cmd := exec.Command("hg", "--debug", "identify", "-i", "--rev", v.defaultBranch())
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
	// TODO: Does Mercurial have stashes? Figure it out, add support, etc.
	return "", fmt.Errorf("Stash is not implemented for hg")
}

func (hg) Contains(dir string, revision string) (bool, error) {
	// TODO: Implement this.
	return false, fmt.Errorf("Contains is not implemented for hg")
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

func (v hg) RemoteRevision(dir string) (string, error) {
	cmd := exec.Command("hg", "--debug", "identify", "-i", "--rev", v.defaultBranch(), "default")
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	// Get the last line of output.
	lines := strings.Split(trim.LastNewline(string(out)), "\n") // lines will always contain at least one element.
	return lines[len(lines)-1], nil
}

func (hg) defaultBranch() string {
	return "default"
}
