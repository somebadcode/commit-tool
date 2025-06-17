/*
 * This file is part of commit-tool which is released under EUPL 1.2.
 * See the file LICENSE in the repository root for full license details.
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package commitlinter

import (
	"errors"

	"github.com/go-git/go-git/v5/plumbing/object"

	"codeberg.org/somebadcode/commit-tool/commitparser"
	"codeberg.org/somebadcode/commit-tool/linter"
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
			return linter.LintError{
				Err:  parseError,
				Hash: commit.Hash,
				Pos:  parseError.Pos,
			}
		}

		return err
	}

	if err = l.Rules.Validate(msg, commit); err != nil {
		return linter.LintError{
			Err:  err,
			Hash: commit.Hash,
		}
	}

	return nil
}
