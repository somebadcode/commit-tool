package commithistorywalker

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type WalkFunc func(*object.Commit) error

// Walk will traverse the git repository, starting at the current head and stops on error.
func Walk(repo *git.Repository, walkFunc WalkFunc) error {
	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("could not get HEAD: %w", err)
	}

	var currentCommit *object.Commit
	currentCommit, err = repo.CommitObject(head.Hash())
	if err != nil {
		return fmt.Errorf("could not resolve commit at HEAD: %w", err)
	}

	seen := make(map[plumbing.Hash]bool)

	iter := object.NewCommitIterBSF(currentCommit, seen, nil)
	defer iter.Close()

	return iter.ForEach(func(commit *object.Commit) error {
		err = walkFunc(commit)
		if err != nil {
			return err
		}

		return nil
	})

}
