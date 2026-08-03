/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package commitparser

import (
	"strings"
)

func countTailers(paragraph string) int {
	var count int

	for pos := 0; pos < len(paragraph); {
		eol := strings.IndexByte(paragraph[pos:], '\n')
		if eol == -1 {
			eol = len(paragraph) - pos
		}

		for pos+eol+1 < len(paragraph) {
			c := paragraph[pos+eol+1]
			if c != ' ' && c != '\t' {
				break
			}

			nextLF := strings.IndexByte(paragraph[pos+eol+1:], '\n')
			if nextLF == -1 {
				eol = len(paragraph) - pos
				break
			}

			eol += nextLF + 1
		}

		trailer := paragraph[pos : pos+eol]

		if pos+eol+1 < len(paragraph) {
			pos += eol + 1
		} else {
			pos = len(paragraph)
		}

		if strings.HasPrefix(trailer, TrailerKeyBreakingChange+": ") {
			count++
			continue
		}

		i := strings.IndexAny(trailer, ": ")
		if i == -1 {
			return -1
		}

		c := trailer[i]
		switch c {
		case ':':
			i += 1
			if i >= len(trailer) {
				return -1
			}

			if trailer[i] != ' ' {
				return -1
			}

		case ' ':
			i += 1
			if i >= len(trailer) {
				return -1
			}

			if trailer[i] != '#' {
				return -1
			}
		}

		if strings.TrimSpace(trailer[i:]) == "" {
			return -1
		}

		count++
	}

	return count
}
