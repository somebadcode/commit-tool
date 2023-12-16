package commitmsg

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

type stateFunc func(*parser) stateFunc

type parser struct {
	msg    string
	start  int
	pos    int
	size   int
	commit CommitMessage
	err    error
	char   rune
}

var (
	ErrInvalidType              = errors.New("invalid commit type")
	ErrInvalidScope             = errors.New("invalid commit scope")
	ErrInvalidSubject           = errors.New("invalid commit subject")
	ErrInvalidMessage           = errors.New("invalid commit message")
	ErrInvalidTrailer           = errors.New("invalid trailer in commit message")
	ErrUnsupportedSpecialCommit = errors.New("unsupported special commit, please report this error if you think it should be supported")
)

const (
	TrailerKeyBreakingChangePrefix = "BREAKING"
	TrailerKeyBreakingChange       = TrailerKeyBreakingChangePrefix + " CHANGE"
	TrailerKeyBreakingChangeAlt    = "BREAKING-CHANGE"
)

func Parse(message string) (CommitMessage, error) {
	return (&parser{msg: message}).parse()
}

func (p *parser) parse() (CommitMessage, error) {
	for state := parseSpecial(p); state != nil; state = state(p) {
		// Simple state machine.
	}

	// Breaking change might be flagged using a trailer key.
	for k := range p.commit.Trailers {
		if k == TrailerKeyBreakingChange || k == TrailerKeyBreakingChangeAlt {
			p.commit.Breaking = true
			break
		}
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

	p.acceptUntil("\"")
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
	}

	p.skip()

	i := strings.LastIndex(p.remains(), "\n\n")
	if i == -1 {
		// No more paragraphs found. Using the remains for the message body and stopping.
		p.commit.Body = strings.TrimSpace(p.remains())

		return nil
	}

	// Paragraphs were found, so move position up to the end of the penultimate paragraph and save the body.
	p.pos += i
	p.commit.Body = strings.TrimSpace(p.token())

	// Move position up two positions ("\n\n") and skip.
	p.pos += 2
	p.skip()

	// Parse trailers.
	return parseTrailers
}

func parseTrailers(p *parser) stateFunc {
	// Save the remains in case the last paragraph can't be parsed as git trailers.
	remains := p.remains()

	p.commit.Trailers = make(map[string][]string)

	// Parse the trailers but if the error is ErrInvalidTrailer then we should append the remains to the body.
	// Any other error should bubble up.
	for state := parseTrailer(p); state != nil; state = state(p) {
	}

	if errors.Is(p.err, ErrInvalidTrailer) {
		p.err = nil
		p.commit.Body += "\n\n" + remains
		p.commit.Trailers = nil
	}

	// If there are no trailers, make sure the map is nil.
	if p.commit.Trailers != nil && len(p.commit.Trailers) == 0 {
		p.commit.Trailers = nil
	}

	return nil
}

func parseTrailer(p *parser) stateFunc {
	r := p.acceptUntil(": ")

	// If there's a space and the text matches the breaking change prefix, continue until `:` and make sure that it's
	// a breaking change.
	if r == ' ' && p.text() == TrailerKeyBreakingChangePrefix {
		r = p.acceptUntil(":")
		if p.text() != TrailerKeyBreakingChange {
			return failParsing(p, fmt.Errorf("git trailer key can not contain space, expected key %q: %w", TrailerKeyBreakingChange, ErrInvalidTrailer))
		}
	}

	if r == utf8.RuneError {
		return nil
	}

	key := p.token()

	// Trailer key must start with upper case.
	if first, _ := utf8.DecodeRuneInString(key); !unicode.IsUpper(first) {
		return failParsing(p, fmt.Errorf("git trailer key must start with upper case: %w", ErrInvalidTrailer))
	}

	switch p.next() {
	case ' ':
		if p.Peek() != '#' {
			return failParsing(p, fmt.Errorf("expected a pound/hash sign: %w", ErrInvalidTrailer))
		}

	case ':':
		if p.Peek() != ' ' {
			return failParsing(p, fmt.Errorf("expected a space: %w", ErrInvalidTrailer))
		}

		p.next()
	}

	p.skip()

	for {
		r = p.acceptUntil("\n")
		if r == utf8.RuneError {
			break
		}

		p.next()
		if !unicode.IsSpace(p.Peek()) {
			break
		}
	}

	// If the trailer key is empty then it's not a valid trailer so skip the paragraph and fall back.
	if len(p.text()) == 0 {
		return failParsing(p, fmt.Errorf("git trailer key is empty: %w", ErrInvalidTrailer))
	}

	lines := strings.FieldsFunc(p.token(), func(r rune) bool {
		return r == '\n'
	})

	sb := strings.Builder{}
	defer sb.Reset()

	for _, line := range lines {
		sb.WriteString(strings.TrimSpace(line))
		sb.WriteRune('\n')
	}

	if key == TrailerKeyBreakingChange || key == TrailerKeyBreakingChangeAlt {
		p.commit.Breaking = true
	}

	p.commit.Trailers[key] = append(p.commit.Trailers[key], strings.TrimSpace(sb.String()))
	if r == utf8.RuneError {
		return nil
	}

	p.skip()

	return parseTrailer
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
