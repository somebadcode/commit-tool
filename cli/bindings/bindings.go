/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package bindings

import (
	"encoding"
	"errors"
	"fmt"
	"io/fs"
	"iter"
	"log/slog"
	"maps"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type SignedInteger interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type UnsignedInteger interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type Float interface {
	~float32 | ~float64
}

type URLRef interface{ ~*url.URL | ~**url.URL }

type URL interface{ url.URL | ~*url.URL }

type ValueSetter func(string) error
type Parser[T any] func(string) (T, error)

type Binder interface {
	Bind(*Bindings)
}

type TextUnmarshaler[T any] interface {
	*T
	encoding.TextUnmarshaler
}

type MissingValueError struct {
	Keyword string
}

func (e *MissingValueError) Error() string {
	return fmt.Sprintf("%s: missing value", e.Keyword)
}

type AmbiguousKeywordError struct {
	Keyword string
	Matches []string
}

func (e *AmbiguousKeywordError) Error() string {
	return fmt.Sprintf("%s: ambiguous keyword matches(%s)", e.Keyword, strings.Join(e.Matches, ", "))
}

type UnknownKeywordError struct {
	Keyword string
}

func (e *UnknownKeywordError) Error() string {
	return fmt.Sprintf("%s: unknown keyword", e.Keyword)
}

type MissingKeywordError struct {
	Keyword string
}

func (e *MissingKeywordError) Error() string {
	return fmt.Sprintf("%s: missing keyword", e.Keyword)
}

type entry struct {
	binder      Binder
	sub         *Bindings
	setter      ValueSetter
	description string
	repeatable  bool
	noArg       bool
}

type Bindings struct {
	entries      map[string]entry
	descriptions map[string]string
	required     []string
	seen         map[string]struct{}
}

type Info struct {
	Description string
	Sub         *Bindings
	Required    bool
	Repeatable  bool
	NoArg       bool
}

func New() *Bindings {
	return &Bindings{
		entries:      make(map[string]entry),
		descriptions: make(map[string]string),
		seen:         make(map[string]struct{}),
	}
}

func (b *Bindings) BindValue[T any](keyword string, target *T, parser Parser[T], options ...Option) {
	b.entries[keyword] = entry{
		setter: func(arg string) error {
			val, err := parser(arg)
			if err != nil {
				return fmt.Errorf("%s: %w", keyword, err)
			}

			*target = val

			return nil
		},
	}

	for _, option := range options {
		option(b, keyword)
	}
}

func (b *Bindings) BindSlice[T any](keyword string, target *[]T, parser Parser[T], options ...Option) {
	b.entries[keyword] = entry{
		setter: func(arg string) error {
			val, err := parser(arg)
			if err != nil {
				return fmt.Errorf("%s: %w", keyword, err)
			}

			*target = append(*target, val)

			return nil
		},
		repeatable: true,
	}

	for _, option := range options {
		option(b, keyword)
	}
}

func (b *Bindings) BindNode(keyword string, target Binder, options ...Option) {
	sub := New()
	target.Bind(sub)

	b.entries[keyword] = entry{
		sub:    sub,
		binder: target,
	}

	for _, option := range options {
		option(b, keyword)
	}
}

func ParseString(s string) (string, error) {
	return s, nil
}

func ParseInt[T SignedInteger](s string) (T, error) {
	bits := int(unsafe.Sizeof(*new(T))) * 8

	v, err := strconv.ParseInt(s, 10, bits)

	return T(v), err
}

func ParseUint[T UnsignedInteger](s string) (T, error) {
	bits := int(unsafe.Sizeof(*new(T))) * 8

	v, err := strconv.ParseUint(s, 10, bits)

	return T(v), err
}

func ParseFloat[T Float](s string) (T, error) {
	bits := int(unsafe.Sizeof(*new(T))) * 8

	v, err := strconv.ParseFloat(s, bits)

	return T(v), err
}

func ParseDuration(s string) (time.Duration, error) {
	d, err := time.ParseDuration(s)
	return d, err
}

func ParseTime(layout string) func(string) (time.Time, error) {
	return func(s string) (time.Time, error) {
		t, err := time.Parse(layout, s)
		return t, err
	}
}

func ParsePath(s string) (path string, err error) {
	if s == "" {
		return "", &fs.PathError{
			Op:   "parse",
			Path: s,
			Err:  fs.ErrInvalid,
		}
	}

	return filepath.Clean(s), nil
}

func ParseFileMode(s string) (os.FileMode, error) {
	v, err := strconv.ParseUint(s, 8, 32)
	if err != nil {
		return 0, err
	}

	return os.FileMode(v), nil
}

func ParseText[T any, PT TextUnmarshaler[T]](s string) (T, error) {
	var v T
	err := PT(&v).UnmarshalText([]byte(s))
	if err != nil {
		return v, err
	}

	return v, nil
}

func ParseURL[T URL](s string) (T, error) {
	var v T

	u, err := url.Parse(s)
	if err != nil {
		return v, err
	}

	switch p := any(&v).(type) {
	case *url.URL:
		*p = *u
	case **url.URL:
		*p = u
	}

	return v, nil
}

func SetBool(v bool) func(_ string) (bool, error) {
	return func(_ string) (bool, error) {
		return v, nil
	}
}

func (b *Bindings) BindString(keyword string, target *string, options ...Option) {
	b.BindValue(keyword, target, ParseString, options...)
}

func (b *Bindings) BindInt[T SignedInteger](keyword string, target *T, options ...Option) {
	b.BindValue(keyword, target, ParseInt, options...)
}

func (b *Bindings) BindUint[T UnsignedInteger](keyword string, target *T, options ...Option) {
	b.BindValue(keyword, target, ParseUint, options...)
}

func (b *Bindings) BindFloat[T Float](keyword string, target *T, options ...Option) {
	b.BindValue(keyword, target, ParseFloat, options...)
}

// BindBool will set the target bool to `true` when the keyword is found.
func (b *Bindings) BindBool(keyword string, target *bool, options ...Option) {
	b.BindValue(keyword, target, SetBool(true), append(options, NoArgument())...)
}

// BindBoolNeg will set the target bool to `false` when the keyword is found.
func (b *Bindings) BindBoolNeg(keyword string, target *bool, options ...Option) {
	b.BindValue(keyword, target, SetBool(false), append(options, NoArgument())...)
}

func (b *Bindings) BindDuration(keyword string, target *time.Duration, options ...Option) {
	b.BindValue(keyword, target, ParseDuration, options...)
}

func (b *Bindings) BindAddr(keyword string, target *netip.Addr, options ...Option) {
	b.BindText(keyword, target, options...)
}

func (b *Bindings) BindAddrPort(keyword string, target *netip.AddrPort, options ...Option) {
	b.BindText(keyword, target, options...)
}

func (b *Bindings) BindNetPrefix(keyword string, target *netip.Prefix, options ...Option) {
	b.BindText(keyword, target, options...)
}

func (b *Bindings) BindTime(keyword string, target *time.Time, layout string, options ...Option) {
	b.BindValue(keyword, target, ParseTime(layout), options...)
}

func (b *Bindings) BindPath(keyword string, target *string, options ...Option) {
	b.BindValue(keyword, target, ParsePath, options...)
}

func (b *Bindings) BindFileMode(keyword string, target *os.FileMode, options ...Option) {
	b.BindValue(keyword, target, ParseFileMode, options...)
}

func (b *Bindings) BindText[T any, PT TextUnmarshaler[T]](keyword string, target PT, options ...Option) {
	b.entries[keyword] = entry{
		setter: func(arg string) error {
			err := target.UnmarshalText([]byte(arg))
			if err != nil {
				return fmt.Errorf("%s: %w", keyword, err)
			}

			return nil
		},
	}

	for _, option := range options {
		option(b, keyword)
	}
}

func (b *Bindings) BindLogLevel(keyword string, target *slog.Level, options ...Option) {
	b.BindText(keyword, target, options...)
}

func (b *Bindings) BindLogLevelVar(keyword string, target *slog.LevelVar, options ...Option) {
	b.BindText(keyword, target, options...)
}

func (b *Bindings) BindURL[T URLRef](keyword string, target T, options ...Option) {
	b.entries[keyword] = entry{
		setter: func(arg string) error {
			var err error

			switch p := any(target).(type) {
			case *url.URL:
				err = p.UnmarshalBinary([]byte(arg))
			case **url.URL:
				if *p == nil {
					*p = new(url.URL)
				}

				err = (*p).UnmarshalBinary([]byte(arg))
			}

			if err != nil {
				return fmt.Errorf("%s: %w", keyword, err)
			}

			return nil
		},
	}

	for _, option := range options {
		option(b, keyword)
	}
}

func (b *Bindings) BindStringSlice(keyword string, target *[]string, options ...Option) {
	b.BindSlice(keyword, target, ParseString, options...)
}

func (b *Bindings) BindIntSlice[T SignedInteger](keyword string, target *[]T, options ...Option) {
	b.BindSlice(keyword, target, ParseInt, options...)
}

func (b *Bindings) BindUintSlice[T UnsignedInteger](keyword string, target *[]T, options ...Option) {
	b.BindSlice(keyword, target, ParseUint, options...)
}

func (b *Bindings) BindFloatSlice[T Float](keyword string, target *[]T, options ...Option) {
	b.BindSlice(keyword, target, ParseFloat, options...)
}

func (b *Bindings) BindDurationSlice(keyword string, target *[]time.Duration, options ...Option) {
	b.BindSlice(keyword, target, ParseDuration, options...)
}

func (b *Bindings) BindAddrSlice(keyword string, target *[]netip.Addr, options ...Option) {
	b.BindTextSlice(keyword, target, options...)
}

func (b *Bindings) BindAddrPortSlice(keyword string, target *[]netip.AddrPort, options ...Option) {
	b.BindTextSlice(keyword, target, options...)
}

func (b *Bindings) BindNetPrefixSlice(keyword string, target *[]netip.Prefix, options ...Option) {
	b.BindTextSlice(keyword, target, options...)
}

func (b *Bindings) BindTimeSlice(keyword string, layout string, target *[]time.Time, options ...Option) {
	b.BindSlice(keyword, target, ParseTime(layout), options...)
}

func (b *Bindings) BindPathSlice(keyword string, target *[]string, options ...Option) {
	b.BindSlice(keyword, target, ParsePath, options...)
}

func (b *Bindings) BindFileModeSlice(keyword string, target *[]os.FileMode, options ...Option) {
	b.BindSlice(keyword, target, ParseFileMode, options...)
}

func (b *Bindings) BindTextSlice[T any, PT TextUnmarshaler[T]](keyword string, target *[]T, options ...Option) {
	b.BindSlice(keyword, target, ParseText[T, PT], options...)
}

func (b *Bindings) BindURLSlice[T URL](keyword string, target *[]T, options ...Option) {
	b.BindSlice(keyword, target, ParseURL[T], options...)
}

func (b *Bindings) Parse(args []string, callback func(string, Binder)) ([]string, error) {
	clear(b.seen)

	var unknownErr *UnknownKeywordError

	for len(args) > 0 {
		keyword := args[0]
		if keyword == "--" {
			args = args[1:]
			break
		}

		var ent entry
		var err error

		var match string

		match, ent, err = b.lookup(keyword)
		if err != nil {
			if _, ok := errors.AsType[*UnknownKeywordError](err); ok {
				unknownErr = &UnknownKeywordError{Keyword: keyword}

				break
			}

			return nil, err
		}

		keyword = match

		if _, seen := b.seen[keyword]; seen && !ent.repeatable {
			break
		}

		b.seen[keyword] = struct{}{}
		args = args[1:]

		if ent.sub != nil {
			args, err = ent.sub.Parse(args, callback)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", keyword, err)
			}

			if callback != nil {
				callback(keyword, ent.binder)
			}

			continue
		}

		if ent.noArg {
			if err = ent.setter(""); err != nil {
				return nil, err
			}

			continue
		}

		if len(args) == 0 {
			return nil, &MissingValueError{
				Keyword: keyword,
			}
		}

		err = ent.setter(args[0])
		if err != nil {
			return nil, err
		}

		args = args[1:]
	}

	var errs error

	for _, required := range b.required {
		if _, ok := b.seen[required]; !ok {
			errs = errors.Join(errs, &MissingKeywordError{
				Keyword: required,
			})
		}
	}

	if errs != nil {
		if unknownErr != nil {
			return nil, errors.Join(errs, unknownErr)
		}

		return nil, errs
	}

	return args, nil
}

func (b *Bindings) Has(keyword string) bool {
	_, ok := b.entries[keyword]

	return ok
}

func (b *Bindings) All() iter.Seq2[string, Info] {
	return func(yield func(string, Info) bool) {
		for _, keyword := range slices.Sorted(maps.Keys(b.entries)) {
			ent := b.entries[keyword]
			info := Info{
				Description: b.descriptions[keyword],
				Sub:         ent.sub,
				Repeatable:  ent.repeatable,
				Required:    slices.Contains(b.required, keyword),
				NoArg:       ent.noArg,
			}

			if !yield(keyword, info) {
				return
			}
		}
	}
}

func (b *Bindings) lookup(keyword string) (string, entry, error) {
	if ent, ok := b.entries[keyword]; ok {
		return keyword, ent, nil
	}

	var matchedEntry entry
	var matchedName string
	var matches int

	for name, ent := range b.entries {
		if strings.HasPrefix(name, keyword) {
			matchedEntry = ent
			matchedName = name
			matches++
		}
	}

	switch matches {
	case 0:
		return "", entry{}, &UnknownKeywordError{
			Keyword: keyword,
		}
	case 1:
		return matchedName, matchedEntry, nil
	}

	err := &AmbiguousKeywordError{
		Keyword: keyword,
		Matches: make([]string, 0, matches),
	}

	for name := range b.entries {
		if strings.HasPrefix(name, keyword) {
			err.Matches = append(err.Matches, name)
		}
	}

	slices.Sort(err.Matches)

	return "", entry{}, err
}
