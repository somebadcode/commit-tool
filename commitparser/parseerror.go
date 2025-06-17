/*
 * This file is part of commit-tool which is released under EUPL 1.2.
 * See the file LICENSE in the repository root for full license details.
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package commitparser

import (
	"fmt"
)

type ParseError struct {
	err error
	Pos int
}

func (err ParseError) Error() string {
	return fmt.Sprintf("unexpected character at %d: %s", err.Pos, err.err)
}

func (err ParseError) Unwrap() error {
	return err.err
}
