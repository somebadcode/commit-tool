package revisionflag

import (
	"github.com/go-git/go-git/v5/plumbing"
)

const (
	FlagType = "revision"
)

type Revision struct {
	Rev plumbing.Revision
}

func New(rev plumbing.Revision) *Revision {
	return &Revision{
		Rev: rev,
	}
}

func (r *Revision) String() string {
	return r.Rev.String()
}

func (r *Revision) Set(s string) error {
	r.Rev = plumbing.Revision(s)

	return nil
}

func (r *Revision) Type() string {
	return FlagType
}
