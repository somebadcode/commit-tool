/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package commitparser

import (
	"testing"
)

type argsCountTrailers struct {
	block string
}

type testCountTrailers struct {
	name string
	args argsCountTrailers
	want int
}

func testCountTrailersCases(tb testing.TB) []testCountTrailers {
	tb.Helper()

	return []testCountTrailers{
		{
			args: argsCountTrailers{
				block: "Key: value",
			},
			want: 1,
		},
		{
			args: argsCountTrailers{
				block: "Key value",
			},
			want: -1,
		},
		{
			args: argsCountTrailers{
				block: "Key #123",
			},
			want: 1,
		},
		{
			args: argsCountTrailers{
				block: "Key #123\nMore: stuff",
			},
			want: 2,
		},
		{
			args: argsCountTrailers{
				block: "Key #123\nMore: stuff\n on next line too\n and the next\n and next\n it just goes on\n and on",
			},
			want: 2,
		},
	}
}

func Test_counterTrailers(t *testing.T) {
	t.Parallel()

	for _, tt := range testCountTrailersCases(t) {
		t.Run(tt.name, func(t *testing.T) {
			if got := countTailers(tt.args.block); got != tt.want {
				t.Errorf("isTrailerBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Benchmark_countTrailers(b *testing.B) {
	for _, tt := range testCountTrailersCases(b) {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				_ = countTailers(tt.args.block)
			}
		})
	}
}

func Fuzz_countTrailers(f *testing.F) {
	for _, tt := range testCountTrailersCases(f) {
		f.Add(tt.args.block)
	}

	f.Fuzz(func(t *testing.T, block string) {
		_ = countTailers(block)
	})
}
