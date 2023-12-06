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
	// Rules is a list of rules to enforce on commit messages.
	Rules []RuleFunc
}

// New returns a commit linter that adheres to Conventional Commits 1.0.0 and Angular project's recommended scopes.
func New(options ...Option) *Linter {
	var linter Linter

	for _, opt := range options {
		opt(&linter)
	}

	if linter.Rules == nil {
		linter.Rules = []RuleFunc{
			RuleSubjectNoLeadingUpperCase,
		}
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

	for _, rule := range l.Rules {
		if err = rule(msg); err != nil {
			return commitlinter.LintError{
				Err:  err,
				Hash: commit.Hash,
			}
		}
	}

	return nil
}
