/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package classifier

import (
	"io"
	"strings"
	"testing"

	"codeberg.org/somebadcode/commit-tool/commitparser"
)

func TestClassifier(t *testing.T) {
	type args struct {
		msg *commitparser.CommitMessage
	}
	tests := []struct {
		name     string
		mappings io.Reader
		args     args
		want     Classification
		wantErr  bool
	}{
		{
			name: "minor_feat",
			args: args{
				msg: &commitparser.CommitMessage{
					Type: "feat",
				},
			},
			want: Minor,
		},
		{
			name: "major_breaking_feat",
			args: args{
				msg: &commitparser.CommitMessage{
					Type:     "feat",
					Breaking: true,
				},
			},
			want: Major,
		},
		{
			name: "patch_fix",
			args: args{
				msg: &commitparser.CommitMessage{
					Type: "fix",
				},
			},
			want: Patch,
		},
		{
			name: "none_chore",
			args: args{
				msg: &commitparser.CommitMessage{
					Type: "chore",
				},
			},
			want: None,
		},
		{
			name:     "custom_patch",
			mappings: strings.NewReader(`{"patch": ["security"] }`),
			args: args{
				msg: &commitparser.CommitMessage{
					Type: "security",
				},
			},
			want: Patch,
		},
		{
			name:     "custom_non_unique",
			mappings: strings.NewReader(`{"patch": ["security"], "None": ["security"] }`),
			args: args{
				msg: &commitparser.CommitMessage{
					Type: "security",
				},
			},
			wantErr: true,
		},
		{
			name:     "custom_invalid",
			mappings: strings.NewReader(`{"patch": ["abc123"] }`),
			wantErr:  true,
		},
		{
			name:     "custom_invalid_json",
			mappings: strings.NewReader(`patch: abc123`),
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := New(tt.mappings)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			got := c.Classify(tt.args.msg)
			if got != tt.want {
				t.Errorf("Classify() got = %v, want %v", got, tt.want)
			}
		})
	}
}
