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
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/somebadcode/commit-tool/commitparser"
)

type FilterFunc func(message commitparser.CommitMessage, commit *object.Commit, err error) error

type Filters []FilterFunc

const (
	initialCommit = "initial commit"
)

func (filters Filters) Filter(msg commitparser.CommitMessage, commit *object.Commit, err error) error {
	for _, filter := range filters {
		if filter(msg, commit, err) == nil {
			return nil
		}
	}

	return err
}

func FilterInitialCommit(_ commitparser.CommitMessage, commit *object.Commit, err error) error {
	// If the commit has one or more parents then there's nothing to do and the error should be returned.
	if commit.NumParents() > 0 {
		return err
	}

	if !strings.EqualFold(commit.Message, initialCommit) {
		return fmt.Errorf("expected commit message %q: %w", initialCommit, err)
	}

	return nil
}
