/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package commitparser

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
	commit CommitMessage
	msg    string
	err    error
	start  int
	pos    int
	size   int
}

var (
	ErrInvalidType    = errors.New("invalid commit type")
	ErrInvalidScope   = errors.New("invalid commit scope")
	ErrInvalidSubject = errors.New("invalid commit subject")
	ErrInvalidMessage = errors.New("invalid commit message")
	ErrInvalidTrailer = errors.New("invalid trailer in commit message")
)

const (
	TrailerKeyBreakingChangePrefix = "BREAKING"
	TrailerKeyBreakingChange       = TrailerKeyBreakingChangePrefix + " CHANGE"
	TrailerKeyBreakingChangeAlt    = TrailerKeyBreakingChangePrefix + "-CHANGE"
)

func Parse(message string) (CommitMessage, error) {
	return (&parser{msg: message}).parse()
}

func (p *parser) parse() (CommitMessage, error) {
	for state := parseType(p); state != nil; state = state(p) {
		// Simple state machine.
	}

	// Breaking change might be flagged using a trailer key.
	for _, tr := range p.commit.Trailers {
		if tr.Key == TrailerKeyBreakingChange || tr.Key == TrailerKeyBreakingChangeAlt {
			p.commit.Breaking = true
			break
		}
	}

	if len(p.commit.Trailers) == 0 {
		p.commit.Trailers = nil
	}

	switch p.commit.Type {
	case "revert":
		p.commit.Revert = true
	case "merge":
		p.commit.Merge = true
	}

	if p.err == io.EOF {
		p.err = nil
	}

	return p.commit, p.err
}

func failParsing(p *parser, err error) stateFunc {
	p.err = &ParseError{
		err: err,
		Pos: p.pos,
	}

	return nil
}

func parseType(p *parser) stateFunc {
	r := p.acceptUntil(":!(")
	if r == utf8.RuneError {
		return failParsing(p, ErrInvalidType)
	}

	p.commit.Type = p.token()

	if p.commit.Type == "" {
		return failParsing(p, ErrInvalidType)
	}

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
		return failParsing(p, fmt.Errorf("expected ':', got %c: %w", r, ErrInvalidTrailer))
	}

	if r := p.next(); r != ' ' {
		return failParsing(p, fmt.Errorf("expected ' ', got %c: %w", r, ErrInvalidTrailer))
	}

	p.skip()

	r := p.acceptUntil("\n")
	if r != utf8.RuneError {
		p.back()
	}

	p.commit.Subject = p.token()

	if p.commit.Subject == "" {
		return failParsing(p, ErrInvalidSubject)
	}

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
		if n := countTailers(strings.TrimSpace(p.remains())); n != -1 {
			p.commit.Trailers = make([]Trailer, 0, n)

			return parseTrailer
		}

		p.commit.Body = strings.TrimSpace(p.remains())

		return nil
	}

	if n := countTailers(p.remains()[i+2:]); n != -1 {
		p.commit.Body = strings.TrimSpace(p.remains()[:i])
		p.skipN(i + 2)

		p.commit.Trailers = make([]Trailer, 0, n)

		return parseTrailer
	}

	p.commit.Body = strings.TrimSpace(p.remains())

	return nil
}

func parseTrailer(p *parser) stateFunc {
	var r rune

	if strings.HasPrefix(p.remains(), TrailerKeyBreakingChange) || strings.HasPrefix(p.text(), TrailerKeyBreakingChangeAlt) {
		r = p.acceptUntil(":")
	} else {
		r = p.acceptUntil(": ")
	}

	if r == utf8.RuneError {
		return nil
	}

	key := strings.TrimSpace(p.token())
	if key == "" {
		return failParsing(p, fmt.Errorf("trailer key is empty: %w", ErrInvalidTrailer))
	}

	p.next()

	switch r {
	case ' ':
		if r = p.Peek(); r != '#' {
			return failParsing(p, fmt.Errorf("expected a pound/hash sign, got '%c': %w", r, ErrInvalidTrailer))
		}

	case ':':
		if r = p.Peek(); r != ' ' {
			return failParsing(p, fmt.Errorf("expected space, got '%c': %w", r, ErrInvalidTrailer))
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

	value := strings.TrimSpace(p.text())
	if value == "" {
		return failParsing(p, fmt.Errorf("empty value: %w", ErrInvalidTrailer))
	}

	if key == TrailerKeyBreakingChange || key == TrailerKeyBreakingChangeAlt {
		p.commit.Breaking = true
	}

	if strings.ContainsAny(key, "\r") {
		return failParsing(p, fmt.Errorf("trailer key is invalid (found carriage return: %w", ErrInvalidTrailer))
	}

	if strings.Contains(value, "\r") {
		return failParsing(p, fmt.Errorf("trailer value is invalid (found carriage return: %w", ErrInvalidTrailer))
	}

	p.commit.Trailers = append(p.commit.Trailers, Trailer{
		Key:   key,
		Value: value,
	})

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
	if r == utf8.RuneError && p.size == 0 {
		p.err = io.EOF
		return utf8.RuneError
	} else if r == utf8.RuneError {
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

func (p *parser) skipN(n int) {
	p.pos += n
	p.skip()
}
