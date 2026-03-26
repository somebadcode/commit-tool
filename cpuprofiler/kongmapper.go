/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package cpuprofiler

import (
	"fmt"
	"reflect"

	"github.com/alecthomas/kong"
)

func KongMapperFunc() kong.MapperFunc {
	return func(ctx *kong.DecodeContext, target reflect.Value) error {
		var filename string

		if err := ctx.Scan.PopValueInto("path", &filename); err != nil {
			return fmt.Errorf("cannot decode repository path: %w", err)
		}

		profiler := &Profiler{
			Filename: filename,
		}

		target.Set(reflect.ValueOf(profiler))

		return nil
	}
}
