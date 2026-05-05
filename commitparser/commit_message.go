/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package commitparser

type CommitMessage struct {
	Type     string
	Scope    string
	Subject  string
	Body     string
	Trailers []Trailer
	Breaking bool
	Revert   bool
	Merge    bool
}

type Trailer struct {
	Key   string
	Value string
}
