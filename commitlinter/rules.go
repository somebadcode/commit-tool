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
)

type RuleFunc func(message commitparser.CommitMessage, commit *object.Commit) error

type Rules []RuleFunc

var (
	ErrInvalidSubject   = errors.New("invalid subject in commit message")
	ErrInvalidCharacter = errors.New("invalid character in commit message")
	ErrInvalidType      = errors.New("invalid type in commit message")
)

func (rules Rules) Validate(message commitparser.CommitMessage, commit *object.Commit) error {
	for _, rule := range rules {
		if err := rule(message, commit); err != nil {
			return err
		}
	}

	return nil
}
