package traverser

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"

	"github.com/somebadcode/commit-tool/slogctx"
)

type ReportFunc func(err error)
type StopFunc func(commit *object.Commit) bool
type VisitFunc func(commit *object.Commit) error

type Traverser struct {
	// Repo is the repository whose commits should be traversed.
	Repo *git.Repository
	// Rev is the revision of where to start the traversal at. Defaults to HEAD.
	Rev plumbing.Revision
	// OtherRev is the revision of a commit whose common ancestor the traversal should stop at.
	OtherRev plumbing.Revision
	// ReportFunc is called after each call to VisitFunc if it returned an error.
	ReportFunc ReportFunc
	// VisitFunc is called for each commit.
	VisitFunc VisitFunc
	// StopFunc is called before VisitFunc is called. Determines if the traversal should stop.
	StopFunc StopFunc
}

var (
	ErrRepositoryRequired = errors.New("repository is required")
	ErrViolationsFound    = errors.New("violations found")
	ErrNoVisitFunc        = errors.New("no traversal function")
)

// NoReporting will cause the linter to not report any of the lint errors, but the linter will still return an error
// if any commit doesn't adhere to the linter's expectations.
func NoReporting(_ error) {}

// SlogReporter will log linter errors using [log/slog].
func SlogReporter(logger *slog.Logger) ReportFunc {
	return func(err error) {
		var lintError TraverseError
		if errors.As(err, &lintError) {
			logger.LogAttrs(context.Background(), slog.LevelError, "bad commit message",
				slog.String("hash", lintError.Hash.String()),
				slog.Int("pos", lintError.Pos),
				slog.String("err", errors.Unwrap(err).Error()),
			)

			return
		}

		logger.LogAttrs(context.Background(), slog.LevelError, "bad commit message",
			slog.String("err", errors.Unwrap(err).Error()),
		)

	}
}

func setDefaults(l *Traverser) error {
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
func (l *Traverser) Validate() error {
	if l.Repo == nil {
		return ErrRepositoryRequired
	}

	if l.VisitFunc == nil {
		return ErrNoVisitFunc
	}

	if err := setDefaults(l); err != nil {
		return err
	}

	return nil
}

// Run will traverse the commit tree, calls Traverser.VisitFunc for each commit message.
func (l *Traverser) Run(ctx context.Context) error {
	if err := l.Validate(); err != nil {
		return err
	}

	hash, err := l.Repo.ResolveRevision(l.Rev)
	if err != nil {
		return fmt.Errorf("unable to resolve revision: %w", err)
	}

	logger := slogctx.L(ctx)

	if logger.Enabled(ctx, slog.LevelDebug) {
		logger.LogAttrs(ctx, slog.LevelDebug, "resolved starting revision",
			slog.String("revision", l.Rev.String()),
			slog.String("hash", hash.String()),
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
			if logger.Enabled(ctx, slog.LevelDebug) {
				logger.LogAttrs(ctx, slog.LevelDebug, "cancelling traversal",
					slog.String("hash", commit.Hash.String()),
					slog.String("cause", ctx.Err().Error()),
				)
			}

			return ctx.Err()
		}

		if l.StopFunc(commit) {
			if logger.Enabled(ctx, slog.LevelDebug) {
				logger.LogAttrs(ctx, slog.LevelDebug, "stopping traversal",
					slog.String("hash", commit.Hash.String()),
				)
			}

			return storer.ErrStop
		}

		err = l.VisitFunc(commit)
		if err != nil {
			errorCount += 1

			l.ReportFunc(err)

			return nil
		}

		if logger.Enabled(ctx, slog.LevelDebug) {
			logger.LogAttrs(ctx, slog.LevelDebug, "commit passed",
				slog.Any("commit", commit),
			)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failure: %w", err)
	}

	if errorCount > 0 {
		return fmt.Errorf("%d commits with violations: %w", errorCount, ErrViolationsFound)
	}

	return nil
}
