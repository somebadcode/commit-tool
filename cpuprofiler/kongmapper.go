/*
 * This file is part of commit-tool which is released under EUPL 1.2.
 * See the file LICENSE in the repository root for full license details.
 *
 * SPDX-License-Identifier: EUPL-1.2
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
