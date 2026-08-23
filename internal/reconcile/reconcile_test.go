/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package reconcile

import (
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"codeberg.org/somebadcode/commit-tool/internal/repobuilder"
)

func helperShallowClone(tb testing.TB, dest string, refName plumbing.ReferenceName, remoteURL string) *git.Repository {
	tb.Helper()

	repo, err := git.PlainCloneContext(tb.Context(), dest, true, &git.CloneOptions{
		URL:           remoteURL,
		RemoteName:    git.DefaultRemoteName,
		ReferenceName: refName,
		SingleBranch:  true,
		NoCheckout:    true,
		Depth:         1,
		Tags:          git.TagFollowing,
	})
	if err != nil {
		tb.Fatalf("helperShallowClone: %s", err)
	}

	var remoteRefName plumbing.ReferenceName
	if refName.IsBranch() {
		remoteRefName = plumbing.NewRemoteReferenceName(git.DefaultRemoteName, refName.Short())
	} else if refName.IsTag() {
		remoteRefName = refName
	} else {
		tb.Fatalf("helperShallowClone: reference name %q is not a branch or tag", refName.Short())
	}

	var remoteRef *plumbing.Reference
	remoteRef, err = repo.Reference(remoteRefName, true)
	if err != nil {
		tb.Fatalf("helperShallowClone: failed to find remote ref: %v", err)
	}

	detachedHEAD := plumbing.NewHashReference(plumbing.HEAD, remoteRef.Hash())
	err = repo.Storer.SetReference(detachedHEAD)
	if err != nil {
		tb.Fatalf("helperShallowClone: failed to set detached HEAD: %v", err)
	}

	return repo
}

func TestReconcile(t *testing.T) {
	type args struct {
		c Config
	}

	commitOpts := git.CommitOptions{
		AllowEmptyCommits: true,
		Author: &object.Signature{
			Name:  "Gopher",
			Email: "gopher@example.com",
			When:  time.Date(2026, 8, 22, 21, 07, 30, 0, time.UTC),
		},
	}

	tests := []struct {
		name    string
		repoOps []repobuilder.OperationFunc
		refName plumbing.ReferenceName
		args    args
		wantN   int
		wantErr bool
	}{
		{
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("Initial commit", commitOpts),
				repobuilder.Commit("feat: add foo", commitOpts),
				repobuilder.CheckoutBranch("fix-foo"),
				repobuilder.Commit("fix(foo): avoid panic", commitOpts),
				repobuilder.Commit("chore(foo): tweaked comments", commitOpts),
				repobuilder.Commit("chore(foo): fixed formatting", commitOpts),
			},
			args: args{
				c: Config{
					RemoteName:     git.DefaultRemoteName,
					SourceRevision: plumbing.Revision(plumbing.HEAD.String()),
					TargetRevisions: []plumbing.Revision{
						plumbing.Revision(plumbing.NewBranchReferenceName(repobuilder.DefaultBranchName).String()),
					},
				},
			},
			refName: plumbing.NewBranchReferenceName("fix-foo"),
			wantN:   1,
		},
		{
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("Initial commit", commitOpts),
				repobuilder.Commit("feat: add foo", commitOpts),
				repobuilder.CheckoutBranch("fix-foo"),
				repobuilder.Commit("fix(foo): avoid panic", commitOpts),
				repobuilder.Commit("chore(foo): tweaked comments", commitOpts),
				repobuilder.Commit("chore(foo): fixed formatting", commitOpts),
			},
			args: args{
				c: Config{
					RemoteName:     git.DefaultRemoteName,
					SourceRevision: plumbing.Revision(plumbing.NewBranchReferenceName("fix-foo").String()),
					TargetRevisions: []plumbing.Revision{
						plumbing.Revision(plumbing.NewBranchReferenceName(repobuilder.DefaultBranchName).String()),
					},
				},
			},
			refName: plumbing.NewBranchReferenceName("fix-foo"),
			wantN:   1,
		},
		{
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("Initial commit", commitOpts),
				repobuilder.Commit("feat: add foo", commitOpts),
				repobuilder.CheckoutBranch("fix-foo"),
				repobuilder.Commit("fix(foo): avoid panic", commitOpts),
				repobuilder.Commit("chore(foo): tweaked comments", commitOpts),
				repobuilder.Commit("chore(foo): fixed formatting", commitOpts),
				repobuilder.Tag("1.0.0", &git.CreateTagOptions{
					Tagger:  commitOpts.Author,
					Message: "Release 1.0.0",
					SignKey: nil,
				}),
			},
			args: args{
				c: Config{
					RemoteName:     git.DefaultRemoteName,
					SourceRevision: plumbing.Revision("refs/tags/1.0.0"),
					TargetRevisions: []plumbing.Revision{
						plumbing.Revision(plumbing.NewBranchReferenceName(repobuilder.DefaultBranchName).String()),
					},
				},
			},
			refName: plumbing.NewTagReferenceName("1.0.0"),
			wantN:   1,
		},
		{
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("Initial commit", commitOpts),
				repobuilder.Commit("feat: add foo", commitOpts),
				repobuilder.CheckoutBranch("fix-foo"),
				repobuilder.Commit("fix(foo): avoid panic", commitOpts),
				repobuilder.Commit("chore(foo): tweaked comments", commitOpts),
				repobuilder.Commit("chore(foo): fixed formatting", commitOpts),
				repobuilder.Tag("1.0.0", nil),
			},
			args: args{
				c: Config{
					RemoteName:     git.DefaultRemoteName,
					SourceRevision: plumbing.Revision("refs/tags/1.0.0"),
					TargetRevisions: []plumbing.Revision{
						plumbing.Revision(plumbing.NewBranchReferenceName(repobuilder.DefaultBranchName).String()),
					},
				},
			},
			refName: plumbing.NewTagReferenceName("1.0.0"),
			wantN:   1,
		},
		{
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("Initial commit", commitOpts),
				repobuilder.Commit("feat: add foo", commitOpts),
				repobuilder.Tag("1.0.0", nil),
				repobuilder.CheckoutBranch("fix-foo"),
				repobuilder.Commit("fix(foo): avoid panic", commitOpts),
				repobuilder.Commit("chore(foo): tweaked comments", commitOpts),
				repobuilder.Commit("chore(foo): fixed formatting", commitOpts),
				repobuilder.Tag("1.0.1", nil),
			},
			args: args{
				c: Config{
					RemoteName:     git.DefaultRemoteName,
					SourceRevision: plumbing.Revision("refs/tags/1.0.1"),
					TargetRevisions: []plumbing.Revision{
						plumbing.Revision("refs/tags/1.0.0"),
					},
				},
			},
			refName: plumbing.NewTagReferenceName("1.0.1"),
			wantN:   1,
		},
		{
			repoOps: []repobuilder.OperationFunc{
				repobuilder.Commit("Initial commit", commitOpts),
				repobuilder.Commit("feat: add foo", commitOpts),
				repobuilder.Tag("1.0.1", nil),
				repobuilder.CheckoutBranch("fix-foo"),
				repobuilder.Commit("fix(foo): avoid panic", commitOpts),
				repobuilder.Commit("chore(foo): tweaked comments", commitOpts),
				repobuilder.Commit("chore(foo): fixed formatting", commitOpts),
				repobuilder.Tag("1.0.2", nil),
			},
			args: args{
				c: Config{
					RemoteName:     git.DefaultRemoteName,
					SourceRevision: plumbing.Revision("refs/tags/1.0.2"),
					TargetRevisions: []plumbing.Revision{
						plumbing.Revision("refs/tags/1.0.0"),
					},
				},
			},
			refName: plumbing.NewTagReferenceName("1.0.1"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := t.TempDir()

			_, closeFunc, err := repobuilder.Build(source, tt.repoOps...)
			if err != nil {
				t.Fatalf("repobuilder.Build: %s", err)
			}
			if err = closeFunc(); err != nil {
				t.Fatalf("closeFunc(): %s", err)
			}

			cloneFS := t.TempDir()
			shallow := helperShallowClone(t, cloneFS, tt.refName, source)

			var got []*object.Commit
			got, err = Reconcile(t.Context(), shallow, tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("Reconcile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.wantN {
				t.Errorf("Reconcile() gotN = %v, wantN %v", len(got), tt.wantN)
			}
		})
	}
}
