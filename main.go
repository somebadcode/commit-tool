/*
 * This file is part of commit-tool which is released under EUPL 1.2.
 * See the file LICENSE in the repository root for full license details.
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package main

import (
	_ "embed"
	"os"

	"codeberg.org/somebadcode/commit-tool/cmd"
)

func main() {
	os.Exit(cmd.Run())
}
