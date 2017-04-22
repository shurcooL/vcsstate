package vcsstate

import (
	"errors"
	"reflect"
	"testing"
)

func TestGuessBranch(t *testing.T) {
	tests := []struct {
		in         []byte
		revision   string
		wantBranch string
		wantErr    error
	}{
		{
			// Issue #10.
			// git ls-remote --symref https://code.googlesource.com/google-api-go-client HEAD 'refs/heads/*'
			in: []byte(`fbbaff1827317122a8a0e1b24de25df8417ce87b	HEAD
fbbaff1827317122a8a0e1b24de25df8417ce87b	refs/heads/master
`),
			revision:   "fbbaff1827317122a8a0e1b24de25df8417ce87b",
			wantBranch: "master",
		},
	}

	for _, test := range tests {
		branch, err := guessBranch(test.in, test.revision)
		if got, want := err, test.wantErr; !reflect.DeepEqual(got, want) {
			t.Errorf("got %#v, want %#v", got, want)
		}
		if test.wantErr != nil {
			continue
		}

		if got, want := branch, test.wantBranch; got != want {
			t.Errorf("got branch %q, want %q", got, want)
		}
	}
}

func TestParseGit17Remote(t *testing.T) {
	tests := []struct {
		in      []byte
		want    string
		wantErr error
	}{
		{
			in: []byte(`foo	https://github.com/foo/vcsstate (fetch)
foo	https://github.com/foo/vcsstate (push)
origin	https://github.com/shurcooL/vcsstate (fetch)
origin	https://github.com/shurcooL/vcsstate (push)
somebody	https://github.com/somebody/vcsstate (fetch)
somebody	https://github.com/somebody/vcsstate (push)
`),
			want: "https://github.com/shurcooL/vcsstate",
		},
		{
			in:      []byte(""),
			wantErr: errors.New("no origin remote"),
		},
		// Only accept "origin" remote, even if others exist.
		{
			in: []byte(`fork	https://github.com/foobar/vcsstate (fetch)
fork	https://github.com/foobar/vcsstate (push)
`),
			wantErr: errors.New("no origin remote"),
		},
	}

	for _, test := range tests {
		url, err := parseGit17Remote(test.in)
		if got, want := err, test.wantErr; !reflect.DeepEqual(got, want) {
			t.Errorf("got %#v, want %#v", got, want)
		}
		if test.wantErr != nil {
			continue
		}

		if got, want := url, test.want; got != want {
			t.Errorf("got url %q, want %q", got, want)
		}
	}
}

func TestParseGit17LsRemote(t *testing.T) {
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
		branch, revision, err := parseGit17LsRemote(test.in)
		if got, want := err, test.wantErr; !reflect.DeepEqual(got, want) {
			t.Errorf("got %#v, want %#v", got, want)
		}
		if test.wantErr != nil {
			continue
		}

		if got, want := branch, test.wantBranch; got != want {
			t.Errorf("got branch %q, want %q", got, want)
		}
		if got, want := revision, test.wantRevision; got != want {
			t.Errorf("got revision %q, want %q", got, want)
		}
	}
}
