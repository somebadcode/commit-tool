/*
 * Copyright 2025 Tobias Dahlberg
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package linter_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/somebadcode/commit-tool/commitlinter"
	"github.com/somebadcode/commit-tool/internal/repobuilder"
	"github.com/somebadcode/commit-tool/linter"
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
		Rev          plumbing.Revision
		OtherRev     plumbing.Revision
		ReportFunc   linter.ReportFunc
		StopFunc     linter.StopFunc
		CommitLinter linter.CommitLinter
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
				CommitLinter: &commitlinter.Linter{
					Filters: commitlinter.Filters{
						commitlinter.FilterInitialCommit,
					},
				},
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
				CommitLinter: &commitlinter.Linter{},
				ReportFunc:   linter.SlogReporter(slog.New(slog.DiscardHandler)),
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
				CommitLinter: &commitlinter.Linter{},
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
				CommitLinter: &commitlinter.Linter{},
				StopFunc:     linter.StopAfterN(1),
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
				CommitLinter: &commitlinter.Linter{},
				OtherRev:     "refs/heads/main",
			},
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
				OtherRev:     tt.fields.OtherRev,
				ReportFunc:   tt.fields.ReportFunc,
				StopFunc:     tt.fields.StopFunc,
				CommitLinter: tt.fields.CommitLinter,
			}

			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			if err = l.Run(ctx); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)

				var lintError linter.LintError

				if errors.As(err, &lintError) {
					t.Errorf("Run() error = %v", lintError)
				}
			}
		})
	}
}
