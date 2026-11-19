/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package bindings

type Option func(*Bindings, string)

func WithDescription(desc string) Option {
	return func(b *Bindings, keyword string) {
		b.descriptions[keyword] = desc
	}
}

func Required() Option {
	return func(b *Bindings, keyword string) {
		b.required = append(b.required, keyword)
	}
}

func NoArgument() Option {
	return func(b *Bindings, keyword string) {
		if ent, ok := b.entries[keyword]; ok {
			ent.noArg = true
			b.entries[keyword] = ent
		}
	}
}

//func Repeatable() Option {
//	return func(b *Bindings, keyword string) {
//		if ent, ok := b.entries[keyword]; ok {
//			ent.repeatable = true
//			b.entries[keyword] = ent
//		}
//	}
//}
