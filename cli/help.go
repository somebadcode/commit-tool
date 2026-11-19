/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package cli

import (
	"context"
	"fmt"
	"io"
	"math"

	"codeberg.org/somebadcode/commit-tool/cli/bindings"
)

type help struct{}

func (h *help) Bind(_ *bindings.Bindings) {}

func (h *help) Priority() int {
	return math.MaxInt
}

func (h *help) Execute(_ context.Context, c *CLI) error {
	printBindings(c.Stdout(), c.bindings, "")

	return ErrStop
}

func printBindings(w io.Writer, b *bindings.Bindings, indent string) {
	first := true
	for keyword, info := range b.All() {
		if indent == "" && !first {
			_, _ = fmt.Fprintln(w)
		}

		first = false

		col := indent + keyword

		switch {
		case info.Description != "" && info.Required:
			_, _ = fmt.Fprintf(w, "%-20s %s (required)\n", col, info.Description)
		case info.Description != "":
			_, _ = fmt.Fprintf(w, "%-20s %s\n", col, info.Description)
		case info.Required:
			_, _ = fmt.Fprintf(w, "%-20s (required)\n", col)
		default:
			_, _ = fmt.Fprintf(w, "%s\n", col)
		}

		if info.Sub != nil {
			printBindings(w, info.Sub, indent+"  ")
		}
	}
}
