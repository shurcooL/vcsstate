// Package vcsstate allows getting the state of version control system repositories.
package vcsstate

import (
	"fmt"

	"golang.org/x/tools/go/vcs"
)

// VCS describes how to use a version control system to get the status of a repository
// rooted at dir.
type VCS interface {
	// DefaultBranch returns default branch name for this VCS type.
	DefaultBranch() string

	// Status gets the status of working directory.
	// It returns empty string if no outstanding status.
	Status(dir string) (string, error)

	// LocalBranch gets currently checked out local branch name.
	LocalBranch(dir string) (string, error)

	// LocalRevision gets current local revision of default branch.
	LocalRevision(dir string) (string, error)

	// Stash returns a non-empty string if the repository has a stash.
	Stash(dir string) (string, error)

	// RemoteURL gets primary remote URL.
	RemoteURL(dir string) (string, error)

	// RemoteRevision gets latest remote revision of default branch.
	RemoteRevision(dir string) (string, error)

	// IsContained returns true iff given commit is contained in the local default branch.
	//
	// TODO: Rename IsContained to a better name.
	IsContained(dir string, revision string) (bool, error)
}

// NewVCS creates a repository using VCS type.
func NewVCS(vcs *vcs.Cmd) (VCS, error) {
	switch vcs.Cmd {
	case "git":
		return git{}, nil
	case "hg":
		return hg{}, nil
	default:
		return nil, fmt.Errorf("unsupported vcs.Cmd: %v", vcs.Cmd)
	}
}

// RemoteVCS describes how to use a version control system to get the remote status of a repository
// with remoteURL.
type RemoteVCS interface {
	// RemoteRevision gets latest remote revision of default branch.
	RemoteRevision(remoteURL string) (string, error)
}

// NewRemoteVCS creates a remote repository using VCS type.
func NewRemoteVCS(vcs *vcs.Cmd) (RemoteVCS, error) {
	switch vcs.Cmd {
	case "git":
		return remoteGit{}, nil
	case "hg":
		return nil, fmt.Errorf("RemoteVCS not implemented for hg")
	default:
		return nil, fmt.Errorf("unsupported vcs.Cmd: %v", vcs.Cmd)
	}
}
