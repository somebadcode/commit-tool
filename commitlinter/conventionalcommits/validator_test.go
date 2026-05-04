/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package conventionalcommits

import (
	"testing"

	"github.com/go-git/go-git/v5/plumbing/object"

	"codeberg.org/somebadcode/commit-tool/commitparser"
)

func TestValidate(t *testing.T) {
	type fields struct {
		settings Settings
	}

	type args struct {
		msg    commitparser.CommitMessage
		commit *object.Commit
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			args: args{
				msg: commitparser.CommitMessage{
					Type:    "feat",
					Subject: "big change",
				},
			},
			fields: fields{
				settings: Settings{
					MaxHeaderLength: DefaultMaxHeaderLength,
				},
			},
		},
		{
			args: args{
				msg: commitparser.CommitMessage{
					Type:    "feat",
					Subject: "this message is too long and the validator should confirm that's the case",
				},
			},
			fields: fields{
				settings: Settings{
					MaxHeaderLength: DefaultMaxHeaderLength,
				},
			},
			wantErr: true,
		},
		{
			args: args{
				msg: commitparser.CommitMessage{
					Type:    "ASD",
					Subject: "big change",
				},
			},
			fields: fields{
				settings: Settings{
					MaxHeaderLength: DefaultMaxHeaderLength,
				},
			},
			wantErr: true,
		},
		{
			args: args{
				msg: commitparser.CommitMessage{
					Type: "feat",
				},
			},
			fields: fields{
				settings: Settings{
					MaxHeaderLength: DefaultMaxHeaderLength,
				},
			},
			wantErr: true,
		},
		{
			args: args{
				msg: commitparser.CommitMessage{
					Type:    "feat",
					Scope:   " foo bar baz ",
					Subject: "foo bar baz",
				},
			},
			fields: fields{
				settings: Settings{
					MaxHeaderLength: DefaultMaxHeaderLength,
				},
			},
			wantErr: true,
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New(tt.fields.settings)

			err := v.Validate(tt.args.msg, tt.args.commit)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateScope(t *testing.T) {
	type fields struct {
		settings Settings
	}

	type args struct {
		msg commitparser.CommitMessage
		in1 *object.Commit
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			args: args{
				msg: commitparser.CommitMessage{
					Scope: "something",
				},
			},
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Validator{
				settings: tt.fields.settings,
			}

			if err := v.ValidateScope(tt.args.msg, tt.args.in1); (err != nil) != tt.wantErr {
				t.Errorf("VerifyScope() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateSubject(t *testing.T) {
	type fields struct {
		settings Settings
	}

	type args struct {
		msg commitparser.CommitMessage
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				msg: commitparser.CommitMessage{
					Subject: "changing things",
				},
			},
		},
		{
			name: "empty",
			args: args{
				msg: commitparser.CommitMessage{},
			},
			wantErr: true,
		},
		{
			name: "punctuation",
			args: args{
				msg: commitparser.CommitMessage{
					Subject: "changing things.",
				},
			},
			wantErr: true,
		},
		{
			name: "leading_space",
			args: args{
				msg: commitparser.CommitMessage{
					Subject: " leading space",
				},
			},
			wantErr: true,
		},
		{
			name: "leading_upper",
			args: args{
				msg: commitparser.CommitMessage{
					Subject: "Feat",
				},
			},
			wantErr: true,
		},
		{
			name: "leading_invalid_code_point",
			args: args{
				msg: commitparser.CommitMessage{
					Subject: "\xffhello",
				},
			},
			wantErr: true,
		},
		{
			name: "trailing_invalid_code_point",
			args: args{
				msg: commitparser.CommitMessage{
					Subject: "hello\xff",
				},
			},
			wantErr: true,
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Validator{
				settings: tt.fields.settings,
			}

			if err := v.ValidateSubject(tt.args.msg, nil); (err != nil) != tt.wantErr {
				t.Errorf("VerifySubject() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateType(t *testing.T) {
	type fields struct {
		settings Settings
	}

	type args struct {
		msg commitparser.CommitMessage
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			args: args{
				msg: commitparser.CommitMessage{
					Type: "feat",
				},
			},
		},
		{
			args: args{
				msg: commitparser.CommitMessage{
					Type: "FEAT",
				},
			},
			wantErr: true,
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Validator{
				settings: tt.fields.settings,
			}

			if err := v.ValidateType(tt.args.msg, nil); (err != nil) != tt.wantErr {
				t.Errorf("VerifyType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
