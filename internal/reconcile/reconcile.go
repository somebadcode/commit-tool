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
)

const depthIncrementalFactor = 4
const maxDepth = 1000

var (
	ErrCannotReconcile     = errors.New("cannot reconcile revisions")
	ErrBadRevision         = errors.New("bad revision")
	ErrUnsupportedRevision = errors.New("unsupported revision")
)

func Reconcile(ctx context.Context, repo *git.Repository, c Config) ([]*object.Commit, error) {
	if err := c.Validate(repo); err != nil {
		return nil, fmt.Errorf("reconcile: validation: %w", err)
	}

	remote, err := repo.Remote(c.RemoteName)
	if err != nil {
		return nil, fmt.Errorf("reconcile: bad remote %q: %w", c.RemoteName, err)
	}

	var branches map[string]struct{}

	branches, err = revToBranch(ctx, repo, c.SourceRevision, remote)
	if err != nil {
		return nil, fmt.Errorf("reconcile: %w", err)
	}

	for _, targetRev := range c.TargetRevisions {
		branches[toBaseName(targetRev)] = struct{}{}
	}

	remoteName := remote.Config().Name
	var remoteURL string
	if len(remote.Config().URLs) > 0 {
		remoteURL = remote.Config().URLs[0]
	}

	refSpecs := make([]config.RefSpec, 0, len(c.TargetRevisions))

	for branch := range branches {
		refSpecs = append(refSpecs, config.RefSpec(fmt.Sprintf("+%s:%s",
			plumbing.NewBranchReferenceName(branch),
			plumbing.NewRemoteReferenceName(remoteName, branch),
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

		mergeBases = mergeBases[:0]

		if len(c.TargetRevisions) > 0 {
			resolved = true
		}

		for _, targetRev := range c.TargetRevisions {
			cleanTarget := toBaseName(targetRev)
			remoteRefName := plumbing.NewRemoteReferenceName(remoteName, cleanTarget)

			var targetHash *plumbing.Hash
			targetHash, err = repo.ResolveRevision(plumbing.Revision(remoteRefName.String()))
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
			if err != nil || len(mergeBases) == 0 {
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

func revToBranch(ctx context.Context, repo *git.Repository, rev plumbing.Revision, remote *git.Remote) (map[string]struct{}, error) {
	hash, err := repo.ResolveRevision(rev)
	if err != nil {
		return nil, fmt.Errorf("invalid revision: %w", err)
	}

	var refs []*plumbing.Reference
	refs, err = remote.ListContext(ctx, &git.ListOptions{
		PeelingOption: git.IgnorePeeled,
	})

	if err != nil {
		return nil, fmt.Errorf("list remote refs: %w", err)
	}

	var realRef *plumbing.Reference
	for _, ref := range refs {
		if ref.Hash() == *hash {
			realRef = ref
			break
		}
	}

	if realRef == nil {
		return nil, fmt.Errorf("revToBranch: revision %s not found: %w", rev.String(), ErrBadRevision)
	}

	refSet := make(map[plumbing.Hash]plumbing.ReferenceName, len(refs))
	for _, ref := range refs {
		if ref.Name().IsBranch() {
			refSet[ref.Hash()] = ref.Name()
		}
	}

	if realRef.Name().IsBranch() {
		shortName := realRef.Name().Short()
		return map[string]struct{}{
			shortName: {},
		}, nil
	}

	var headCommit *object.Commit
	headCommit, err = repo.CommitObject(*hash)
	if err != nil {
		return nil, fmt.Errorf("revToBranch: cannot read commit: %w", err)
	}

	branches := make(map[string]struct{})

	for _, parent := range headCommit.ParentHashes {
		if branchName, found := refSet[parent]; found {
			branches[branchName.Short()] = struct{}{}
		}
	}

	if len(branches) == 0 {
		return nil, fmt.Errorf("resolve: fictitious reference(%s) did not resolve to a known branch: %w", realRef.String(), ErrCannotReconcile)
	}

	return branches, nil
}

func toBaseName(rev plumbing.Revision) string {
	s := rev.String()

	if idx := strings.IndexAny(s, "~^:"); idx != -1 {
		s = s[:idx]
	}

	var name string

	for _, rule := range plumbing.RefRevParseRules[1:] {
		_, err := fmt.Sscanf(s, rule, &name)
		if err != nil {
			continue
		}

		return name
	}

	return rev.String()
}
