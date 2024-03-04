package traverser

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"
)

type TraverseError struct {
	Err  error
	Hash plumbing.Hash
	Pos  int
}

func (err TraverseError) Unwrap() error {
	return err.Err
}

func (err TraverseError) Error() string {
	return fmt.Sprintf("bad commit message at %s on line %d: %s", err.Hash, err.Pos, err.Err)
}
