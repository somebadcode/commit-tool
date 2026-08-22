/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package nextversion

import (
	"io"
	"log/slog"
	"reflect"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"codeberg.org/somebadcode/commit-tool/classifier"
	"codeberg.org/somebadcode/commit-tool/internal/repobuilder"
	"codeberg.org/somebadcode/commit-tool/semver"
)

func TestNextVersion_ComputeNextVersion(t *testing.T) {
	type fields struct {
		Logger           *slog.Logger
		Repository       *git.Repository
		Revision         plumbing.Revision
		Prerelease       []string
		Metadata         []string
		Writer           io.Writer
		CommitClassifier *classifier.Classifier
		VSuffix          bool
	}

	commitOpts := git.CommitOptions{
		AllowEmptyCommits: true,
		Author: &object.Signature{
			Name:  "Gopher",
			Email: "gopher@example.com",
		},
	}

	tests := []struct {
		name    string
		fields  fields
		want    *semver.Version
		wantErr bool
	}{
		{
			fields: fields{
				Repository: repobuilder.BuildTestHelper(t, t.TempDir(),
					repobuilder.Commit("feat: initial commit", commitOpts),
					repobuilder.Commit("fix: prevent panic", commitOpts),
				),
			},
			want: &semver.Version{
				Minor: 1,
			},
		},
		{
			fields: fields{
				Repository: repobuilder.BuildTestHelper(t, t.TempDir(),
					repobuilder.Commit("feat: initial commit", commitOpts),
					repobuilder.Commit("fix: prevent panic", commitOpts),
					repobuilder.Tag("14.1.4", &git.CreateTagOptions{
						Tagger:  commitOpts.Author,
						Message: "Release 14.1.4",
					}),
					repobuilder.Commit("fix: use equal fold", commitOpts),
				),
			},
			want: &semver.Version{
				Major: 14,
				Minor: 1,
				Patch: 5,
			},
		},
		{
			fields: fields{
				Repository: repobuilder.BuildTestHelper(t, t.TempDir(),
					repobuilder.Commit("feat: initial commit", commitOpts),
					repobuilder.Commit("fix: prevent panic", commitOpts),
					repobuilder.Tag("0.1.1", nil),
					repobuilder.Commit("fix: use equal fold", commitOpts),
				),
			},
			want: &semver.Version{
				Minor: 1,
				Patch: 2,
			},
		},
		{
			fields: fields{
				Repository: repobuilder.BuildTestHelper(t, t.TempDir(),
					repobuilder.Commit("feat: initial commit", commitOpts),
					repobuilder.Commit("fix: prevent panic", commitOpts),
					repobuilder.Tag("0.1.1", nil),
					repobuilder.Commit("fix: use equal fold", commitOpts),
				),
				Prerelease: []string{"rc", "0"},
			},
			want: &semver.Version{
				Minor: 1,
				Patch: 2,
				Prerelease: []semver.Segment{
					semver.NewSegment("rc"),
					semver.NewNumericSegment(0),
				},
			},
		},
		{
			fields: fields{
				Repository: repobuilder.BuildTestHelper(t, t.TempDir(),
					repobuilder.Commit("feat: initial commit", commitOpts),
					repobuilder.Commit("fix: prevent panic", commitOpts),
					repobuilder.Tag("0.1.1", nil),
					repobuilder.Commit("fix: use equal fold", commitOpts),
				),
				Prerelease: []string{"rc", "0"},
				Metadata:   []string{"dev"},
			},
			want: &semver.Version{
				Minor: 1,
				Patch: 2,
				Prerelease: []semver.Segment{
					semver.NewSegment("rc"),
					semver.NewNumericSegment(0),
				},
				BuildMetadata: []semver.Segment{
					semver.NewSegment("dev"),
				},
			},
		},
		{
			fields: fields{
				Repository: repobuilder.BuildTestHelper(t, t.TempDir(),
					repobuilder.Commit("feat: initial commit", commitOpts),
					repobuilder.Commit("fix: prevent panic", commitOpts),
					repobuilder.Tag("0.1.1", nil),
					repobuilder.Commit("fix: use equal fold", commitOpts),
					repobuilder.Commit("refactor!: make api spec compliant", commitOpts),
				),
			},
			want: &semver.Version{
				Minor: 2,
				Patch: 0,
			},
		},
		{
			fields: fields{
				Repository: repobuilder.BuildTestHelper(t, t.TempDir(),
					repobuilder.Commit("feat: initial commit", commitOpts),
					repobuilder.Commit("fix: prevent panic", commitOpts),
					repobuilder.Tag("0.1.1-rc.1", nil),
					repobuilder.Commit("fix: use equal fold", commitOpts),
					repobuilder.Commit("refactor!: make api spec compliant", commitOpts),
				),
			},
			want: &semver.Version{
				Minor: 1,
				Patch: 1,
			},
		},
		{
			fields: fields{
				Repository: repobuilder.BuildTestHelper(t, t.TempDir(),
					repobuilder.Commit("feat: initial commit", commitOpts),
					repobuilder.Commit("fix: prevent panic", commitOpts),
					repobuilder.Tag("1.0.0-rc.1", nil),
					repobuilder.Commit("fix: use equal fold", commitOpts),
					repobuilder.Commit("refactor!: make api spec compliant", commitOpts),
				),
			},
			want: &semver.Version{
				Major: 1,
			},
		},
		{
			fields: fields{
				Repository: repobuilder.BuildTestHelper(t, t.TempDir(),
					repobuilder.Commit("feat: initial commit", commitOpts),
					repobuilder.Commit("fix: prevent panic", commitOpts),
					repobuilder.Tag("1.3.0", nil),
					repobuilder.Tag("1.3.4", nil),
					repobuilder.Commit("fix: use equal fold", commitOpts),
					repobuilder.Commit("refactor!: make api spec compliant", commitOpts),
				),
			},
			want: &semver.Version{
				Major: 2,
			},
		},
		{
			fields: fields{
				Repository: repobuilder.BuildTestHelper(t, t.TempDir(),
					repobuilder.Commit("feat: initial commit", commitOpts),
					repobuilder.Commit("fix: prevent panic", commitOpts),
					repobuilder.Tag("1.3.0", nil),
					repobuilder.Tag("1.3.4", nil),
					repobuilder.Commit("fix: use equal fold", commitOpts),
				),
			},
			want: &semver.Version{
				Major: 1,
				Minor: 3,
				Patch: 5,
			},
		},
		{
			fields: fields{
				Repository: repobuilder.BuildTestHelper(t, t.TempDir(),
					repobuilder.Commit("feat: initial commit", commitOpts),
					repobuilder.Commit("fix: prevent panic", commitOpts),
					repobuilder.Tag("not-semver", nil),
					repobuilder.Tag("1.3.0", nil),
					repobuilder.Commit("fix: use equal fold", commitOpts),
				),
			},
			want: &semver.Version{
				Major: 1,
				Minor: 3,
				Patch: 1,
			},
		},
		{
			fields: fields{
				Repository: repobuilder.BuildTestHelper(t, t.TempDir(),
					repobuilder.Commit("feat: initial commit", commitOpts),
					repobuilder.Commit("fix: prevent panic", commitOpts),
					repobuilder.Tag("not-semver", &git.CreateTagOptions{
						Tagger:  commitOpts.Author,
						Message: "This doesn't matter!",
					}),
					repobuilder.Tag("1.3.4", nil),
					repobuilder.Commit("fix: use equal fold", commitOpts),
				),
			},
			want: &semver.Version{
				Major: 1,
				Minor: 3,
				Patch: 5,
			},
		},
		{
			fields: fields{
				Repository: repobuilder.BuildTestHelper(t, t.TempDir(),
					repobuilder.Commit("feat: initial commit", commitOpts),
					repobuilder.Commit("fix: prevent panic", commitOpts),
					repobuilder.Tag("1.0.1-beta.1", nil),
					repobuilder.Commit("refactor!: use equal fold", commitOpts),
				),
			},
			want: &semver.Version{
				Major: 2,
			},
		},
		{
			fields: fields{
				Repository: repobuilder.BuildTestHelper(t, t.TempDir(),
					repobuilder.Commit("feat: initial commit", commitOpts),
					repobuilder.Commit("fix: prevent panic", commitOpts),
					repobuilder.Tag("1.0.1-beta.1", nil),
					repobuilder.Commit("chore: bad comment grammar", commitOpts),
				),
				Prerelease: []string{"beta", "2"},
			},
			want: &semver.Version{
				Major: 1,
				Minor: 0,
				Patch: 1,
				Prerelease: []semver.Segment{
					semver.NewSegment("beta"),
					semver.NewNumericSegment(2),
				},
			},
		},
		{
			fields: fields{
				Repository: repobuilder.BuildTestHelper(t, t.TempDir(),
					repobuilder.Commit("feat: initial commit", commitOpts),
					repobuilder.Commit("fix: prevent panic", commitOpts),
					repobuilder.Tag("1.0.1", nil),
					repobuilder.Commit("chore: bad comment grammar", commitOpts),
				),
			},
			want: &semver.Version{
				Major: 1,
				Minor: 0,
				Patch: 1,
			},
		},
		{
			fields: fields{
				Repository: repobuilder.BuildTestHelper(t, t.TempDir(),
					repobuilder.Commit("feat: initial commit", commitOpts),
					repobuilder.Commit("fix: prevent panic", commitOpts),
					repobuilder.Commit("chore: bad comment grammar", commitOpts),
				),
				Prerelease: []string{"."},
			},
			wantErr: true,
		},
		{
			fields: fields{
				Repository: repobuilder.BuildTestHelper(t, t.TempDir(),
					repobuilder.Commit("feat: initial commit", commitOpts),
					repobuilder.Commit("fix: prevent panic", commitOpts),
					repobuilder.Commit("chore: bad comment grammar", commitOpts),
				),
				Metadata: []string{"."},
			},
			wantErr: true,
		},
		{
			fields: fields{
				Repository: repobuilder.BuildTestHelper(t, t.TempDir(),
					repobuilder.Commit("feat: initial commit", commitOpts),
					repobuilder.Commit("fix: prevent panic", commitOpts),
					repobuilder.Commit("chore: bad comment grammar", commitOpts),
				),
				Metadata: []string{"+"},
			},
			wantErr: true,
		},
		{
			fields: fields{
				Repository: repobuilder.BuildTestHelper(t, t.TempDir(),
					repobuilder.Commit("feat: initial commit", commitOpts),
					repobuilder.Commit("fix: prevent panic", commitOpts),
					repobuilder.Commit("chore: bad comment grammar", commitOpts),
				),
				Metadata: []string{"-"},
			},
			want: &semver.Version{
				Minor: 1,
				BuildMetadata: []semver.Segment{
					semver.NewSegment("-"),
				},
			},
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nv := &NextVersion{
				Logger:           tt.fields.Logger,
				Repository:       tt.fields.Repository,
				Revision:         tt.fields.Revision,
				Prerelease:       tt.fields.Prerelease,
				Metadata:         tt.fields.Metadata,
				Writer:           tt.fields.Writer,
				CommitClassifier: tt.fields.CommitClassifier,
				VSuffix:          tt.fields.VSuffix,
			}

			if err := nv.Validate(); err != nil {
				t.Fatalf("Validate() returned error: %v", err)
			}

			tags, err := tagsAsMap(t.Context(), nv.Repository)
			if err != nil {
				t.Fatalf("tagsAsMap() returned error: %v", err)
			}

			var got *semver.Version

			got, err = nv.ComputeNextVersion(t.Context(), tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("ComputeNextVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ComputeNextVersion() got = %v, want %v", got, tt.want)
			}
		})
	}
}
