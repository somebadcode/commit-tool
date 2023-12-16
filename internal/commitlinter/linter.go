package commitlinter

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"go.uber.org/zap"

	"github.com/somebadcode/commit-tool/internal/zapctx"
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
	// OtherRev is the revision of a commit whose common ancestor the linting should stop at.
	OtherRev plumbing.Revision
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

// NoReporting will cause the linter to not report any of the lint errors, but the linter will still return an error
// if any commit doesn't adhere to the linter's expectations.
func NoReporting(_ error) {}

// ZapReporter will cause the linter to report lint errors using the specified Zap logger.
func ZapReporter(logger *zap.Logger) ReportFunc {
	return func(err error) {
		var lintError LintError
		if errors.As(err, &lintError) {
			logger.Error("bad commit message",
				zap.Stringer("hash", lintError.Hash),
				zap.Int("pos", lintError.Pos),
				zap.Error(errors.Unwrap(err)),
			)

			return
		}

		logger.Error("bad commit message",
			zap.Error(err),
		)
	}
}

func setDefaults(l *CommitLinter) error {
	if l.Rev == "" {
		l.Rev = plumbing.Revision(plumbing.HEAD)
	}

	if l.ReportFunc == nil {
		l.ReportFunc = NoReporting
	}

	if l.OtherRev != "" && l.StopFunc == nil {
		stopFunc, err := stopAtActualOther(l.Repo, l.Rev, l.OtherRev)
		if err != nil {
			return err
		}

		l.StopFunc = stopFunc
	}

	if l.StopFunc == nil {
		l.StopFunc = NoStop()
	}

	return nil
}

func stopAtActualOther(repo *git.Repository, rev plumbing.Revision, otherRev plumbing.Revision) (StopFunc, error) {
	// Get start commit.
	hash, err := repo.ResolveRevision(rev)
	if err != nil {
		return nil, fmt.Errorf("bad revision %q: %w", otherRev, err)
	}

	var commit *object.Commit
	commit, err = repo.CommitObject(*hash)
	if err != nil {
		return nil, fmt.Errorf("bad revision %q: %w", otherRev, err)
	}

	// Get other commit.
	hash, err = repo.ResolveRevision(otherRev)
	if err != nil {
		return nil, fmt.Errorf("bad revision %q: %w", otherRev, err)
	}

	var other *object.Commit
	other, err = repo.CommitObject(*hash)
	if err != nil {
		return nil, fmt.Errorf("bad revision %q: %w", otherRev, err)
	}

	var mergeBases []*object.Commit
	mergeBases, err = commit.MergeBase(other)
	if err != nil {
		return nil, fmt.Errorf("revisions %q and %q do not have a common ancestor: %w", rev, otherRev, err)
	}

	return StopAtMergeBases(mergeBases), nil
}

// StopAtMergeBases will cause the linter to stop at any of the merge bases.
func StopAtMergeBases(mergeBases []*object.Commit) StopFunc {
	return func(commit *object.Commit) bool {
		for _, ancestor := range mergeBases {
			if ancestor.Hash == commit.Hash {
				return true
			}
		}

		return false
	}
}

// NoStop will cause the linter to never stop. Allows the linter to run until the last commit.
func NoStop() StopFunc {
	return func(_ *object.Commit) bool {
		return false
	}
}

// StopAfterN will cause the linter to stop after N commits.
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

// Run will traverse the commit tree, lint each commit message. The structure of the tree is ignored.
func (l *CommitLinter) Run(ctx context.Context) error {
	if err := l.Validate(); err != nil {
		return err
	}

	hash, err := l.Repo.ResolveRevision(l.Rev)
	if err != nil {
		return fmt.Errorf("unable to resolve revision: %w", err)
	}

	logger := zapctx.L(ctx)

	if ev := logger.Check(zap.DebugLevel, "resolved starting revision"); ev != nil {
		ev.Write(
			zap.Stringer("revision", l.Rev),
			zap.Stringer("hash", hash),
		)
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
			if ev := logger.Check(zap.DebugLevel, "cancelling linting"); ev != nil {
				ev.Write(
					zap.Stringer("hash", commit.Hash),
					zap.NamedError("cause", ctx.Err()),
				)
			}

			return ctx.Err()
		}

		if l.StopFunc(commit) {
			if ev := logger.Check(zap.DebugLevel, "stopping linting"); ev != nil {
				ev.Write(zap.Stringer("hash", commit.Hash))
			}

			return storer.ErrStop
		}

		lintErr := l.Linter.Lint(commit)
		if lintErr != nil {
			errorCount += 1

			l.ReportFunc(lintErr)

			return nil
		}

		if ev := logger.Check(zap.DebugLevel, "commit passed"); ev != nil {
			ev.Write(
				zap.Stringer("hash", commit.Hash),
				zap.Dict("author",
					zap.String("name", commit.Author.Name),
					zap.String("email", commit.Author.Email),
					zap.Time("when", commit.Author.When),
				),
			)
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
