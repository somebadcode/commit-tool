package linter

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"
)

type LintError struct {
	Err  error
	Hash plumbing.Hash
	Pos  int
}

func (err LintError) Unwrap() error {
	return err.Err
}

func (err LintError) Error() string {
	return fmt.Sprintf("bad commit message at %s on line %d: %s", err.Hash, err.Pos, err.Err)
}

type Error struct {
	errs []error
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d violations found", len(e.errs))
}

func (e *Error) Unwrap() []error {
	return e.errs
}
