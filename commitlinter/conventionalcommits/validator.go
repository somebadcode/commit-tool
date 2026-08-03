/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package conventionalcommits

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-git/go-git/v5/plumbing/object"

	"codeberg.org/somebadcode/commit-tool/commitparser"
)

var conventionalTypes = map[string]struct{}{
	"build":    {},
	"chore":    {},
	"ci":       {},
	"docs":     {},
	"feat":     {},
	"fix":      {},
	"perf":     {},
	"refactor": {},
	"revert":   {},
	"style":    {},
	"test":     {},
}

type Settings struct {
	MaxHeaderLength uint
}

type Validator struct {
	settings Settings
}

const (
	DefaultMaxHeaderLength = 72
)

var (
	ErrInvalidSubject   = errors.New("invalid subject in commit message")
	ErrInvalidCharacter = errors.New("invalid character in commit message")
	ErrInvalidType      = errors.New("invalid type in commit message")
	ErrInvalidScope     = errors.New("invalid scope in commit message")
)

func New(s Settings) *Validator {
	if s.MaxHeaderLength == 0 {
		s.MaxHeaderLength = DefaultMaxHeaderLength
	}

	return &Validator{
		settings: s,
	}
}

// ValidateSubject validates that the commit message's subject is not empty and does not start with upper case.
func (v *Validator) ValidateSubject(msg commitparser.CommitMessage, _ *object.Commit) error {
	if msg.Subject == "" {
		return fmt.Errorf("subject must not be empty: %w", ErrInvalidSubject)
	}

	first, size := utf8.DecodeRuneInString(msg.Subject)
	if first == utf8.RuneError {
		return fmt.Errorf("bad subject: %w", ErrInvalidCharacter)
	}

	if unicode.IsSpace(first) {
		return fmt.Errorf("subject must not start with space %q: %w", msg.Subject, ErrInvalidCharacter)
	}

	if unicode.IsUpper(first) {
		return fmt.Errorf("subject must not start with upper case %q: %w", msg.Subject, ErrInvalidCharacter)
	}

	var last rune

	last, size = utf8.DecodeLastRuneInString(msg.Subject)
	if last == utf8.RuneError && size == 1 {
		return fmt.Errorf("bad subject: %w", ErrInvalidCharacter)
	}

	if unicode.IsPunct(last) && last != '/' {
		return fmt.Errorf("subject must not end with punctuation %q: %w", msg.Subject, ErrInvalidCharacter)
	}

	return nil
}

func (v *Validator) ValidateType(msg commitparser.CommitMessage, _ *object.Commit) error {
	if _, found := conventionalTypes[msg.Type]; !found {
		return fmt.Errorf("unknown type %q: %w", msg.Type, ErrInvalidType)
	}

	return nil
}

func (v *Validator) ValidateScope(msg commitparser.CommitMessage, _ *object.Commit) error {
	if strings.TrimSpace(msg.Scope) != msg.Scope {
		return fmt.Errorf("trailing or leading space in commit message scope: %w", ErrInvalidScope)
	}

	return nil
}

func (v *Validator) Validate(msg commitparser.CommitMessage, commit *object.Commit) error {
	sizeType := utf8.RuneCountInString(msg.Type) + 2   // includes `: `.
	sizeScope := utf8.RuneCountInString(msg.Scope) + 2 // includes `()`
	sizeSubject := utf8.RuneCountInString(msg.Subject)

	size := sizeType + sizeScope + sizeSubject
	if size > int(v.settings.MaxHeaderLength) {
		return fmt.Errorf("commit header too long (%d>%d): %w", size, v.settings.MaxHeaderLength, ErrInvalidScope)
	}

	if err := v.ValidateType(msg, commit); err != nil {
		return err
	}

	if err := v.ValidateSubject(msg, commit); err != nil {
		return err
	}

	if err := v.ValidateScope(msg, commit); err != nil {
		return err
	}

	return nil
}
