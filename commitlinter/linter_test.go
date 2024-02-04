package commitlinter_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/somebadcode/commit-tool/commitlinter"
	"github.com/somebadcode/commit-tool/commitlinter/defaultlinter"
	"github.com/somebadcode/commit-tool/internal/repobuilder"
	"github.com/somebadcode/commit-tool/slogctx"
	"github.com/somebadcode/commit-tool/slognop"
)

func TestCommitLinter_Lint(t *testing.T) {
	commitOpts := git.CommitOptions{
		AllowEmptyCommits: true,
		Author: &object.Signature{
			Name:  "Gopher",
			Email: "gopher@example.com",
			When:  time.Date(2023, 2, 4, 23, 22, 0, 0, time.UTC),
		},
	}

	type fields struct {
		Rev        plumbing.Revision
		OtherRev   plumbing.Revision
		ReportFunc commitlinter.ReportFunc
		StopFunc   commitlinter.StopFunc
		Linter     commitlinter.Linter
	}

	tests := []struct {
		name    string
		repoOps []repobuilder.OperationFunc
		fields  fields
		wantErr bool
	}{
		{
			name: "simple_with_branch",
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("Initial commit", commitOpts),
				repobuilder.Commit("feat: add foo", commitOpts),
				repobuilder.CheckoutBranch("fix-foo"),
				repobuilder.Commit("fix(foo): avoid panic", commitOpts),
				repobuilder.Commit("chore(foo): tweaked comments", commitOpts),
				repobuilder.Commit("chore(foo): fixed formatting", commitOpts),
			},
			fields: fields{
				Linter: defaultlinter.New(defaultlinter.AllowInitialCommit()),
			},
		},
		{
			name: "bad_commit",
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("Initial commit", commitOpts),
				repobuilder.Commit("add foo", commitOpts),
				repobuilder.CheckoutBranch("fix-foo"),
				repobuilder.Commit("fix(foo): avoid panic", commitOpts),
				repobuilder.Commit("tweaked comments", commitOpts),
				repobuilder.Commit("chore(foo): fixed formatting", commitOpts),
			},
			fields: fields{
				Linter:     defaultlinter.New(),
				ReportFunc: commitlinter.SlogReporter(slognop.New()),
			},
			wantErr: true,
		},
		{
			name: "bad_commit_short",
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("Initial commit", commitOpts),
				repobuilder.Commit("add foo", commitOpts),
			},
			fields: fields{
				Linter: defaultlinter.New(),
			},
			wantErr: true,
		},
		{
			name: "StopAfterN(1)",
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("bad commit", commitOpts),
				repobuilder.Commit("fix: bug #1", commitOpts),
				repobuilder.Commit("feat: add foo", commitOpts),
			},
			fields: fields{
				Linter:   defaultlinter.New(),
				StopFunc: commitlinter.StopAfterN(1),
			},
		},
		{
			name: "OtherRev",
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("bad commit #1", commitOpts),
				repobuilder.Commit("bad commit #2", commitOpts),
				repobuilder.Commit("bad commit #3", commitOpts),
				repobuilder.CheckoutBranch("feature/xyz"),
				repobuilder.Commit("fix: bug #1", commitOpts),
				repobuilder.Commit("feat: add foo", commitOpts),
			},
			fields: fields{
				Linter:   defaultlinter.New(),
				OtherRev: "refs/heads/main",
			},
		},
	}

	t.Parallel()

	ctx := slogctx.Context(context.TODO(), slognop.New())

	for _, tc := range tests {
		tt := tc

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo, err := repobuilder.Build(tt.repoOps...)
			if err != nil {
				t.Errorf("failed to build repo: %v", err)
				return
			}

			l := &commitlinter.CommitLinter{
				Repo:       repo,
				Rev:        tt.fields.Rev,
				OtherRev:   tt.fields.OtherRev,
				ReportFunc: tt.fields.ReportFunc,
				StopFunc:   tt.fields.StopFunc,
				Linter:     tt.fields.Linter,
			}

			if err = l.Run(ctx); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)

				var lintError commitlinter.LintError

				if errors.As(err, &lintError) {
					t.Errorf("Run() error = %v", lintError)
				}
			}
		})
	}
}
