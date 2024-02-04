package defaultlinter

import (
	"errors"
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/somebadcode/commit-tool/commitmsg"
)

type RuleFunc func(message commitmsg.CommitMessage) error

var (
	ErrInvalidSubject   = errors.New("invalid subject in commit message")
	ErrInvalidCharacter = errors.New("invalid character in commit message")
)

func RuleSubjectNoLeadingUpperCase(msg commitmsg.CommitMessage) error {
	first, size := utf8.DecodeRuneInString(msg.Subject)
	if first == utf8.RuneError && size == 0 {
		return fmt.Errorf("no subject detected: %w", ErrInvalidSubject)
	} else if first == utf8.RuneError && size == 1 {
		return fmt.Errorf("bad subject: %w", ErrInvalidCharacter)
	}

	if unicode.IsUpper(first) {
		return fmt.Errorf("subject must not start with upper case %q: %w", msg.Subject, ErrInvalidCharacter)
	}

	return nil
}
