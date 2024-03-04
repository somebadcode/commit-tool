package linter_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/somebadcode/commit-tool/internal/repobuilder"
	"github.com/somebadcode/commit-tool/linter"
	"github.com/somebadcode/commit-tool/traverser"
)

func TestLinter_Lint(t *testing.T) {
	type fields struct {
		Rev        plumbing.Revision
		ReportFunc traverser.ReportFunc
		StopFunc   traverser.StopFunc
		VisitFunc  traverser.VisitFunc
	}

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
		repoOps []repobuilder.OperationFunc
		fields  fields
		wantErr bool
	}{
		{
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("Initial commit", commitOpts),
				repobuilder.Commit("feat: add foo", commitOpts),
				repobuilder.CheckoutBranch("fix-foo"),
				repobuilder.Commit("fix(foo): avoid panic", commitOpts),
				repobuilder.Commit("chore(foo): tweaked comments", commitOpts),
				repobuilder.Commit("chore(foo): fixed formatting", commitOpts),
			},
			fields: fields{
				Rev: "HEAD",
				VisitFunc: linter.Linter{
					Filters: linter.Filters{
						linter.FilterInitialCommit,
					},
				}.Lint,
				StopFunc: func(commit *object.Commit) bool {
					return false
				},
			},
		},
		{
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("Initial commit", commitOpts),
				repobuilder.Commit("feat: add foo", commitOpts),
				repobuilder.CheckoutBranch("fix-foo"),
				repobuilder.Commit("fix(foo): avoid panic", commitOpts),
				repobuilder.Commit("chore(foo): tweaked comments", commitOpts),
				repobuilder.Commit("chore(foo): fixed formatting", commitOpts),
			},
			fields: fields{
				Rev: "HEAD",
				VisitFunc: linter.Linter{
					Filters: linter.Filters{
						linter.FilterInitialCommit,
					},
				}.Lint,
			},
		},
		{
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("Initial commit", commitOpts),
				repobuilder.Commit("add foo", commitOpts),
				repobuilder.CheckoutBranch("fix-foo"),
				repobuilder.Commit("fix(foo): avoid panic", commitOpts),
				repobuilder.Commit("tweaked comments", commitOpts),
				repobuilder.Commit("chore(foo): fixed formatting", commitOpts),
			},
			fields: fields{
				Rev: "HEAD",
				VisitFunc: linter.Linter{
					Filters: linter.Filters{
						linter.FilterInitialCommit,
					},
				}.Lint,
			},
			wantErr: true,
		},
		{
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("Initial commit", commitOpts),
				repobuilder.Commit("add foo", commitOpts),
			},
			fields: fields{
				Rev:       "HEAD",
				VisitFunc: linter.Linter{}.Lint,
			},
			wantErr: true,
		},
		{
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("Initial commit", commitOpts),
				repobuilder.Commit("woop: add foo", commitOpts),
			},
			fields: fields{
				Rev: "HEAD",
				VisitFunc: linter.Linter{
					Filters: linter.Filters{
						linter.FilterInitialCommit,
					},
				}.Lint,
			},
		},
		{
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("Revert \"feat: add foo\"", commitOpts),
				repobuilder.Commit("chore(foo): fixed formatting", commitOpts),
			},
			fields: fields{
				Rev:       "HEAD",
				VisitFunc: linter.Linter{}.Lint,
			},
		},
		{
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("Merge 'foo' into 'bar'", commitOpts),
				repobuilder.Commit("chore(foo): fixed formatting", commitOpts),
			},
			fields: fields{
				Rev:       "HEAD",
				VisitFunc: linter.Linter{}.Lint,
			},
		},
		{
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("improvement(bah): ", commitOpts),
				repobuilder.Commit("chore(foo): fixed formatting", commitOpts),
			},
			fields: fields{
				Rev: "HEAD",
				VisitFunc: linter.Linter{
					Rules: linter.Rules{
						linter.RuleConventionalCommit,
					},
				}.Lint,
			},
			wantErr: true,
		},
		{
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("improvement(bah): Apple", commitOpts),
				repobuilder.Commit("chore(foo): fixed formatting", commitOpts),
			},
			fields: fields{
				Rev: "HEAD",
				VisitFunc: linter.Linter{
					Rules: linter.Rules{
						linter.RuleConventionalCommit,
					},
				}.Lint,
			},
			wantErr: true,
		},
	}

	t.Parallel()

	for _, tc := range tests {
		tt := tc

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo, err := repobuilder.Build(tt.repoOps...)
			if err != nil {
				t.Errorf("failed to build repo: %v", err)
				return
			}

			l := &traverser.Traverser{
				Repo:       repo,
				Rev:        tt.fields.Rev,
				ReportFunc: tt.fields.ReportFunc,
				StopFunc:   tt.fields.StopFunc,
				VisitFunc:  tt.fields.VisitFunc,
			}

			if err = l.Run(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
