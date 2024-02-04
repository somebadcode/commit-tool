package commitslogvalue_test

import (
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/somebadcode/commit-tool/commitslogvalue"
)

func TestValue(t *testing.T) {
	type args struct {
		commit *object.Commit
	}
	tests := []struct {
		name string
		args args
		want slog.Value
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
			want: slog.GroupValue(
				slog.String("hash", "0000000000000000000000000000000000000000"),
				slog.Group("author",
					slog.String("name", "Gopher"),
					slog.String("email", "gopher@example.com"),
					slog.Time("when", time.Date(2024, 02, 04, 13, 53, 00, 0, time.UTC)),
				),
			),
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
			want: slog.GroupValue(
				slog.String("hash", "0000000000000000000000000000000000000000"),
				slog.Group("author",
					slog.String("name", "Gopher"),
					slog.String("email", "gopher@example.com"),
					slog.Time("when", time.Date(2024, 02, 04, 13, 53, 00, 0, time.UTC)),
				),
				slog.Group("author",
					slog.String("name", "Capybara"),
					slog.String("email", "capyubara@example.com"),
					slog.Time("when", time.Date(2024, 02, 04, 14, 53, 00, 0, time.UTC)),
				),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := commitslogvalue.Value(tt.args.commit); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Value() = %v, want %v", got, tt.want)
			}
		})
	}
}
