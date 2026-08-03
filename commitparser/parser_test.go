/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package commitparser

import (
	"reflect"
	"runtime/debug"
	"testing"

	gocmp "github.com/google/go-cmp/cmp"
)

type argsParse struct {
	message string
}

type testParseCase struct {
	name    string
	args    argsParse
	want    CommitMessage
	wantErr bool
}

var testParseCases = []testParseCase{
	{
		name: "type scope subject and body",
		args: argsParse{
			message: "feat(woop): something\n\nAdded more features\n",
		},
		want: CommitMessage{
			Type:     "feat",
			Scope:    "woop",
			Subject:  "something",
			Body:     "Added more features",
			Trailers: nil,
		},
	},
	{
		name: "type subject and body",
		args: argsParse{
			message: "feat: something\n\nAdded more features\n",
		},
		want: CommitMessage{
			Type:     "feat",
			Subject:  "something",
			Body:     "Added more features",
			Trailers: nil,
		},
	},
	{
		name: "type and subject",
		args: argsParse{
			message: "change: stuff",
		},
		want: CommitMessage{
			Type:     "change",
			Subject:  "stuff",
			Trailers: nil,
		},
	},
	{
		name: "type but no subject",
		args: argsParse{
			message: "change:",
		},
		wantErr: true,
	},
	{
		name: "only type",
		args: argsParse{
			message: "change",
		},
		wantErr: true,
	},
	{
		name: "breaking feature",
		args: argsParse{
			message: "feat!: refactored to support Y\n\nDid stuff!",
		},
		want: CommitMessage{
			Type:     "feat",
			Subject:  "refactored to support Y",
			Breaking: true,
			Body:     "Did stuff!",
		},
	},
	{
		name: "breaking feature with scope",
		args: argsParse{
			message: "feat(cli)!: refactored to support Y\n\nDid stuff!",
		},
		want: CommitMessage{
			Type:     "feat",
			Subject:  "refactored to support Y",
			Scope:    "cli",
			Breaking: true,
			Body:     "Did stuff!",
		},
	},
	{
		name: "misplaced exclamation mark",
		args: argsParse{
			message: "feat!(cli): refactored to support Y\n\nDid stuff!",
		},
		want: CommitMessage{
			Type:     "feat",
			Breaking: true,
		},
		wantErr: true,
	},
	{
		name: "empty",
		args: argsParse{
			message: "",
		},
		wantErr: true,
	},
	{
		name: "missing space",
		args: argsParse{
			message: "feat:new stuff",
		},
		want:    CommitMessage{Type: "feat"},
		wantErr: true,
	},
	{
		name: "missing colon",
		args: argsParse{
			message: "feat! new stuff",
		},
		want: CommitMessage{
			Type:     "feat",
			Breaking: true,
		},
		wantErr: true,
	},
	{
		name: "revert",
		args: argsParse{
			message: "revert: new stuff",
		},
		want: CommitMessage{
			Type:    "revert",
			Subject: "new stuff",
			Revert:  true,
		},
	},
	{
		name: "merge",
		args: argsParse{
			message: "merge: branch 'foo' into 'bar'",
		},
		want: CommitMessage{
			Type:    "merge",
			Subject: "branch 'foo' into 'bar'",
			Merge:   true,
		},
	},
	{
		name: "single_paragraph",
		args: argsParse{
			message: "feat: hej\n\nSomething fun",
		},
		want: CommitMessage{
			Type:    "feat",
			Subject: "hej",
			Body:    "Something fun",
		},
	},
	{
		name: "two_paragraphs",
		args: argsParse{
			message: "feat: hallo\n\nSomething fun\n\nSecond paragraph",
		},
		want: CommitMessage{
			Type:    "feat",
			Subject: "hallo",
			Body:    "Something fun\n\nSecond paragraph",
		},
	},
	{
		name: "trailer",
		args: argsParse{
			message: "feat: oi\n\nSomething fun\n\nSecond paragraph\n\nTicket: ABC-4321",
		},
		want: CommitMessage{
			Type:    "feat",
			Subject: "oi",
			Body:    "Something fun\n\nSecond paragraph",
			Trailers: []Trailer{
				{
					Key:   "Ticket",
					Value: "ABC-4321",
				},
			},
		},
	},
	{
		name: "trailers",
		args: argsParse{
			message: "feat: oi\n\nSomething fun\n\nSecond paragraph\n\nTicket: ABC-4321\nFix #12",
		},
		want: CommitMessage{
			Type:    "feat",
			Subject: "oi",
			Body:    "Something fun\n\nSecond paragraph",
			Trailers: []Trailer{
				{
					Key:   "Ticket",
					Value: "ABC-4321",
				}, {
					Key:   "Fix",
					Value: "#12",
				},
			},
		},
	},
	{
		name: "not_a_trailer",
		args: argsParse{
			message: "feat: oi\n\nSomething fun\n\nSecond paragraph\n\nTicket: ABC-4321\nMalformed commit",
		},
		want: CommitMessage{
			Type:    "feat",
			Subject: "oi",
			Body:    "Something fun\n\nSecond paragraph\n\nTicket: ABC-4321\nMalformed commit",
		},
	},
	{
		name: "not_a_trailer_alt",
		args: argsParse{
			message: "feat: oi\n\nSomething fun\n\nSecond paragraph\n\nFix #900\nMalformed commit",
		},
		want: CommitMessage{
			Type:    "feat",
			Subject: "oi",
			Body:    "Something fun\n\nSecond paragraph\n\nFix #900\nMalformed commit",
		},
	},
	{
		name: "bad_trailer_key",
		args: argsParse{
			message: "feat: oi\n\nSomething fun\n\nSecond paragraph\n\nReference to: abcdef1234\nMalformed commit",
		},
		want: CommitMessage{
			Type:    "feat",
			Subject: "oi",
			Body:    "Something fun\n\nSecond paragraph\n\nReference to: abcdef1234\nMalformed commit",
		},
	},
	{
		name: "more_bad_trailers3",
		args: argsParse{
			message: "feat: oi\n\nSomething fun\n\nSecond paragraph\n\nReference-to: ",
		},
		want: CommitMessage{
			Type:    "feat",
			Subject: "oi",
			Body:    "Something fun\n\nSecond paragraph\n\nReference-to:",
		},
	},
	{
		name: "multi_value_git_trailer",
		args: argsParse{
			message: "fix: foo parse bad\n\nFixed this and that\n\nTicket: ABC-100\nTicket: DEF-321",
		},
		want: CommitMessage{
			Type:    "fix",
			Subject: "foo parse bad",
			Body:    "Fixed this and that",
			Trailers: []Trailer{
				{
					Key:   "Ticket",
					Value: "ABC-100",
				}, {
					Key:   "Ticket",
					Value: "DEF-321",
				},
			},
		},
	},
	{
		name: "multi_line_git_trailer",
		args: argsParse{
			message: "fix: foo parse bad\n\nFixed this and that\n\nTicket: ABC-100\nSecurity: Addresses 1234 by this and that method\n and according to the discussion in incident SEC-34, blah boo blaha\n ",
		},
		want: CommitMessage{
			Type:    "fix",
			Subject: "foo parse bad",
			Body:    "Fixed this and that",
			Trailers: []Trailer{
				{
					Key:   "Ticket",
					Value: "ABC-100",
				},
				{
					Key:   "Security",
					Value: "Addresses 1234 by this and that method\n and according to the discussion in incident SEC-34, blah boo blaha",
				},
			},
		},
	},
	{
		name: "breaking_change",
		args: argsParse{
			message: "fix: trailer parsing\n\nFixed this and that\n\nTicket: ABC-100\nBREAKING CHANGE: Yup!",
		},
		want: CommitMessage{
			Type:     "fix",
			Subject:  "trailer parsing",
			Body:     "Fixed this and that",
			Breaking: true,
			Trailers: []Trailer{
				{
					Key:   "Ticket",
					Value: "ABC-100",
				}, {
					Key:   "BREAKING CHANGE",
					Value: "Yup!",
				},
			},
		},
	},
	{
		name: "breaking_stuff_bad",
		args: argsParse{
			message: "fix: trailer parsing\n\nFixed this and that\n\nBREAKING STUFF: Yup!",
		},
		want: CommitMessage{
			Type:    "fix",
			Subject: "trailer parsing",
			Body:    "Fixed this and that\n\nBREAKING STUFF: Yup!",
		},
	},
	{
		name: "trailer_not_upper",
		args: argsParse{
			message: "fix: trailer parsing\n\nFixed this and that\n\nthis-is-not-upper: nope!",
		},
		want: CommitMessage{
			Type:    "fix",
			Subject: "trailer parsing",
			Body:    "Fixed this and that",
			Trailers: []Trailer{
				{
					Key:   "this-is-not-upper",
					Value: "nope!",
				},
			},
		},
	},
	{
		name: "trailer_not_upper",
		args: argsParse{
			message: "fix: trailer parsing\n\nFixed this and that\n\nthis-is-not-upper: nope!\n",
		},
		want: CommitMessage{
			Type:    "fix",
			Subject: "trailer parsing",
			Body:    "Fixed this and that",
			Trailers: []Trailer{
				{
					Key:   "this-is-not-upper",
					Value: "nope!",
				},
			},
		},
	},
	{
		name: "trailer_bad_key",
		args: argsParse{
			message: "fix: trailer parsing\n\nFixed this and that\n\nSomething strange: Yup!",
		},
		want: CommitMessage{
			Type:    "fix",
			Subject: "trailer parsing",
			Body:    "Fixed this and that\n\nSomething strange: Yup!",
		},
	},
	{
		name: "bad_trailer_key_colon",
		args: argsParse{
			message: "feat: add foo\n\nkey:  ",
		},
		want: CommitMessage{
			Type:    "feat",
			Subject: "add foo",
			Body:    "key:",
		},
	},
	{
		name: "extra_linefeed",
		args: argsParse{
			message: "fix: trailer parsing\n\nFixed this and that\n\n\n",
		},
		want: CommitMessage{
			Type:    "fix",
			Subject: "trailer parsing",
			Body:    "Fixed this and that",
		},
	},
	{
		name: "trailer_key_carriage_return",
		args: argsParse{
			message: "fix: trailer parsing\n\nFixed this and that\n\nBREAKING\rCHANGE: Yup!",
		},
		wantErr: true,
	},
	{
		name: "trailer_value_carriage_return",
		args: argsParse{
			message: "fix: trailer parsing\n\nFixed this and that\n\nBREAKING CHANGE: Yup!\r and more",
		},
		wantErr: true,
	},
}

func TestParse(t *testing.T) {
	for _, tt := range testParseCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Parse(tt.args.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
				t.Errorf("parse() error = %v", err)
				return
			}

			if tt.wantErr {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				diff := gocmp.Diff(tt.want, got)
				t.Errorf("parse() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func BenchmarkParse(b *testing.B) {
	for _, tt := range testParseCases {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				_, _ = Parse(tt.args.message)
			}
		})
	}
}

func FuzzParse(f *testing.F) {
	f.Add("feat(parser)!: added support for trailers\n\nAdded support for Git trailers.")
	f.Add("feat(woop): something\n\nAdded more features\n")
	f.Add("feat: something\n\nAdded more features\n")
	f.Add("change: stuff")
	f.Add("feat!: refactored to support Y\n\nDid stuff!")

	for _, tt := range testParseCases {
		f.Add(tt.args.message)
	}

	f.Fuzz(func(t *testing.T, message string) {
		var msg CommitMessage
		defer func() {
			if err := recover(); err != nil {
				t.Errorf("Parser(%q) caused a panic (%#v): %v\n%s", message, msg, err, debug.Stack())
			}
		}()
		var err error

		msg, err = Parse(message)
		if err != nil {
			return
		}

		if msg.Type == "" {
			t.Errorf("Parser() did not return a type. Input: %q", message)
		}

		if msg.Subject == "" {
			t.Errorf("Parser() did not return a subject. Input: %q", message)
		}

		for _, tr := range msg.Trailers {
			if tr.Key == "" || tr.Value == "" {
				t.Errorf("Parse() yielded a trailer with an empty key(%q) or value(%q). Input: %q", tr.Key, tr.Value, message)
			}
		}
	})
}
