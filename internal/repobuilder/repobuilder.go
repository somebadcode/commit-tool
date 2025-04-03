/*
 * Copyright 2025 Tobias Dahlberg
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package repobuilder

import (
	"fmt"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

type OperationFunc func(repo *git.Repository, tree *git.Worktree) error

const DefaultBranchName = "main"

func Build(ops ...OperationFunc) (*git.Repository, error) {
	repo, err := build(memfs.New(), DefaultBranchName)
	if err != nil {
		return nil, err
	}

	var tree *git.Worktree
	tree, err = repo.Worktree()
	if err != nil {
		panic(err)
	}

	// Build repository.
	for _, op := range ops {
		if err = op(repo, tree); err != nil {
			return repo, err
		}
	}

	return repo, nil
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
