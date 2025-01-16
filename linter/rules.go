package linter

import (
	"errors"
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/somebadcode/commit-tool/commitparser"
)

type RuleFunc func(message commitparser.CommitMessage, commit *object.Commit) error

type Rules []RuleFunc

var (
	ErrInvalidSubject   = errors.New("invalid subject in commit message")
	ErrInvalidCharacter = errors.New("invalid character in commit message")
)

func (rules Rules) Validate(message commitparser.CommitMessage, commit *object.Commit) error {
	for _, rule := range rules {
		if err := rule(message, commit); err != nil {
			return err
		}
	}

	return nil
}

// RuleConventionalSubject verifies that the commit message's subject is not empty and does not start with upper case.
func RuleConventionalSubject(msg commitparser.CommitMessage, _ *object.Commit) error {
	first, size := utf8.DecodeRuneInString(msg.Subject)
	if first == utf8.RuneError && size == 0 {
		return fmt.Errorf("subject must not be empty: %w", ErrInvalidSubject)
	} else if first == utf8.RuneError && size == 1 {
		return fmt.Errorf("bad subject: %w", ErrInvalidCharacter)
	}

	if unicode.IsUpper(first) {
		return fmt.Errorf("subject must not start with upper case %q: %w", msg.Subject, ErrInvalidCharacter)
	}

	return nil
}

func RuleConventionalCommit(msg commitparser.CommitMessage, commit *object.Commit) error {
	if err := RuleConventionalSubject(msg, commit); err != nil {
		return err
	}

	return nil
}
