/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package repobuilder

import (
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

type OperationFunc func(repo *git.Repository, tree *git.Worktree) error
type CloseFunc func() error

const DefaultBranchName = "main"

func Build(p string, ops ...OperationFunc) (*git.Repository, CloseFunc, error) {
	repo, err := build(osfs.New(p, osfs.WithBoundOS()), DefaultBranchName)
	if err != nil {
		return nil, nil, err
	}

	closeFunc := func() error {
		// TODO: In `go-git/go-git/v6`, replace this with calling `Close() error` on the repository directly.
		if c, ok := repo.Storer.(io.Closer); ok {
			return c.Close()
		}

		return nil
	}

	defer func() {
		e := recover()
		if e != nil {
			_ = closeFunc()

			panic(e)
		}
	}()

	var tree *git.Worktree
	tree, err = repo.Worktree()
	if err != nil {
		panic(err)
	}

	// Build repository.
	for _, op := range ops {
		if err = op(repo, tree); err != nil {
			_ = closeFunc()

			return nil, nil, err
		}
	}

	return repo, closeFunc, nil
}

func BuildTestHelper(tb testing.TB, p string, ops ...OperationFunc) *git.Repository {
	tb.Helper()

	repo, closeFunc, err := Build(p, ops...)
	if err != nil {
		tb.Fatalf("failed to build repo: %v", err)
	}

	tb.Cleanup(func() {
		_ = closeFunc()
	})

	return repo
}

func build(worktree billy.Filesystem, defaultBranch string) (*git.Repository, error) {
	dot, err := worktree.Chroot(git.GitDirName)
	if err != nil {
		return nil, fmt.Errorf("failed to chroot dot directory: %w", err)
	}

	// Create object storage, using dot directory and LRU cache.
	storage := filesystem.NewStorage(dot, cache.NewObjectLRUDefault())

	var repo *git.Repository
	repo, err = git.InitWithOptions(storage, worktree, git.InitOptions{
		DefaultBranch: plumbing.NewBranchReferenceName(defaultBranch),
	})
	if err != nil {
		_ = storage.Close()

		return nil, fmt.Errorf("failed to initialise repository: %w", err)
	}

	return repo, nil
}

func checkoutBranch(repo *git.Repository, name string, worktree *git.Worktree) error {
	branch, _ := repo.Branch(name)

	opts := git.CheckoutOptions{
		// Create the branch if it doesn't exist.
		Create: branch == nil,
		Branch: plumbing.NewBranchReferenceName(name),
	}

	err := worktree.Checkout(&opts)
	if err != nil {
		return fmt.Errorf("failed to checkout branch %q: %w", opts.Branch, err)
	}

	return nil
}

func CheckoutBranch(name string) OperationFunc {
	return func(repo *git.Repository, worktree *git.Worktree) error {
		return checkoutBranch(repo, name, worktree)
	}
}

func Commit(message string, options git.CommitOptions) OperationFunc {
	return func(repo *git.Repository, worktree *git.Worktree) error {
		if err := options.Validate(repo); err != nil {
			return fmt.Errorf("invalid options: %w", err)
		}

		if options.Committer.When.Equal(time.Time{}) {
			options.Committer.When = time.Now()
		}

		// Parents can't be nil, it must be empty or tree of commits will not be formed correctly when start from nothing.
		options.Parents = plumbing.HashSlice{}

		_, err := worktree.Commit(message, &options)
		if err != nil {
			return fmt.Errorf("failed to commit %q: %w", message, err)
		}

		return nil
	}
}

func Tag(tag string, options *git.CreateTagOptions) OperationFunc {
	return func(repo *git.Repository, worktree *git.Worktree) error {
		head, err := repo.Head()
		if err != nil {
			return fmt.Errorf("failed to get HEAD: %w", err)
		}

		if options != nil {
			if err = options.Validate(repo, head.Hash()); err != nil {
				return fmt.Errorf("invalid options: %w", err)
			}
		}

		_, err = repo.CreateTag(tag, head.Hash(), options)
		if err != nil {
			return fmt.Errorf("failed to tag %q: %w", tag, err)
		}

		return nil
	}
}
