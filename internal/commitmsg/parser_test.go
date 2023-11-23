package commitmsg

import (
	"reflect"
	"runtime/debug"
	"testing"
)

func TestParse(t *testing.T) {
	type args struct {
		message string
	}

	tests := []struct {
		name    string
		args    args
		want    CommitMessage
		wantErr bool
	}{
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
			want: CommitMessage{
				Type: "change",
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
				message: "Revert \"feat: new stuff\"",
			},
			want: CommitMessage{
				Type:    "feat",
				Subject: "new stuff",
				Revert:  true,
			},
		},
		{
			name: "merge",
			args: args{
				message: "Merge branch 'foo' into 'bar'",
			},
			want: CommitMessage{
				Type:    "merge",
				Subject: "Merge branch 'foo' into 'bar'",
				Merge:   true,
			},
		},
	}

	t.Parallel()

	for _, tc := range tests {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Parse(tt.args.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parse() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Parse("feat(parser)!: added support for trailers\n\nAdded support for Git trailers.")
	}
}

func FuzzParse(f *testing.F) {
	f.Add("feat(parser)!: added support for trailers\n\nAdded support for Git trailers.")
	f.Add("feat(woop): something\n\nAdded more features\n")
	f.Add("feat: something\n\nAdded more features\n")
	f.Add("change: stuff")
	f.Add("feat!: refactored to support Y\n\nDid stuff!")

	f.Fuzz(func(t *testing.T, message string) {
		var commit CommitMessage
		defer func() {
			if err := recover(); err != nil {
				t.Errorf("Parser(%q) caused a panic (%#v): %v\n%s", message, commit, err, debug.Stack())
			}
		}()
		commit, _ = Parse(message)
	})
}
