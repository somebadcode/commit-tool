/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package cmd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"codeberg.org/somebadcode/commit-tool/commitlinter"
	"codeberg.org/somebadcode/commit-tool/commitlinter/conventionalcommits"
	"codeberg.org/somebadcode/commit-tool/linter"
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
			Validator: conventionalcommits.New(conventionalcommits.Settings{}),
		},
		Logger: l,
	}

	err := lint.Run(ctx)
	if err != nil {
		return fmt.Errorf("lint failed: %w", err)
	}

	return nil
}
