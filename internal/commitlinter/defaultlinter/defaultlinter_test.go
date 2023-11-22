package defaultlinter_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/somebadcode/conventional-commits-tool/internal/commitlinter"
	"github.com/somebadcode/conventional-commits-tool/internal/commitlinter/defaultlinter"
	"github.com/somebadcode/conventional-commits-tool/internal/repobuilder"
)

func TestLinter_Lint(t *testing.T) {
	type fields struct {
		Rev        plumbing.Revision
		ReportFunc commitlinter.ReportFunc
		StopFunc   commitlinter.StopFunc
		Linter     commitlinter.Linter
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
				Rev:    "HEAD",
				Linter: defaultlinter.New(defaultlinter.AllowInitialCommit()),
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
				Rev:    "HEAD",
				Linter: defaultlinter.New(defaultlinter.AllowInitialCommit()),
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
				Rev:    "HEAD",
				Linter: defaultlinter.New(defaultlinter.AllowInitialCommit()),
			},
			wantErr: true,
		},
		{
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("Initial commit", commitOpts),
				repobuilder.Commit("add foo", commitOpts),
			},
			fields: fields{
				Rev:    "HEAD",
				Linter: defaultlinter.New(),
			},
			wantErr: true,
		},
		{
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("Initial commit", commitOpts),
				repobuilder.Commit("woop: add foo", commitOpts),
			},
			fields: fields{
				Rev:    "HEAD",
				Linter: defaultlinter.New(),
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

			l := &commitlinter.CommitLinter{
				Repo:       repo,
				Rev:        tt.fields.Rev,
				ReportFunc: tt.fields.ReportFunc,
				StopFunc:   tt.fields.StopFunc,
				Linter:     tt.fields.Linter,
			}

			if err = l.Run(context.TODO()); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
