/*
 * This file is part of commit-tool which is released under EUPL 1.2.
 * See the file LICENSE in the repository root for full license details.
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package commitlinter_test

import (
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"codeberg.org/somebadcode/commit-tool/commitlinter"
	"codeberg.org/somebadcode/commit-tool/commitlinter/conventionalcommits"
	"codeberg.org/somebadcode/commit-tool/internal/repobuilder"
	"codeberg.org/somebadcode/commit-tool/linter"
)

func TestLinter_Lint(t *testing.T) {
	type fields struct {
		Rev        plumbing.Revision
		ReportFunc linter.ReportFunc
		StopFunc   linter.StopFunc
		Linter     linter.CommitLinter
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
				Linter: &commitlinter.Linter{
					Filters: commitlinter.Filters{
						commitlinter.FilterInitialCommit,
					},
				},
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
				Linter: &commitlinter.Linter{
					Filters: commitlinter.Filters{
						commitlinter.FilterInitialCommit,
					},
				},
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
				Linter: &commitlinter.Linter{
					Filters: commitlinter.Filters{
						commitlinter.FilterInitialCommit,
					},
				},
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
				Linter: &commitlinter.Linter{},
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
				Linter: &commitlinter.Linter{
					Filters: commitlinter.Filters{
						commitlinter.FilterInitialCommit,
					},
				},
			},
		},
		{
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("Revert \"feat: add foo\"", commitOpts),
				repobuilder.Commit("chore(foo): fixed formatting", commitOpts),
			},
			fields: fields{
				Rev:    "HEAD",
				Linter: &commitlinter.Linter{},
			},
		},
		{
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("Merge 'foo' into 'bar'", commitOpts),
				repobuilder.Commit("chore(foo): fixed formatting", commitOpts),
			},
			fields: fields{
				Rev:    "HEAD",
				Linter: &commitlinter.Linter{},
			},
			wantErr: true,
		},
		{
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("improvement(bah): ", commitOpts),
				repobuilder.Commit("chore(foo): fixed formatting", commitOpts),
			},
			fields: fields{
				Rev: "HEAD",
				Linter: &commitlinter.Linter{
					Rules: commitlinter.Rules{
						conventionalcommits.Verify,
					},
				},
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
				Linter: &commitlinter.Linter{
					Rules: commitlinter.Rules{
						conventionalcommits.Verify,
					},
				},
			},
			wantErr: true,
		},
		{
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("apple(worm): do not eat", commitOpts),
				repobuilder.Commit("chore(foo): fixed formatting", commitOpts),
			},
			fields: fields{
				Rev: "HEAD",
				Linter: &commitlinter.Linter{
					Rules: commitlinter.Rules{
						conventionalcommits.Verify,
					},
				},
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

			l := &linter.Linter{
				Repo:         repo,
				Rev:          tt.fields.Rev,
				ReportFunc:   tt.fields.ReportFunc,
				StopFunc:     tt.fields.StopFunc,
				CommitLinter: tt.fields.Linter,
			}

			if err = l.Run(t.Context()); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)

				for _, errParse := range err.(interface{ Unwrap() []error }).Unwrap() {
					t.Error(errParse)
				}
			}
		})
	}
}
