package vcsstate

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseGitLsRemote(t *testing.T) {
	tests := []struct {
		in           []byte
		wantBranch   string
		wantRevision string
		wantErr      error
	}{
		{
			in: []byte(`7cafcd837844e784b526369c9bce262804aebc60	HEAD
0a50dc0e5a012dbf22f1289471dc52bc0fe44e9a	refs/heads/cb
7cafcd837844e784b526369c9bce262804aebc60	refs/heads/main
67253a98dcf0dd3273ea58929d7aaaa781ef6d13	refs/heads/wip
`),
			wantBranch:   "main",
			wantRevision: "7cafcd837844e784b526369c9bce262804aebc60",
		},
		{
			in:      []byte(""),
			wantErr: errors.New("empty ls-remote output"),
		},
		// Prefer "master" even though "datetimes" has same revision.
		{
			in: []byte(`f0aeabca5a127c4078abb8c8d64298b147264b55	HEAD
f0aeabca5a127c4078abb8c8d64298b147264b55	refs/heads/datetimes
f42bdee2ab503fed466739d8e8c55ae34fd9be45	refs/heads/encode-tagged-anon-structs
f0aeabca5a127c4078abb8c8d64298b147264b55	refs/heads/master
f93697607c2406ba00b57fe418f4101b4e447eb8	refs/heads/numbers
`),
			wantBranch:   "master",
			wantRevision: "f0aeabca5a127c4078abb8c8d64298b147264b55",
		},
	}

	for _, test := range tests {
		branch, revision, err := parseGitLsRemote(test.in)
		if got, want := err, test.wantErr; !reflect.DeepEqual(got, want) {
			t.Errorf("got %q, want %q", got, want)
		}
		if test.wantErr != nil {
			continue
		}

		if got, want := branch, test.wantBranch; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
		if got, want := revision, test.wantRevision; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}
