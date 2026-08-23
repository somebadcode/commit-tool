/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package repobuilder

import (
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

type OperationFunc func(repo *git.Repository) error
type CloseFunc func() error

const DefaultBranchName = "main"

func Build(p string, ops ...OperationFunc) (*git.Repository, CloseFunc, error) {
	storage := osfs.New(p, osfs.WithBoundOS())
	repo, err := build(storage, DefaultBranchName)
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

	headRef := plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.NewBranchReferenceName(DefaultBranchName))
	if err = repo.Storer.SetReference(headRef); err != nil {
		_ = closeFunc()
		return nil, nil, fmt.Errorf("failed to initialize symbolic HEAD: %w", err)
	}

	err = CheckoutBranch(DefaultBranchName)(repo)
	if err != nil {
		return nil, nil, err
	}

	// Build repository.
	for _, op := range ops {
		if err = op(repo); err != nil {
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
	storage := filesystem.NewStorage(worktree, cache.NewObjectLRUDefault())

	repo, err := git.InitWithOptions(storage, nil, git.InitOptions{
		DefaultBranch: plumbing.NewBranchReferenceName(defaultBranch),
	})
	if err != nil {
		_ = storage.Close()

		return nil, fmt.Errorf("failed to initialise repository: %w", err)
	}

	return repo, nil
}

func checkoutBranch(repo *git.Repository, name string) error {
	target := plumbing.NewBranchReferenceName(name)

	if headRef, err := repo.Head(); err == nil {
		_ = repo.Storer.SetReference(plumbing.NewHashReference(target, headRef.Hash()))
	}

	symRef := plumbing.NewSymbolicReference(plumbing.HEAD, target)

	return repo.Storer.SetReference(symRef)
}

func CheckoutBranch(name string) OperationFunc {
	return func(repo *git.Repository) error {
		return checkoutBranch(repo, name)
	}
}

func Commit(message string, options git.CommitOptions) OperationFunc {
	return func(repo *git.Repository) error {
		if err := options.Validate(repo); err != nil {
			return fmt.Errorf("invalid options: %w", err)
		}

		if options.Committer.When.Equal(time.Time{}) {
			options.Committer.When = time.Now()
		}

		var treeHash plumbing.Hash
		var parents []plumbing.Hash

		headRef, err := repo.Head()
		if err == nil {
			parents = append(parents, headRef.Hash())

			var currentCommit *object.Commit

			currentCommit, err = repo.CommitObject(headRef.Hash())
			if err != nil {
				return fmt.Errorf("failed to get current commit: %w", err)
			}

			treeHash = currentCommit.TreeHash

		} else if errors.Is(err, plumbing.ErrReferenceNotFound) {
			emptyTree := new(object.Tree)

			treeObject := repo.Storer.NewEncodedObject()
			treeObject.SetType(plumbing.TreeObject)

			if err = emptyTree.Encode(treeObject); err != nil {
				return fmt.Errorf("failed to encode tree: %w", err)
			}

			treeHash, err = repo.Storer.SetEncodedObject(treeObject)
			if err != nil {
				return fmt.Errorf("failed to set encoded object: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get HEAD: %w", err)
		}

		commit := &object.Commit{
			TreeHash:     treeHash,
			ParentHashes: parents,
			Author:       *options.Author,
			Committer:    *options.Committer,
			Message:      message,
		}

		commitObj := repo.Storer.NewEncodedObject()
		commitObj.SetType(plumbing.CommitObject)
		if err = commit.Encode(commitObj); err != nil {
			return fmt.Errorf("failed to encode commit: %w", err)
		}

		var commitHash plumbing.Hash
		commitHash, err = repo.Storer.SetEncodedObject(commitObj)
		if err != nil {
			return fmt.Errorf("failed to set encoded commit object: %w", err)
		}

		var realHeadRef *plumbing.Reference
		realHeadRef, err = repo.Storer.Reference(plumbing.HEAD)
		if err != nil {
			return fmt.Errorf("failed to get HEAD reference: %w", err)
		}

		if realHeadRef.Type() == plumbing.SymbolicReference {
			target := realHeadRef.Target()

			branchRef := plumbing.NewHashReference(target, commitHash)
			if err = repo.Storer.SetReference(branchRef); err != nil {
				return fmt.Errorf("failed to set branch reference: %w", err)
			}

			return repo.Storer.SetReference(plumbing.NewHashReference(realHeadRef.Target(), commitHash))
		}

		return repo.Storer.SetReference(plumbing.NewHashReference(plumbing.HEAD, commitHash))
	}
}

func Tag(tag string, options *git.CreateTagOptions) OperationFunc {
	return func(repo *git.Repository) error {
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
