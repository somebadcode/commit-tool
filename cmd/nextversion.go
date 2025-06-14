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
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/somebadcode/commit-tool/nextversion"
)

type NextVersionCommand struct {
	Repository     *git.Repository   `kong:"arg,placeholder='path',default='.',help='repository to lint'"`
	Revision       plumbing.Revision `kong:"arg,name='revision',aliases='rev',optional,default='HEAD',placeholder='REVISION',help='revision to start at'"`
	Output         string            `kong:"arg,type='path',default='-',help='where to output the next version'"`
	VSuffix        bool              `kong:"default='true',negatable,help='output with v-suffix, i.e. v1.3.2'"`
	WithPrerelease string            `kong:"optional,help='add prerelease information to tag'"`
	WithMetadata   string            `kong:"optional,help='add metadata to tag'"`
}

func (cmd *NextVersionCommand) Run(ctx context.Context, l *slog.Logger) error {
	var f io.WriteCloser
	if cmd.Output == "-" {
		f = os.Stdout
	} else {
		var err error
		f, err = os.OpenFile(cmd.Output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
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

	return next.Run(ctx)
}
