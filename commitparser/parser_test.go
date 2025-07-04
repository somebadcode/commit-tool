/*
 * This file is part of commit-tool which is released under EUPL 1.2.
 * See the file LICENSE in the repository root for full license details.
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package commitparser

import (
	"reflect"
	"runtime/debug"
	"testing"
)

type args struct {
	message string
}

type testCase struct {
	name    string
	args    args
	want    CommitMessage
	wantErr bool
}

var tests = []testCase{
	{
		name: "type scope subject and body",
		args: args{
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
		args: args{
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
		args: args{
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
		args: args{
			message: "change:",
		},
		wantErr: true,
	},
	{
		name: "only type",
		args: args{
			message: "change",
		},
		wantErr: true,
	},
	{
		name: "breaking feature",
		args: args{
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
		args: args{
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
		args: args{
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
		args: args{
			message: "",
		},
		wantErr: true,
	},
	{
		name: "missing space",
		args: args{
			message: "feat:new stuff",
		},
		want:    CommitMessage{Type: "feat"},
		wantErr: true,
	},
	{
		name: "missing colon",
		args: args{
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
		args: args{
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
		args: args{
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
		args: args{
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
		args: args{
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
		args: args{
			message: "feat: oi\n\nSomething fun\n\nSecond paragraph\n\nTicket: ABC-4321",
		},
		want: CommitMessage{
			Type:    "feat",
			Subject: "oi",
			Body:    "Something fun\n\nSecond paragraph",
			Trailers: map[string][]string{
				"Ticket": {"ABC-4321"},
			},
		},
	},
	{
		name: "trailers",
		args: args{
			message: "feat: oi\n\nSomething fun\n\nSecond paragraph\n\nTicket: ABC-4321\nFix #12",
		},
		want: CommitMessage{
			Type:    "feat",
			Subject: "oi",
			Body:    "Something fun\n\nSecond paragraph",
			Trailers: map[string][]string{
				"Ticket": {"ABC-4321"},
				"Fix":    {"#12"},
			},
		},
	},
	{
		name: "not_a_trailer",
		args: args{
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
		args: args{
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
		args: args{
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
		args: args{
			message: "feat: oi\n\nSomething fun\n\nSecond paragraph\n\nReference-to: ",
		},
		want: CommitMessage{
			Type:    "feat",
			Subject: "oi",
			Body:    "Something fun\n\nSecond paragraph\n\nReference-to: ",
		},
	},
	{
		name: "multi_value_git_trailer",
		args: args{
			message: "fix: foo parse bad\n\nFixed this and that\n\nTicket: ABC-100\nTicket: DEF-321",
		},
		want: CommitMessage{
			Type:    "fix",
			Subject: "foo parse bad",
			Body:    "Fixed this and that",
			Trailers: map[string][]string{
				"Ticket": {"ABC-100", "DEF-321"},
			},
		},
	},
	{
		name: "multi_line_git_trailer",
		args: args{
			message: "fix: foo parse bad\n\nFixed this and that\n\nTicket: ABC-100\nSecurity: Addresses 1234 by this and that method\n and according to the discussion in incident SEC-34, blah boo blaha\n ",
		},
		want: CommitMessage{
			Type:    "fix",
			Subject: "foo parse bad",
			Body:    "Fixed this and that",
			Trailers: map[string][]string{
				"Ticket":   {"ABC-100"},
				"Security": {"Addresses 1234 by this and that method\nand according to the discussion in incident SEC-34, blah boo blaha"},
			},
		},
	},
	{
		name: "breaking_change",
		args: args{
			message: "fix: trailer parsing\n\nFixed this and that\n\nTicket: ABC-100\nBREAKING CHANGE: Yup!",
		},
		want: CommitMessage{
			Type:     "fix",
			Subject:  "trailer parsing",
			Body:     "Fixed this and that",
			Breaking: true,
			Trailers: map[string][]string{
				"Ticket":          {"ABC-100"},
				"BREAKING CHANGE": {"Yup!"},
			},
		},
	},
	{
		name: "breaking_stuff_bad",
		args: args{
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
		args: args{
			message: "fix: trailer parsing\n\nFixed this and that\n\nthis-is-not-upper: nope!",
		},
		want: CommitMessage{
			Type:    "fix",
			Subject: "trailer parsing",
			Body:    "Fixed this and that\n\nthis-is-not-upper: nope!",
		},
	},
	{
		name: "trailer_bad_key",
		args: args{
			message: "fix: trailer parsing\n\nFixed this and that\n\nSomething strange: Yup!",
		},
		want: CommitMessage{
			Type:    "fix",
			Subject: "trailer parsing",
			Body:    "Fixed this and that\n\nSomething strange: Yup!",
		},
	},
}

func TestParse(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Parse(tt.args.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parse() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func BenchmarkParse(b *testing.B) {
	for _, tt := range tests {
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

	for _, tt := range tests {
		f.Add(tt.args.message)
	}

	f.Fuzz(func(t *testing.T, message string) {
		var msg CommitMessage
		defer func() {
			if err := recover(); err != nil {
				t.Errorf("Parser(%q) caused a panic (%#v): %v\n%s", message, msg, err, debug.Stack())
			}
		}()
		_, _ = Parse(message)
	})
}
