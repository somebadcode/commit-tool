/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package cli

import (
	"io"
	"strings"

	"codeberg.org/somebadcode/commit-tool/cli/bindings"
)

type Option func(*CLI)

func WithStdin(r io.Reader) Option {
	return func(cli *CLI) {
		if r == nil {
			r = strings.NewReader("")
		}
		cli.stdin = r
	}
}

func WithStdout(w io.Writer) Option {
	return func(cli *CLI) {
		if w == nil {
			w = io.Discard
		}

		cli.stdout = w
	}
}

func WithStderr(w io.Writer) Option {
	return func(cli *CLI) {
		if w == nil {
			w = io.Discard
		}

		cli.stderr = w
	}
}

func WithHelp() Option {
	return func(cli *CLI) {
		_ = cli.Register("help", &help{},
			bindings.WithDescription("Show available commands"),
		)
	}
}
