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

package conventionalcommits

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/somebadcode/commit-tool/commitlinter"
	"github.com/somebadcode/commit-tool/commitparser"
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

// VerifySubject verifies that the commit message's subject is not empty and does not start with upper case.
func VerifySubject(msg commitparser.CommitMessage, _ *object.Commit) error {
	first, size := utf8.DecodeRuneInString(msg.Subject)
	if first == utf8.RuneError && size == 0 {
		return fmt.Errorf("subject must not be empty: %w", commitlinter.ErrInvalidSubject)
	} else if first == utf8.RuneError && size == 1 {
		return fmt.Errorf("bad subject: %w", commitlinter.ErrInvalidCharacter)
	}

	if unicode.IsUpper(first) {
		return fmt.Errorf("subject must not start with upper case %q: %w", msg.Subject, commitlinter.ErrInvalidCharacter)
	}

	if unicode.IsSpace(first) {
		return fmt.Errorf("subject must not start with space %q: %w", msg.Subject, commitlinter.ErrInvalidCharacter)
	}

	var last rune

	last, size = utf8.DecodeRuneInString(msg.Subject)
	if first == utf8.RuneError && size == 0 {
		return fmt.Errorf("unexpectedly short subject: %w", commitlinter.ErrInvalidSubject)
	} else if first == utf8.RuneError && size == 1 {
		return fmt.Errorf("bad subject: %w", commitlinter.ErrInvalidCharacter)
	}

	if unicode.IsPunct(last) {
		return fmt.Errorf("subject must now end with punctuation %q: %w", msg.Subject, commitlinter.ErrInvalidCharacter)
	}

	return nil
}

func VerifyType(msg commitparser.CommitMessage, _ *object.Commit) error {
	if _, found := conventionalTypes[msg.Type]; !found {
		return fmt.Errorf("unknown type %q: %w", msg.Type, commitlinter.ErrInvalidType)
	}

	return nil
}

func VerifyScope(msg commitparser.CommitMessage, _ *object.Commit) error {
	if strings.TrimSpace(msg.Scope) != msg.Scope {
		return fmt.Errorf("")
	}

	return nil
}

func Verify(msg commitparser.CommitMessage, commit *object.Commit) error {
	if err := VerifyType(msg, commit); err != nil {
		return err
	}

	if err := VerifySubject(msg, commit); err != nil {
		return err
	}

	return nil
}
