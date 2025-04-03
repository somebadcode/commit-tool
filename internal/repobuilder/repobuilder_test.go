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

package repobuilder_test

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/somebadcode/commit-tool/internal/repobuilder"
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
