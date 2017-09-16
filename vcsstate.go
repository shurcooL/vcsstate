// Package vcsstate allows getting the state of version control system repositories.
package vcsstate

import (
	"bytes"
	"errors"
	"fmt"

	"golang.org/x/tools/go/vcs"
)

// ErrNoRemote is the error used when the local repository doesn't have a valid remote.
var ErrNoRemote = errors.New("local repository has no valid remote")

// NotFoundError records an error where the remote repository is not found.
type NotFoundError struct {
	Err error // Underlying error with more details.
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("remote repository not found:\n%v", e.Err)
}

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

	// Contains reports whether the local default branch contains
	// the commit specified by revision.
	Contains(dir string, revision string, defaultBranch string) (bool, error)

	// RemoteContains reports whether the remote default branch contains
	// the commit specified by revision.
	RemoteContains(dir string, revision string, defaultBranch string) (bool, error)

	// RemoteURL returns primary remote URL, as set in the local repository.
	// If there's no remote, then ErrNoRemote is returned.
	RemoteURL(dir string) (string, error)

	// RemoteBranchAndRevision returns the name and latest revision of the default branch
	// from the remote. If there's no remote, then ErrNoRemote is returned, and the
	// default branch can be queried with NoRemoteDefaultBranch.
	// If the remote repository is not found, NotFoundError is returned,
	// and the default branch can be queried with NoRemoteDefaultBranch.
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
		if gitBinaryError != nil {
			return nil, gitBinaryError
		}
		var major, minor int
		_, err := fmt.Fscanf(bytes.NewReader(gitBinaryVersion), "git version %d.%d", &major, &minor)
		if err != nil {
			return nil, err
		}
		if major > 2 || major == 2 && minor >= 8 {
			return git28{}, nil
		} else if major > 1 || major == 1 && minor >= 7 {
			return git17{}, nil
		} else {
			return nil, fmt.Errorf("git support requires git binary version 1.7+, but you have: %q", gitBinaryVersion)
		}
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
	// from the remote. If the remote repository is not found, NotFoundError is returned.
	RemoteBranchAndRevision(remoteURL string) (branch string, revision string, err error)
}

// NewRemoteVCS creates a RemoteVCS with same type as vcs.
func NewRemoteVCS(vcs *vcs.Cmd) (RemoteVCS, error) {
	switch vcs.Cmd {
	case "git":
		if gitBinaryError != nil {
			return nil, gitBinaryError
		}
		var major, minor int
		_, err := fmt.Fscanf(bytes.NewReader(gitBinaryVersion), "git version %d.%d", &major, &minor)
		if err != nil {
			return nil, err
		}
		if major > 2 || major == 2 && minor >= 8 {
			return remoteGit28{}, nil
		} else if major > 1 || major == 1 && minor >= 7 {
			return remoteGit17{}, nil
		} else {
			return nil, fmt.Errorf("remote git support requires git binary version 1.7+, but you have: %q", gitBinaryVersion)
		}
	case "hg":
		return remoteHg{}, hgBinaryError
	default:
		return nil, fmt.Errorf("%v (%v) support not implemented", vcs.Name, vcs.Cmd)
	}
}
