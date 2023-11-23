package defaultlinter

import (
	"errors"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/somebadcode/commit-tool/internal/commitlinter"
	"github.com/somebadcode/commit-tool/internal/commitmsg"
)

type Linter struct {
	// AllowInitialCommit will cause the linter to ignore initial commit if it's written correctly.
	AllowInitialCommit bool
	// AllowedTypes is a set of commit types that should be allowed. Defaults to conventional commits and angular types.
	AllowedTypes map[string]struct{}
}

var (
	ErrUnconventionalType = errors.New("non-conforming commit type")
)

var (
	conventionalCommitTypes = []string{"fix", "feat", "BREAKING CHANGE"}
	angularCommitTypes      = []string{"docs", "ci", "chore", "style", "refactor", "improvement", "perf", "test"}
)

func New(options ...Option) *Linter {
	var linter Linter
	for _, opt := range options {
		opt(&linter)
	}

	return &linter
}

func (l *Linter) Lint(commit *object.Commit) error {
	if len(l.AllowedTypes) == 0 {
		WithCommitTypes(conventionalCommitTypes)(l)
		WithCommitTypes(angularCommitTypes)(l)
	}

	msg, err := commitmsg.Parse(commit.Message)
	if err != nil {
		// Ignore the error if the commit has no parents and is "initial commit"
		if commit.NumParents() == 0 && strings.EqualFold(commit.Message, "initial commit") {
			return nil
		}

		var parseError commitmsg.ParseError
		if errors.As(err, &parseError) {
			return commitlinter.LintError{
				Err:  parseError,
				Hash: commit.Hash,
				Pos:  parseError.Pos,
			}
		} else {
			return err
		}
	}

	// TODO: Add support for linting revert and merge commits.
	if msg.Revert || msg.Merge {
		return nil
	}

	if _, ok := l.AllowedTypes[msg.Type]; !ok {
		return commitlinter.LintError{
			Err:  ErrUnconventionalType,
			Hash: commit.Hash,
		}
	}

	return nil
}
