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

package cmd

import (
	"context"
	"log/slog"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/somebadcode/commit-tool/commitlinter"
	"github.com/somebadcode/commit-tool/linter"
)

type LintCommand struct {
	Repository    *git.Repository   `kong:"arg,placeholder='path',default='.',help='repository to lint'"`
	Revision      plumbing.Revision `kong:"name='revision',aliases='rev',optional,default='HEAD',placeholder='REVISION',help='revision to start at'"`
	OtherRevision plumbing.Revision `kong:"name='other-revision',aliases='other',optional,placeholder='REVISION',help='revision (actual other) to stop at (exclusive)'"`
}

func (cmd *LintCommand) Run(ctx context.Context, l *slog.Logger) error {
	lint := linter.Linter{
		Repo:       cmd.Repository,
		Rev:        cmd.Revision,
		OtherRev:   cmd.OtherRevision,
		ReportFunc: linter.SlogReporter(l),
		CommitLinter: &commitlinter.Linter{
			Rules: commitlinter.Rules{
				commitlinter.RuleConventionalCommit,
			},
		},
		Logger: l,
	}

	return lint.Run(ctx)
}
