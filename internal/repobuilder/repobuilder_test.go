/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package repobuilder_test

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"

	"codeberg.org/somebadcode/commit-tool/internal/repobuilder"
)

func ExampleBuild() {
	commitOpts := git.CommitOptions{
		AllowEmptyCommits: true,
		Author: &object.Signature{
			Name:  "Gopher",
			Email: "gopher@example.com",
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
		// Do not panic!
		panic(err)
	}

	var iter object.CommitIter
	iter, err = repo.Log(&git.LogOptions{
		Order: git.LogOrderBSF,
	})
	if err != nil {
		// Do not panic!
		panic(err)
	}
	defer iter.Close()

	err = iter.ForEach(func(commit *object.Commit) error {
		fmt.Println(commit.Message)
		return nil
	})
	if err != nil {
		// Do not panic!
		panic(err)
	}

	// Output:
	// feat: add bar
	// fix: foo is bar
	// test: something
	// initial commit
}
