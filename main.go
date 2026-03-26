/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package main

import (
	"os"

	"codeberg.org/somebadcode/commit-tool/cmd"
)

func main() {
	os.Exit(cmd.Run())
}
