package commithistorywalker_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/somebadcode/conventional-commits-tool/internal/commithistorywalker"
	"github.com/somebadcode/conventional-commits-tool/internal/repobuilder"
)

func TestVisit(t *testing.T) {
	commitOpts := git.CommitOptions{
		AllowEmptyCommits: true,
		Author: &object.Signature{
			Name:  "Gopher",
			Email: "gopher@example.com",
			When:  time.Date(2023, 2, 4, 23, 22, 0, 0, time.UTC),
		},
	}

	tests := []struct {
		name    string
		ops     []repobuilder.OperationFunc
		want    []string
		wantErr bool
	}{
		{
			name: "initial",
			ops: []repobuilder.OperationFunc{
				repobuilder.Commit("Initial commit", commitOpts),
			},
			want: []string{
				"Initial commit",
			},
		},
		{
			name: "second",
			ops: []repobuilder.OperationFunc{
				repobuilder.Commit("Initial commit", commitOpts),
				repobuilder.Commit("feat: add foo", commitOpts),
				repobuilder.CheckoutBranch("fix-foo"),
				repobuilder.Commit("fix(foo): avoid panic", commitOpts),
				repobuilder.Commit("chore(foo): tweaked comments", commitOpts),
				repobuilder.Commit("chore(foo): fixed formatting", commitOpts),
			},
			want: []string{
				"chore(foo): fixed formatting",
				"chore(foo): tweaked comments",
				"fix(foo): avoid panic",
				"feat: add foo",
				"Initial commit",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := repobuilder.Build(tt.ops...)
			if err != nil {
				t.Errorf("failed to create test repository: %v", err)
				return
			}

			got := make([]string, 0, len(tt.ops))

			err = commithistorywalker.Walk(repo, func(commit *object.Commit) error {
				got = append(got, commit.Message)

				return nil
			})
			if err != nil {
				t.Errorf("Error while traversing ancestors: %v", err)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Walk() did not walk over expected commits, got %#v, want %#v", got, tt.want)
			}
		})
	}
}
