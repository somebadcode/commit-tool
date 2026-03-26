/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
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
