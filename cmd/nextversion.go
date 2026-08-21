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
	"io"
	"log/slog"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"codeberg.org/somebadcode/commit-tool/nextversion"
)

type NextVersionCommand struct {
	Repository     *git.Repository   `kong:"arg,placeholder='path',default='.',help='repository to lint'"`
	Revision       plumbing.Revision `kong:"arg,name='revision',aliases='rev',optional,default='HEAD',placeholder='REVISION',help='revision to start at'"`
	Output         string            `kong:"arg,type='path',default='-',help='where to output the next version'"`
	VSuffix        bool              `kong:"default='false',negatable,help='output with v-suffix, i.e. v1.3.2'"`
	WithPrerelease []string          `kong:"optional,help='add prerelease information to tag'"`
	WithMetadata   []string          `kong:"optional,help='add metadata to tag'"`
}

func (cmd *NextVersionCommand) Run(ctx context.Context, l *slog.Logger) error {
	var f io.WriteCloser
	if cmd.Output == "-" {
		f = os.Stdout
	} else {
		var err error
		f, err = os.OpenFile(cmd.Output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
		if err != nil {
			return fmt.Errorf("opening output file: %w", err)
		}

		defer func() {
			_ = f.Close()
		}()
	}

	next := nextversion.NextVersion{
		Repository: cmd.Repository,
		Revision:   cmd.Revision,
		VSuffix:    cmd.VSuffix,
		Prerelease: cmd.WithPrerelease,
		Metadata:   cmd.WithMetadata,
		Writer:     f,
		Logger:     l,
	}

	err := next.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to run next version: %w", err)
	}

	return nil
}
