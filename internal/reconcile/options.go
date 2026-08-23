/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package reconcile

import (
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type Config struct {
	RemoteName      string
	SourceRevision  plumbing.Revision
	TargetRevisions []plumbing.Revision
}

var (
	ErrReconcileOnlyBranches = errors.New("automatic reconciliation only supports branches")
)

func (c *Config) Validate(repo *git.Repository) error {
	if c.RemoteName == "" {
		c.RemoteName = git.DefaultRemoteName
	}

	if c.SourceRevision == "" {
		c.SourceRevision = plumbing.Revision(plumbing.HEAD)
	}

	_, err := repo.ResolveRevision(c.SourceRevision)
	if err != nil {
		return fmt.Errorf("reconcile: resolve source revision(%s): %w", c.SourceRevision, err)
	}

	refName := expandRevision(c.SourceRevision)
	if !(refName.IsBranch() || refName.IsTag()) && refName != plumbing.HEAD {
		return fmt.Errorf("reconcile: source revision (%s) is not a branch: %w", c.SourceRevision, ErrReconcileOnlyBranches)
	}

	for _, target := range c.TargetRevisions {
		refName = expandRevision(target)
		if !(refName.IsBranch() || refName.IsTag()) {
			return fmt.Errorf("reconcile: target revision (%s) is not a branch: %w", target, ErrReconcileOnlyBranches)
		}
	}

	return nil
}
