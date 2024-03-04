package commitslogvalue_test

import (
	"log/slog"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/somebadcode/commit-tool/commitslogvalue"
)

func TestCommit(t *testing.T) {
	type args struct {
		commit *object.Commit
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{
				commit: &object.Commit{
					Hash: plumbing.Hash{},
					Author: object.Signature{
						Name:  "Gopher",
						Email: "gopher@example.com",
						When:  time.Date(2024, 02, 04, 13, 53, 00, 0, time.UTC),
					},
					Committer: object.Signature{
						Name:  "Gopher",
						Email: "gopher@example.com",
						When:  time.Date(2024, 02, 04, 13, 53, 00, 0, time.UTC),
					},
				},
			},
			want: `{"level":"DEBUG","msg":"","commit":{"hash":"0000000000000000000000000000000000000000","author":{"name":"Gopher","email":"gopher@example.com","when":"2024-02-04T13:53:00Z"}}}`,
		},
		{
			args: args{
				commit: &object.Commit{
					Hash: plumbing.Hash{},
					Author: object.Signature{
						Name:  "Gopher",
						Email: "gopher@example.com",
						When:  time.Date(2024, 02, 04, 13, 53, 00, 0, time.UTC),
					},
					Committer: object.Signature{
						Name:  "Capybara",
						Email: "capybara@example.com",
						When:  time.Date(2024, 02, 04, 14, 53, 00, 0, time.UTC),
					},
				},
			},
			want: `{"level":"DEBUG","msg":"","commit":{"hash":"0000000000000000000000000000000000000000","author":{"name":"Gopher","email":"gopher@example.com","when":"2024-02-04T13:53:00Z"},"committer":{"name":"Capybara","email":"capybara@example.com","when":"2024-02-04T14:53:00Z"}}}`,
		},
	}

	stripTimeRegexp := regexp.MustCompile(`"time":"[-[:digit:]:.+T]+",`)
	transformer := cmpopts.AcyclicTransformer("StripTime", func(v string) string {
		return stripTimeRegexp.ReplaceAllString(v, "")
	})

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf strings.Builder

			l := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
				AddSource:   false,
				Level:       slog.LevelDebug,
				ReplaceAttr: commitslogvalue.ReplaceAttr,
			}))

			l.Debug(tt.name, slog.Any("commit", tt.args.commit))

			if got := stripTimeRegexp.ReplaceAllString(strings.TrimSpace(buf.String()), ""); !cmp.Equal(got, tt.want, transformer) {
				t.Errorf("Commit() = got %s", got)
				t.Errorf("         =     %s", tt.want)
			}
		})
	}
}
