package repobuilder_test

import (
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/somebadcode/commitlint/internal/repobuilder"
)

func ExampleBuild() {
	commitOpts := git.CommitOptions{
		AllowEmptyCommits: true,
		Author: &object.Signature{
			Name:  "Gopher",
			Email: "gopher@example.com",
			When:  time.Now().UTC(),
		},
	}

	repo, err := repobuilder.Build(
		repobuilder.Commit("initial commit", commitOpts),
		repobuilder.CheckoutBranch("test"),
		repobuilder.Commit("test: something", commitOpts),
		repobuilder.Commit("fix: foo is bar", commitOpts),
		repobuilder.Commit("feat: add bar", commitOpts),
	)
	if err != nil {
		// Handle this...
		panic(err)
	}

	var iter object.CommitIter
	iter, err = repo.Log(&git.LogOptions{
		Order: git.LogOrderBSF,
	})
	if err != nil {
		// Handle this...
		panic(err)
	}
	defer iter.Close()

	err = iter.ForEach(func(commit *object.Commit) error {
		fmt.Println(commit.Message)
		return nil
	})
	if err != nil {
		// Handle this...
		panic(err)
	}

	// Output:
	// feat: add bar
	// fix: foo is bar
	// test: something
	// initial commit
}
