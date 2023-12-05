package defaultlinter

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

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
	ErrInvalidSubject     = errors.New("invalid subject in commit message")
	ErrInvalidCharacter   = errors.New("invalid character in commit message")
)

var (
	conventionalCommitTypes = []string{"fix", "feat", "BREAKING CHANGE"}
	angularCommitTypes      = []string{"docs", "ci", "chore", "style", "refactor", "improvement", "perf", "test"}
)

// New returns a commit linter that adheres to Conventional Commits 1.0.0 and Angular project's recommended scopes.
func New(options ...Option) *Linter {
	var linter Linter
	for _, opt := range options {
		opt(&linter)
	}

	if len(linter.AllowedTypes) == 0 {
		WithCommitTypes(conventionalCommitTypes...)(&linter)
		WithCommitTypes(angularCommitTypes...)(&linter)
	}

	return &linter
}

// Lint will lint the commit and return a LintError.
func (l *Linter) Lint(commit *object.Commit) error {
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

	// Commit type must a known one.
	if _, ok := l.AllowedTypes[msg.Type]; !ok {
		return commitlinter.LintError{
			Err:  ErrUnconventionalType,
			Hash: commit.Hash,
		}
	}

	// Subject must not start with an upper case letter.
	first, size := utf8.DecodeRuneInString(msg.Subject)
	if first == utf8.RuneError && size == 0 {
		return commitlinter.LintError{
			Err:  fmt.Errorf("no subject detected: %w", ErrInvalidSubject),
			Hash: commit.Hash,
		}
	} else if first == utf8.RuneError && size == 1 {
		return commitlinter.LintError{
			Err:  fmt.Errorf("bad subject: %w", ErrInvalidCharacter),
			Hash: commit.Hash,
		}
	}

	if unicode.IsUpper(first) {
		return commitlinter.LintError{
			Err:  fmt.Errorf("subject must not start with upper case %q: %w", msg.Subject, ErrInvalidCharacter),
			Hash: commit.Hash,
		}
	}

	return nil
}
