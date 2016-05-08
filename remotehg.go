package vcsstate

import (
	"os/exec"
	"strings"

	"github.com/shurcooL/go/trim"
)

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
