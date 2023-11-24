package commitmsg

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

type stateFunc func(*parser) stateFunc

type parser struct {
	msg    string
	start  int
	pos    int
	char   rune
	size   int
	commit CommitMessage
	err    error
}

var (
	ErrInvalidType              = errors.New("invalid commit type")
	ErrInvalidScope             = errors.New("invalid commit scope")
	ErrInvalidSubject           = errors.New("invalid commit subject")
	ErrInvalidMessage           = errors.New("invalid commit message")
	ErrUnsupportedSpecialCommit = errors.New("unsupported special commit, please report")
)

func Parse(message string) (CommitMessage, error) {
	p := &parser{msg: message}
	return p.parse()
}

func (p *parser) parse() (CommitMessage, error) {
	for state := parseSpecial(p); state != nil; state = state(p) {
		// Simple state machine.
	}

	return p.commit, p.err
}

func failParsing(p *parser, err error) stateFunc {
	p.err = ParseError{
		err: err,
		Pos: p.pos,
	}

	return nil
}

func parseSpecial(p *parser) stateFunc {
	i := strings.IndexRune(p.msg, ' ')
	if i <= 0 {
		return failParsing(p, ErrInvalidType)
	}

	switch p.msg[:i] {
	case "Merge":
		return parseMerge
	case "Revert":
		return parseRevert
	default:
		return parseType
	}
}

func parseRevert(p *parser) stateFunc {
	r := p.acceptUntil(" ")
	if r == utf8.RuneError {
		return failParsing(p, ErrInvalidMessage)
	}

	if p.token() != "Revert" {
		return failParsing(p, ErrInvalidMessage)
	}

	r = p.acceptUntil("\"")
	if p.next() != '"' {
		return failParsing(p, fmt.Errorf("expected a quotation mark (\") after \"Revert \": %w", ErrInvalidMessage))
	}

	p.commit.Revert = true

	p.skip()

	return parseType
}

func parseMerge(p *parser) stateFunc {
	r := p.acceptUntil("\n")
	if r == utf8.RuneError {
		p.back()
	}

	p.commit.Merge = true
	p.commit.Type = "merge"
	p.commit.Subject = p.token()

	return parseBody
}

func parseType(p *parser) stateFunc {
	r := p.acceptUntil(":!(")
	if r == utf8.RuneError {
		return failParsing(p, ErrInvalidType)
	}

	p.commit.Type = p.token()
	p.skip()

	switch p.Peek() {
	case '!':
		return parseBreaking
	case '(':
		return parseScope
	}

	return parseSubject
}

func parseBreaking(p *parser) stateFunc {
	if r := p.next(); r != '!' {
		return failParsing(p, ErrInvalidMessage)
	}
	p.skip()

	p.commit.Breaking = true

	return parseSubject
}

func parseSubject(p *parser) stateFunc {
	if r := p.next(); r != ':' {
		return failParsing(p, ErrInvalidSubject)
	}
	if r := p.next(); r != ' ' {
		return failParsing(p, ErrInvalidSubject)
	}

	p.skip()

	r := p.acceptUntil("\n")
	if r != utf8.RuneError {
		p.back()
	}

	if p.commit.Revert {
		token := p.token()
		if !strings.HasSuffix(token, "\"") {
			return failParsing(p, ErrUnsupportedSpecialCommit)
		}

		p.commit.Subject = token[:len(token)-1]

		return parseBody
	}

	p.commit.Subject = p.token()

	return parseBody
}

func parseBody(p *parser) stateFunc {
	r := p.next()
	if r == utf8.RuneError {
		return nil
	} else if r != '\n' {
		return failParsing(p, ErrInvalidMessage)
	}

	p.skip()

	p.commit.Body = strings.TrimSpace(p.remains())

	return nil
}

func parseScope(p *parser) stateFunc {
	if r := p.next(); r != '(' {
		return failParsing(p, ErrInvalidSubject)
	}
	p.skip()

	r := p.acceptUntil(")")
	if r == utf8.RuneError {
		return failParsing(p, ErrInvalidScope)
	}

	p.commit.Scope = p.token()
	if p.commit.Scope == "" {
		return failParsing(p, fmt.Errorf("parenthesis found but scope is empty: %w", ErrInvalidScope))
	}

	p.next()

	if p.Peek() == '!' {
		return parseBreaking
	}

	return parseSubject
}

func (p *parser) remains() string {
	return p.msg[p.start:]
}

func (p *parser) next() rune {
	if p.err != nil {
		return utf8.RuneError
	}

	var r rune
	r, p.size = utf8.DecodeRuneInString(p.msg[p.pos:])
	if p.char == utf8.RuneError && p.size == 0 {
		p.err = io.EOF
		return utf8.RuneError
	} else if p.char == utf8.RuneError {
		p.err = ErrInvalidMessage
		return utf8.RuneError
	}

	p.pos += p.size

	return r
}

func (p *parser) acceptUntil(set string) rune {
	var r rune
	for r = p.next(); r != utf8.RuneError; r = p.next() {
		if strings.ContainsRune(set, r) {
			p.back()
			return r
		}
	}

	return r
}

func (p *parser) back() {
	if p.size == 0 || p.pos == 0 || p.err != nil {
		return
	}

	p.pos -= p.size
	p.size = 0
}

func (p *parser) Peek() rune {
	r := p.next()
	p.back()
	return r
}

func (p *parser) token() string {
	s := p.text()
	p.skip()

	return s
}

func (p *parser) text() string {
	return p.msg[p.start:p.pos]
}

func (p *parser) skip() {
	p.start = p.pos
}

/*func (p *parser) line() int {
	return strings.Count(p.msg[:p.pos], "\n") + 1
}
*/
