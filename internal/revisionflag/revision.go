package revisionflag

import (
	"github.com/go-git/go-git/v5/plumbing"
)

const (
	FlagType = "revision"
)

type Revision struct {
	rev plumbing.Revision
}

func New(rev plumbing.Revision) *Revision {
	return &Revision{
		rev: rev,
	}
}

func (r *Revision) String() string {
	return r.rev.String()
}

func (r *Revision) Set(s string) error {
	r.rev = plumbing.Revision(s)

	return nil
}

func (r *Revision) Type() string {
	return FlagType
}
