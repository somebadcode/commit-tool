/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package semver

import (
	"reflect"
	"testing"
)

func TestNewSegment(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name string
		args args
		want Segment
	}{
		{
			args: args{"999"},
			want: Segment{
				text: "999",
			},
		},
		{
			args: args{"abc"},
			want: Segment{
				text: "abc",
			},
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewSegment(tt.args.text)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSegment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewNumericSegment(t *testing.T) {
	type args struct {
		n uint64
	}

	tests := []struct {
		name string
		args args
		want Segment
	}{
		{
			args: args{999},
			want: Segment{
				text:      "999",
				numeric:   999,
				isNumeric: true,
			},
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewNumericSegment(tt.args.n)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNumericSegment() = %v, want %v", got, tt.want)
			}
		})
	}
}
