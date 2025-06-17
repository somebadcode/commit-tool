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
