package cmd

import (
	"errors"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func Test_replaceAttr(t *testing.T) {
	type args struct {
		in0  []string
		attr slog.Attr
	}
	tests := []struct {
		name string
		args args
		want slog.Attr
	}{
		{
			name: "commit",
			args: args{
				attr: slog.Any("commit", &object.Commit{
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
				}),
			},
			want: slog.Any("commit", slog.GroupValue(
				slog.String("hash", "0000000000000000000000000000000000000000"),
				slog.Group("author",
					slog.String("name", "Gopher"),
					slog.String("email", "gopher@example.com"),
					slog.Time("when", time.Date(2024, 02, 04, 13, 53, 00, 0, time.UTC)),
				),
			)),
		},
		{
			name: "stringer",
			args: args{
				attr: slog.Any("err",
					errors.New("something")),
			},
			want: slog.Any("err", errors.New("something")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := replaceAttr(tt.args.in0, tt.args.attr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("replaceAttr() = %v, want %v", got, tt.want)
			}
		})
	}
}
