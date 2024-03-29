package linter

import (
	"errors"

	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/somebadcode/commit-tool/commitparser"
	"github.com/somebadcode/commit-tool/traverser"
)

type Linter struct {
	Filters Filters
	Rules   Rules
}

// Lint commit to ensures that it adheres to Conventional Commits 1.0.0.
func (l Linter) Lint(commit *object.Commit) error {
	msg, err := commitparser.Parse(commit.Message)
	if err != nil {
		if l.Filters.Filter(msg, commit, err) == nil {
			return nil
		}

		var parseError commitparser.ParseError
		if errors.As(err, &parseError) {
			return traverser.TraverseError{
				Err:  parseError,
				Hash: commit.Hash,
				Pos:  parseError.Pos,
			}
		}

		return err
	}

	// TODO: Add support for linting revert and merge commits.
	if msg.Revert || msg.Merge {
		return nil
	}

	if err = l.Rules.Validate(msg, commit); err != nil {
		return traverser.TraverseError{
			Err:  err,
			Hash: commit.Hash,
		}
	}

	return nil
}
