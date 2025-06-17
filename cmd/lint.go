/*
 * This file is part of commit-tool which is released under EUPL 1.2.
 * See the file LICENSE in the repository root for full license details.
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package cmd

import (
	"context"
	"log/slog"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/somebadcode/commit-tool/commitlinter"
	"github.com/somebadcode/commit-tool/commitlinter/conventionalcommits"
	"github.com/somebadcode/commit-tool/linter"
)

type LintCommand struct {
	Repository    *git.Repository   `kong:"placeholder='path',default='.',help='repository to lint'"`
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
				conventionalcommits.Verify,
			},
		},
		Logger: l,
	}

	return lint.Run(ctx)
}
