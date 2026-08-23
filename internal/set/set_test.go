/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package set

import (
	"slices"
	"testing"
)

func TestSet(t *testing.T) {
	type testCase[E comparable] struct {
		name string
		add  []E
		want []E
	}

	tests := []testCase[string]{
		{
			name: "empty",
			want: nil,
		},
		{
			name: "one",
			add:  []string{"hello"},
			want: []string{"hello"},
		},
		{
			name: "two",
			add:  []string{"hello", "world"},
			want: []string{"hello", "world"},
		},
		{
			name: "two",
			add:  []string{"world", "hello"},
			want: []string{"hello", "world"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss := New[string]()

			for _, e := range tt.add {
				ss.Add(e)
			}

			got := make([]string, 0, ss.Len())
			for e := range ss.Items() {
				got = append(got, e)
			}

			slices.Sort(got)
			slices.Sort(tt.want)

			if slices.Compare(got, tt.want) != 0 {
				t.Errorf("slices.Compare() = got %v, want %v", got, tt.want)
			}
		})
	}
}
