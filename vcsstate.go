// Package vcsstate allows getting the state of version control system repositories.
package vcsstate

import (
	"errors"
	"fmt"

	"golang.org/x/tools/go/vcs"
)

// ErrNoRemote is returned by RemoteURL and RemoteBranchAndRevision when the local
// repository doesn't have a valid remote.
var ErrNoRemote = errors.New("local repository has no valid remote")

// VCS describes how to use a version control system to get the status of a repository
// rooted at dir.
type VCS interface {
	// Status returns the status of working directory.
	// It returns empty string if no outstanding status.
	Status(dir string) (string, error)

	// Branch returns the name of the locally checked out branch.
	Branch(dir string) (string, error)

	// LocalRevision returns current local revision of default branch.
	LocalRevision(dir string, defaultBranch string) (string, error)

	// Stash returns a non-empty string if the repository has a stash.
	Stash(dir string) (string, error)

	// Contains reports if the local default branch contains the commit specified by revision.
	Contains(dir string, revision string, defaultBranch string) (bool, error)

	// RemoteURL returns primary remote URL, as set in the local repository.
	// If there's no remote, then ErrNoRemote is returned.
	RemoteURL(dir string) (string, error)

	// RemoteBranchAndRevision returns the name and latest revision of the default branch
	// from the remote. If there's no remote, then ErrNoRemote is returned, and the
	// default branch can be queried with NoRemoteDefaultBranch.
	// This operation requires the use of network, and will fail if offline.
	// When offline, CachedRemoteDefaultBranch can be used as a fallback.
	RemoteBranchAndRevision(dir string) (branch string, revision string, err error)

	// CachedRemoteDefaultBranch returns a locally cached remote default branch,
	// if it can do so successfully. It can be used to make a best effort guess
	// of the remote default branch when offline. If it fails, the only viable
	// next best fallback before online again is to use NoRemoteDefaultBranch.
	CachedRemoteDefaultBranch() (string, error)

	// NoRemoteDefaultBranch returns the default value of default branch for this vcs.
	// It can only be relied on when there's no remote, since remote can have a custom
	// value of default branch.
	NoRemoteDefaultBranch() string
}

// NewVCS creates a VCS with same type as vcs.
func NewVCS(vcs *vcs.Cmd) (VCS, error) {
	switch vcs.Cmd {
	case "git":
		return git{}, gitBinaryError
	case "hg":
		return hg{}, hgBinaryError
	default:
		return nil, fmt.Errorf("%v (%v) support not implemented", vcs.Name, vcs.Cmd)
	}
}

// RemoteVCS describes how to use a version control system to get the remote status of a repository
// with remoteURL.
type RemoteVCS interface {
	// RemoteBranchAndRevision returns the name and latest revision of the default branch
	// from the remote.
	RemoteBranchAndRevision(remoteURL string) (branch string, revision string, err error)
}

// NewRemoteVCS creates a RemoteVCS with same type as vcs.
func NewRemoteVCS(vcs *vcs.Cmd) (RemoteVCS, error) {
	switch vcs.Cmd {
	case "git":
		return remoteGit{}, gitBinaryError
	case "hg":
		return remoteHg{}, hgBinaryError
	default:
		return nil, fmt.Errorf("%v (%v) support not implemented", vcs.Name, vcs.Cmd)
	}
}
