/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package cli_test

import (
	"context"
	"fmt"

	"codeberg.org/somebadcode/commit-tool/cli"
	"codeberg.org/somebadcode/commit-tool/cli/bindings"
)

type myCommand struct {
	message string
}

func (c *myCommand) Bind(bindings *bindings.Bindings) {
	bindings.BindString("msg", &c.message)
}

func (c *myCommand) Execute(_ context.Context, cli *cli.CLI) error {
	_, _ = fmt.Fprintln(cli.Stdout(), c.message)
	return nil
}

func Example() {
	app := cli.New()

	_ = app.Register("run", &myCommand{})

	_ = app.Parse([]string{"run", "msg", "Hello, World!"})

	_ = app.Execute(context.TODO())

	// Output:
	// Hello, World!
}
