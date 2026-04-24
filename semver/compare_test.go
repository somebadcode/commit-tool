/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package semver

import (
	"testing"
)

func TestCompare(t *testing.T) {
	tests := []struct {
		name string
		v1   Version
		v2   Version
		want int
	}{
		{
			v1:   Version{Major: 1, Minor: 2, Patch: 3},
			v2:   Version{Major: 1, Minor: 2, Patch: 3},
			want: 0,
		},
		{
			v1:   Version{Major: 1, Minor: 2, Patch: 3},
			v2:   Version{Major: 0, Minor: 3, Patch: 4},
			want: 1,
		},
		{
			v1:   Version{Major: 1, Minor: 2, Patch: 3},
			v2:   Version{Major: 1, Minor: 3, Patch: 0},
			want: -1,
		},
		{
			v1:   Version{Major: 1, Minor: 3, Patch: 0},
			v2:   Version{Major: 1, Minor: 0, Patch: 0},
			want: 1,
		},
		{
			v1:   Version{Major: 1, Minor: 2, Patch: 3},
			v2:   Version{Major: 2, Minor: 0, Patch: 3},
			want: -1,
		},
		{
			v1:   Version{Major: 1, Minor: 2, Patch: 4},
			v2:   Version{Major: 1, Minor: 2, Patch: 3},
			want: 1,
		},
		{
			v1:   Version{Major: 1, Minor: 2, Patch: 3},
			v2:   Version{Major: 1, Minor: 2, Patch: 6},
			want: -1,
		},
		{
			v1:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewNumericSegment(0)}},
			v2:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewNumericSegment(0)}},
			want: 0,
		},
		{
			v1:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewNumericSegment(1)}},
			v2:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewNumericSegment(0)}},
			want: 1,
		},
		{
			v1:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewNumericSegment(0)}},
			v2:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewNumericSegment(3)}},
			want: -1,
		},
		{
			v1:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewSegment("beta")}},
			v2:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewNumericSegment(1)}},
			want: 1,
		},
		{
			v1:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewNumericSegment(1)}},
			v2:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewSegment("beta")}},
			want: -1,
		},
		{
			v1:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewNumericSegment(1)}},
			v2:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewNumericSegment(1)}},
			want: 0,
		},
		{
			v1:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewSegment("alpha")}},
			v2:   Version{Major: 1, Minor: 2, Patch: 3},
			want: -1,
		},
		{
			v1:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewSegment("alpha")}},
			v2:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewSegment("alpha")}},
			want: 0,
		},
		{
			v1:   Version{Major: 1, Minor: 2, Patch: 3},
			v2:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewSegment("alpha")}},
			want: 1,
		},
		{
			v1:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewSegment("alpha")}},
			v2:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewSegment("alpha")}},
			want: 0,
		},
		{
			v1:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewSegment("alpha")}},
			v2:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewSegment("alpha")}},
			want: 0,
		},
		{
			v1:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewSegment("alpha")}},
			v2:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewSegment("beta")}},
			want: -1,
		},
		{
			v1:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewSegment("beta")}},
			v2:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewSegment("alpha")}},
			want: 1,
		},
		{
			v1:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewSegment("beta")}},
			v2:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewSegment("beta"), NewSegment("carrot")}},
			want: -1,
		},
		{
			v1:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewSegment("beta"), NewSegment("carrot")}},
			v2:   Version{Major: 1, Minor: 2, Patch: 3, Prerelease: []Segment{NewSegment("beta")}},
			want: 1,
		},
		{
			v1:   Version{Major: 1, Minor: 2, Patch: 3, BuildMetadata: []Segment{NewSegment("dev")}},
			v2:   Version{Major: 1, Minor: 2, Patch: 3},
			want: 0,
		},
		{
			v1:   Version{Major: 1, Minor: 2, Patch: 3},
			v2:   Version{Major: 1, Minor: 2, Patch: 3, BuildMetadata: []Segment{NewSegment("dev")}},
			want: 0,
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Compare(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}
