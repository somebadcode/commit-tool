/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package cli_test

import (
	"context"
	"io"
	"os"

	"codeberg.org/somebadcode/commit-tool/cli"
	"codeberg.org/somebadcode/commit-tool/cli/bindings"
)

type exampleNewHelp struct {
	Text     string
	Extra    string
	FileMode os.FileMode
	Verbose  bool
}

func (h *exampleNewHelp) Bind(b *bindings.Bindings) {
	b.BindString("set", &h.Text,
		bindings.WithDescription("setting the text"),
		bindings.Required(),
	)

	b.BindString("xtra", &h.Extra,
		bindings.WithDescription("extra data"),
	)

	b.BindFileMode("mode", &h.FileMode, bindings.Required())

	b.BindBool("verbose", &h.Verbose)
}

func (h *exampleNewHelp) Execute(_ context.Context, c *cli.CLI) error {
	_, _ = io.WriteString(c.Stdout(), h.Text)

	return nil
}

func ExampleWithHelp() {
	app := cli.New(cli.WithHelp())

	_ = app.Register("message", &exampleNewHelp{})

	_ = app.Parse([]string{"help"})

	_ = app.Execute(context.TODO())

	// Output:
	// help                 Show available commands
	//
	// message
	//   mode               (required)
	//   set                setting the text (required)
	//   verbose
	//   xtra               extra data
}
