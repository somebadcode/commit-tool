/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package revtype

import (
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
)

type RevisionType int

const (
	TypeUnknown RevisionType = iota
	TypeBranch
	TypeTag
	TypeHash
	TypeRelative
	TypeRemote
)

func TypeOf(rev plumbing.Revision) RevisionType {
	s := rev.String()

	switch {
	case strings.HasPrefix(s, "refs/heads/"):
		return TypeBranch
	case strings.HasPrefix(s, "refs/tags/"):
		return TypeTag
	case strings.HasPrefix(s, "refs/remotes/"):
		return TypeRemote
	case strings.HasPrefix(s, "refs/"):
		return TypeUnknown
	case strings.ContainsAny(s, "~^:"):
		return TypeRelative
	case isHash(s):
		return TypeHash
	default:
		return TypeBranch
	}
}

func isHash(s string) bool {
	if len(s) < 7 || len(s) > 64 {
		return false
	}

	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}

	return true
}
