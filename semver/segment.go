/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package semver

import (
	"strconv"
	"strings"
)

type Segment struct {
	text      string
	numeric   uint64
	isNumeric bool
}

func NewSegment(text string) Segment {
	return Segment{
		text: text,
	}
}

func NewNumericSegment(numeric uint64) Segment {
	return Segment{
		text:      strconv.FormatUint(numeric, 10),
		numeric:   numeric,
		isNumeric: true,
	}
}

func (s *Segment) String() string {
	return s.text
}

func (s *Segment) IsNumeric() bool {
	return s.isNumeric
}

func (s *Segment) Uint64() uint64 {
	return s.numeric
}

func JoinSegments(segments []Segment, sep string) string {
	if len(segments) == 0 {
		return ""
	}
	if len(segments) == 1 {
		return segments[0].String()
	}

	var b strings.Builder

	b.Grow(len(segments) * (20 + len(sep)))

	b.WriteString(segments[0].String())
	for i := 1; i < len(segments); i++ {
		b.WriteString(sep)
		b.WriteString(segments[i].String())
	}
	return b.String()
}
