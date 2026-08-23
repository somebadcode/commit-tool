/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package set

import (
	"iter"
)

type Set[E comparable] map[E]struct{}

func New[E comparable]() Set[E] {
	return make(Set[E])
}

func (s Set[E]) Add(e E) {
	s[e] = struct{}{}
}

func (s Set[E]) Items() iter.Seq[E] {
	return func(yield func(E) bool) {
		for e := range s {
			if !yield(e) {
				return
			}
		}
	}
}

func (s Set[E]) Len() int {
	return len(s)
}
