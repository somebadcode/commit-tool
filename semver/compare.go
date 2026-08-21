/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package semver

func Compare(v1, v2 Version) int {
	if v1.Major < v2.Major {
		return -1
	}

	if v1.Major > v2.Major {
		return 1
	}

	if v1.Minor < v2.Minor {
		return -1
	}

	if v1.Minor > v2.Minor {
		return 1
	}

	if v1.Patch < v2.Patch {
		return -1
	}

	if v1.Patch > v2.Patch {
		return 1
	}

	p1 := len(v1.Prerelease) > 0
	p2 := len(v2.Prerelease) > 0

	if p1 != p2 {
		if p1 {
			return -1
		}
		return 1
	}

	if !p1 {
		return 0
	}

	n := min(len(v1.Prerelease), len(v2.Prerelease))

	for i := range n {
		a, b := v1.Prerelease[i], v2.Prerelease[i]

		// If both are numeric, do a simple comparison.
		if a.isNumeric && b.isNumeric {
			if a.numeric < b.numeric {
				return -1
			}

			if a.numeric > b.numeric {
				return 1
			}

			continue
		}

		if a.isNumeric && !b.isNumeric {
			return -1
		}

		if !a.isNumeric && b.isNumeric {
			return 1
		}

		if a.text != b.text {
			if a.text < b.text {
				return -1
			}

			return 1
		}
	}

	if len(v1.Prerelease) < len(v2.Prerelease) {
		return -1
	}
	if len(v1.Prerelease) > len(v2.Prerelease) {
		return 1
	}

	return 0
}
