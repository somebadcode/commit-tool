/*
 * Copyright 2025 Tobias Dahlberg
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package commitlinter

import (
	"errors"

	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/somebadcode/commit-tool/commitparser"
	"github.com/somebadcode/commit-tool/linter"
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
