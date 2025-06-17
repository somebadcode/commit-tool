/*
 * This file is part of commit-tool which is released under EUPL 1.2.
 * See the file LICENSE in the repository root for full license details.
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package kongmappings

import (
	"fmt"
	"reflect"

	"github.com/alecthomas/kong"
	"github.com/go-git/go-git/v5"
)

func Repository() kong.MapperFunc {
	return func(ctx *kong.DecodeContext, target reflect.Value) error {
		var repoPath string

		if err := ctx.Scan.PopValueInto("path", &repoPath); err != nil {
			return fmt.Errorf("cannot decode repository path: %w", err)
		}

		repo, err := git.PlainOpen(repoPath)
		if err != nil {
			return fmt.Errorf("cannot open repository %q: %w", repoPath, err)
		}

		target.Set(reflect.ValueOf(repo))

		return nil
	}
}
