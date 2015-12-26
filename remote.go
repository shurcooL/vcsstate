package vcs

import "html/template"

type Remote interface {
	GetRemote() string // Get primary remote repository url.
	Type() Type        // Returns the type of vcs implementation.

	// TODO: Is this needed?
	GetDefaultBranch() string // Get default branch name for this vcs.

	// TODO: Add error return value?
	GetRemoteRev() string // Get latest remote revision of default branch.
}

type commonRemote struct {
	remote template.URL
}

func (this *commonRemote) GetRemote() string {
	return string(this.remote)
}

func NewRemote(t Type, remote template.URL) Remote {
	switch t {
	case Git:
		return &gitRemote{commonRemote: commonRemote{remote: remote}}
	case Hg:
		// TODO.
		panic("not implemented")
	default:
		panic("bad vcs.Type")
	}
}
