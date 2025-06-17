/*
 * This file is part of commit-tool which is released under EUPL 1.2.
 * See the file LICENSE in the repository root for full license details.
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package commitparser

type CommitMessage struct {
	Type     string
	Scope    string
	Subject  string
	Body     string
	Trailers map[string][]string
	Breaking bool
	Revert   bool
	Merge    bool
}
