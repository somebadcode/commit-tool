package linter

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"log/slog"
	"slices"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type ReportFunc func(err error)
type StopFunc func(commit *object.Commit) bool

type CommitLinter interface {
	Lint(*object.Commit) error
}

type Linter struct {
	// Repo is the repository whose commits should be traversed.
	Repo *git.Repository
	// Rev is the revision of where to start the traversal at. Defaults to HEAD.
	Rev plumbing.Revision
	// OtherRev is the revision of a commit whose common ancestor the traversal should stop at.
	OtherRev plumbing.Revision
	// ReportFunc is called after each call to CommitLinter if it returned an error.
	ReportFunc ReportFunc
	// CommitLinter is called for each commit.
	CommitLinter CommitLinter
	// StopFunc is called before CommitLinter is called. Determines if the traversal should stop.
	StopFunc StopFunc

	Logger *slog.Logger
}

var (
	ErrRepositoryRequired = errors.New("repository is required")
	ErrNoLinter           = errors.New("no linter")
)

// Validate will verify that required values are set and sets default values.
func (l *Linter) Validate() error {
	if l.Repo == nil {
		return ErrRepositoryRequired
	}

	if l.CommitLinter == nil {
		return ErrNoLinter
	}

	if l.Rev == "" {
		l.Rev = plumbing.Revision(plumbing.HEAD)
	}

	if l.OtherRev != "" && l.StopFunc == nil {
		stopFunc, err := stopAtActualOther(l.Repo, l.Rev, l.OtherRev)
		if err != nil {
			return err
		}

		l.StopFunc = stopFunc
	}

	if l.ReportFunc == nil {
		l.ReportFunc = NoReporting
	}

	if l.StopFunc == nil {
		l.StopFunc = NoStop()
	}

	if l.Logger == nil {
		l.Logger = slog.New(slog.DiscardHandler)
	}

	return nil
}

// Run will traverse the commit tree, calls [Linter.CommitLinter.Lint] for each commit message.
func (l *Linter) Run(ctx context.Context) error {
	if err := l.Validate(); err != nil {
		return err
	}

	hash, err := l.Repo.ResolveRevision(l.Rev)
	if err != nil {
		return fmt.Errorf("unable to resolve revision: %w", err)
	}

	if l.Logger.Enabled(ctx, slog.LevelDebug) {
		l.Logger.LogAttrs(ctx, slog.LevelDebug, "resolved starting revision",
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

	var accumulatedErrs []error

	err = iter.ForEach(func(commit *object.Commit) error {
		if ctx.Err() != nil {
			if l.Logger.Enabled(ctx, slog.LevelDebug) {
				l.Logger.LogAttrs(ctx, slog.LevelDebug, "cancelling linting",
					slog.String("hash", commit.Hash.String()),
					slog.String("cause", ctx.Err().Error()),
				)
			}

			return &Error{errs: accumulatedErrs}
		}

		if l.StopFunc(commit) {
			if l.Logger.Enabled(ctx, slog.LevelDebug) {
				l.Logger.LogAttrs(ctx, slog.LevelDebug, "stopping linting",
					slog.String("hash", commit.Hash.String()),
				)
			}

			return storer.ErrStop
		}

		err = l.CommitLinter.Lint(commit)
		if err != nil {
			accumulatedErrs = append(accumulatedErrs, err)

			l.ReportFunc(err)

			return nil
		}

		if l.Logger.Enabled(ctx, slog.LevelDebug) {
			l.Logger.LogAttrs(ctx, slog.LevelDebug, "commit passed",
				slog.Any("commit", commit),
			)
		}

		return nil
	})

	if err != nil && !errors.Is(err, storer.ErrStop) {
		return fmt.Errorf("failure: %w", err)
	}

	if len(accumulatedErrs) > 0 {
		return &Error{errs: accumulatedErrs}
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
		return slices.IndexFunc(mergeBases, func(c *object.Commit) bool {
			return c.Hash == commit.Hash
		}) >= 0
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

// NoReporting will cause the linter to not report any of the lint errors, but the linter will still return an error
// if any commit doesn't adhere to the linter's expectations.
func NoReporting(_ error) {}

// SlogReporter will log linter errors using [log/slog].
func SlogReporter(logger *slog.Logger) ReportFunc {
	return func(err error) {
		var lintError LintError
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
