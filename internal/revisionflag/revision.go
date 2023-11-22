package revisionflag

import (
	"flag"

	"github.com/go-git/go-git/v5/plumbing"
)

const Usage string = "git revision (https://mirrors.edge.kernel.org/pub/software/scm/git/docs/gitrevisions.html)"

type Revision struct {
	name  string
	rev   plumbing.Revision
	isSet bool
}

func New(name string, value plumbing.Revision) *Revision {
	return &Revision{
		name: name,
		rev:  value,
	}
}

func (r *Revision) MarshalText() (text []byte, err error) {
	r.isSet = true
	return []byte(r.rev), nil
}

func (r *Revision) UnmarshalText(text []byte) error {
	r.rev = plumbing.Revision(text)

	return nil
}

func (r *Revision) Revision() plumbing.Revision {
	return r.rev
}

func (r *Revision) Apply(set *flag.FlagSet) error {
	set.TextVar(r, r.name, r, Usage)

	return nil
}

func (r *Revision) Names() []string {
	return []string{"revision"}
}

func (r *Revision) IsSet() bool {
	return r.isSet
}

func (r *Revision) String() string {
	return r.rev.String()
}
