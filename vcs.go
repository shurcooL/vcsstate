// Package vcsstate allows getting the state of version control system repositories.
package vcsstate

import "golang.org/x/tools/go/vcs"

// VCS describes how to use a version control system to get the status of a repository
// rooted at dir.
//
// TODO: Better/more consistent documentation for the methods.
type VCS interface {
	DefaultBranch() string // Get default branch name for this vcs.

	Status(dir string) (string, error) // Returns empty string if no outstanding status.
	Stash(dir string) (string, error)  // Returns empty string if no stash.

	RemoteURL(dir string) (string, error) // Get primary remote URL.

	LocalBranch(dir string) (string, error) // Get currently checked out local branch name.

	LocalRevision(dir string) (string, error)  // Get current local revision of default branch.
	RemoteRevision(dir string) (string, error) // Get latest remote revision of default branch.

	// Returns true if given commit is contained in the local default branch.
	//
	// TODO: Rename IsContained to a better name.
	IsContained(dir string, revision string) (bool, error)
}

// NewVCS creates a repository using vcs type.
func NewVCS(vcs *vcs.Cmd) VCS {
	switch vcs.Cmd {
	case "git":
		return git{}
	case "hg":
		return hg{}
	default:
		// TODO: No panic.
		panic("unsupported *vcs.Cmd type")
	}
}

// RemoteVCS describes how to use a version control system to get the remote status of a repository
// with remoteURL.
type RemoteVCS interface {
	RemoteRevision(remoteURL string) (string, error) // Get latest remote revision of default branch.
}

// NewRemoteVCS creates a remote repository using vcs type.
func NewRemoteVCS(vcs *vcs.Cmd) RemoteVCS {
	switch vcs.Cmd {
	case "git":
		return remoteGit{}
	case "hg":
		// TODO: No panic.
		panic("not implemented")
	default:
		// TODO: No panic.
		panic("unsupported *vcs.Cmd type")
	}
}
