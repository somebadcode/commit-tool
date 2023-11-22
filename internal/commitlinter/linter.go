package commitlinter

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

type ReportFunc func(err error)
type StopFunc func(commit *object.Commit) bool

type Linter interface {
	Lint(commit *object.Commit) error
}

type CommitLinter struct {
	// Repo is the repository whose commits should be linted.
	Repo *git.Repository
	// Rev is the start of the linter. Defaults to HEAD.
	Rev plumbing.Revision
	// UntilRev is the revision where the linter will stop. Will only work if StopFunc is nil.
	UntilRev plumbing.Revision
	// ReportFunc is called after each call to Linter if it returned an error.
	ReportFunc ReportFunc
	// Linter is called for each commit.
	Linter Linter
	// StopFunc is called before Linter is called. Determines if the linting should stop.
	StopFunc StopFunc
}

var (
	ErrRepositoryRequired = errors.New("repository is required")
	ErrViolationsFound    = errors.New("violations found")
	ErrNoLinter           = errors.New("no linter has been specified")
)

func NoReporting(_ error) {}

func setDefaults(l *CommitLinter) error {
	if l.Rev == "" {
		l.Rev = plumbing.Revision(plumbing.HEAD)
	}

	if l.ReportFunc == nil {
		l.ReportFunc = NoReporting
	}

	if l.UntilRev != "" && l.StopFunc == nil {
		hash, err := l.Repo.ResolveRevision(l.UntilRev)
		if err != nil {
			return fmt.Errorf("bad stop revision: %w", err)
		}

		l.StopFunc = func(commit *object.Commit) bool {
			return commit.Hash == *hash
		}
	}

	if l.StopFunc == nil {
		l.StopFunc = NoStop()
	}

	return nil
}

func NoStop() StopFunc {
	return func(_ *object.Commit) bool {
		return false
	}
}

func StopAfterN(n uint) StopFunc {
	return func(_ *object.Commit) bool {
		n--
		return n == 0
	}
}

// Validate will verify that required values are set and sets default values.
func (l *CommitLinter) Validate() error {
	if l.Repo == nil {
		return ErrRepositoryRequired
	}

	if l.Linter == nil {
		return ErrNoLinter
	}

	if err := setDefaults(l); err != nil {
		return err
	}

	return nil
}

func (l *CommitLinter) Run(ctx context.Context) error {
	if err := l.Validate(); err != nil {
		return err
	}

	hash, err := l.Repo.ResolveRevision(l.Rev)
	if err != nil {
		return fmt.Errorf("unable to resolve revision: %w", err)
	}

	var iter object.CommitIter
	iter, err = l.Repo.Log(&git.LogOptions{
		From:  *hash,
		Order: git.LogOrderBSF,
	})
	if err != nil {
		return fmt.Errorf("could not iterate over commits: %w", err)
	}
	defer iter.Close()

	var errorCount uint
	err = iter.ForEach(func(commit *object.Commit) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if l.StopFunc(commit) {
			return storer.ErrStop
		}

		lintErr := l.Linter.Lint(commit)
		if lintErr != nil {
			errorCount += 1

			l.ReportFunc(lintErr)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("linter failed: %w", err)
	}

	if errorCount > 0 {
		return fmt.Errorf("%d commits with violations: %w", errorCount, ErrViolationsFound)
	}

	return nil
}
