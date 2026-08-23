/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package reconcile

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"codeberg.org/somebadcode/commit-tool/internal/set"
)

const depthIncrementalFactor = 4
const maxDepth = 1000

var (
	ErrCannotReconcile     = errors.New("cannot reconcile revisions")
	ErrUnsupportedRevision = errors.New("unsupported revision")
)

func Reconcile(ctx context.Context, repo *git.Repository, c Config) ([]*object.Commit, error) {
	if err := c.Validate(repo); err != nil {
		return nil, err
	}

	remote, err := repo.Remote(c.RemoteName)
	if err != nil {
		return nil, fmt.Errorf("reconcile: bad remote %q: %w", c.RemoteName, err)
	}

	names := set.New[plumbing.ReferenceName]()

	ref := expandRevision(c.SourceRevision)
	if ref == plumbing.HEAD {
		var headRef *plumbing.Reference
		if headRef, err = repo.Head(); err != nil {
			return nil, fmt.Errorf("reconcile: HEAD: %w", err)
		}

		names.Add(headRef.Name())

	} else if ref.IsBranch() || ref.IsTag() {
		names.Add(ref)
	} else {
		return nil, fmt.Errorf("reconcile: invalid reference %q: %w", ref, ErrUnsupportedRevision)
	}

	for _, targetRev := range c.TargetRevisions {
		name := expandRevision(targetRev)
		if name.IsBranch() || name.IsTag() {
			names.Add(name)
		}
	}

	remoteName := remote.Config().Name
	var remoteURL string
	if len(remote.Config().URLs) > 0 {
		remoteURL = remote.Config().URLs[0]
	}

	refSpecs := make([]config.RefSpec, 0, len(c.TargetRevisions))

	for name := range names.Items() {
		if name.IsTag() {
			refSpecs = append(refSpecs, config.RefSpec(fmt.Sprintf("+%s:%s",
				name,
				name,
			)))

			continue
		}
		refSpecs = append(refSpecs, config.RefSpec(fmt.Sprintf("+%s:%s",
			name,
			plumbing.NewRemoteReferenceName(remoteName, name.Short()),
		)))
	}

	fetchOpts := &git.FetchOptions{
		RemoteName: remoteName,
		RemoteURL:  remoteURL,
		RefSpecs:   refSpecs,
		Depth:      10,
		Tags:       git.TagFollowing,
	}

	var sourceHash *plumbing.Hash

	mergeBases := make([]*object.Commit, 0)

	var resolved bool

	for {
		if fetchOpts.Depth >= maxDepth {
			return nil, fmt.Errorf("reconcile: max depth reached (%d>=%d): %w", fetchOpts.Depth, maxDepth, ErrCannotReconcile)
		}

		fetchErr := remote.FetchContext(ctx, fetchOpts)
		if fetchErr != nil && !errors.Is(fetchErr, git.NoErrAlreadyUpToDate) {
			return nil, fmt.Errorf("reconcile: fetch: %w", fetchErr)
		}

		sourceHash, err = repo.ResolveRevision(c.SourceRevision)
		if err != nil {
			return nil, fmt.Errorf("reconcile: resolve source revision(%s): %w", c.SourceRevision, err)
		}

		var sourceCommit *object.Commit
		sourceCommit, err = repo.CommitObject(*sourceHash)
		if err != nil {
			return nil, fmt.Errorf("reconcile: resolve source revision(%s) commit: %w", c.SourceRevision, err)
		}

		if len(c.TargetRevisions) > 0 {
			resolved = true
		}

		for _, targetRev := range c.TargetRevisions {
			name := expandRevision(targetRev)

			var resolveRef plumbing.ReferenceName
			if name.IsTag() {
				resolveRef = name
			} else {
				resolveRef = plumbing.NewRemoteReferenceName(remoteName, name.Short())
			}

			var targetHash *plumbing.Hash
			targetHash, err = repo.ResolveRevision(plumbing.Revision(resolveRef.String()))
			if err != nil {
				resolved = false
				break
			}

			var targetCommit *object.Commit
			targetCommit, err = repo.CommitObject(*targetHash)
			if err != nil {
				resolved = false
				break
			}

			var bases []*object.Commit
			bases, err = sourceCommit.MergeBase(targetCommit)
			if err != nil || len(bases) == 0 {
				resolved = false
				break
			}

			mergeBases = append(mergeBases, bases...)
		}

		if resolved {
			break
		}

		fetchOpts.Depth *= depthIncrementalFactor
	}

	return mergeBases, nil
}

func expandRevision(rev plumbing.Revision) plumbing.ReferenceName {
	s := rev.String()

	if s == "HEAD" {
		return plumbing.HEAD
	}

	if strings.HasPrefix(s, "refs/heads/") || strings.HasPrefix(rev.String(), "refs/tags/") {
		return plumbing.ReferenceName(s)
	}

	if after, ok := strings.CutPrefix(s, "refs/remotes/"); ok {
		clean := after
		if _, after, ok = strings.Cut(clean, "/"); ok {
			return plumbing.ReferenceName(after)
		}

		return plumbing.NewBranchReferenceName(clean)
	}

	return plumbing.NewBranchReferenceName(s)
}
