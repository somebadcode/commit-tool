/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package cli

import (
	"errors"
	"strings"
	"testing"

	gcmp "github.com/google/go-cmp/cmp"

	"codeberg.org/somebadcode/commit-tool/cli/bindings"
)

func TestCLI_Execute(t *testing.T) {
	type node struct {
		name   string
		binder bindings.Binder
	}

	tests := []struct {
		name       string
		args       []string
		nodes      []node
		options    []Option
		wantStdout string
		wantStderr string
		wantErr    bool
	}{
		{
			name: "stdout_stderr_args",
			nodes: []node{
				{
					name:   "dummy",
					binder: NewDummyNode(),
				},
			},
			args: []string{"dummy", "lines", "hello,", "extra", "foo"},
			options: []Option{
				WithStdin(strings.NewReader(" world")),
			},
			wantStdout: "hello, world",
			wantStderr: "foo",
		},
		{
			name: "all_discard",
			nodes: []node{
				{
					name:   "dummy",
					binder: NewDummyNode(),
				},
			},
			args: []string{"dummy", "lines", "hello,", "extra", "foo"},
			options: []Option{
				WithStdin(nil),
				WithStdout(nil),
				WithStderr(nil),
			},
		},
		{
			name: "execute_multiple",
			nodes: []node{
				{
					name:   "dummy",
					binder: NewDummyNode(WithPriority(100)),
				},
				{
					name:   "another-dummy",
					binder: NewDummyNode(WithPriority(50)),
				},
			},
			args: []string{"dummy", "lines", "hello,", "extra", "foo", "another-dummy"},
			options: []Option{
				WithStdin(nil),
				WithStdout(nil),
				WithStderr(nil),
			},
		},
		{
			name: "fail_execute",
			nodes: []node{
				{
					name:   "dummy",
					binder: NewDummyNode(FailExecute()),
				},
			},
			args:    []string{"dummy", "lines", "hello,", "extra", "foo"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr strings.Builder

			options := []Option{
				WithStdin(nil),
				WithStdout(&stdout),
				WithStderr(&stderr),
			}

			app := New(append(options, tt.options...)...)

			for _, v := range tt.nodes {
				err := app.Register(v.name, v.binder)
				if err != nil {
					t.Fatalf("Register() error = %v", err)
				}
			}

			err := app.Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			err = app.Execute(t.Context())
			if (err != nil) != tt.wantErr {
				t.Fatalf("Execute() error = %v, wantErr = %v", err, tt.wantErr)
			} else if tt.wantErr {
				return
			}

			if gotStdout := stdout.String(); gotStdout != tt.wantStdout {
				diff := gcmp.Diff(tt.wantStdout, gotStdout)
				t.Errorf("Stdout mismatch (-want +got):\n%s", diff)
			}

			if gotStderr := stderr.String(); gotStderr != tt.wantStderr {
				diff := gcmp.Diff(tt.wantStderr, gotStderr)
				t.Errorf("Stderr mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCLI_Register(t *testing.T) {
	type node struct {
		name   string
		binder bindings.Binder
	}

	tests := []struct {
		name    string
		args    []string
		nodes   []node
		wantErr bool
	}{
		{
			name: "success",
			nodes: []node{
				{
					name:   "dummy",
					binder: NewDummyNode(),
				},
			},
		},
		{
			name: "double_register",
			nodes: []node{
				{
					name:   "dummy",
					binder: NewDummyNode(),
				},
				{
					name:   "dummy",
					binder: NewDummyNode(),
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New()

			var errs error

			for _, v := range tt.nodes {
				err := app.Register(v.name, v.binder)
				errs = errors.Join(errs, err)
			}

			if (errs != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

func TestCLI_Parse(t *testing.T) {
	type node struct {
		name   string
		binder bindings.Binder
	}

	tests := []struct {
		name    string
		args    []string
		nodes   []node
		wantErr bool
	}{
		{
			name: "success",
			nodes: []node{
				{
					name:   "dummy",
					binder: NewDummyNode(),
				},
			},
			args: []string{"dummy", "lines", "hello,", "extra", "foo"},
		},
		{
			name: "bad_keyword",
			nodes: []node{
				{
					name:   "dummy",
					binder: NewDummyNode(),
				},
			},
			args:    []string{"dummy", "coffee", "black"},
			wantErr: true,
		},
		{
			name: "missing_value",
			nodes: []node{
				{
					name:   "dummy",
					binder: NewDummyNode(),
				},
			},
			args:    []string{"dummy", "lines"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New()

			for _, v := range tt.nodes {
				err := app.Register(v.name, v.binder)
				if err != nil {
					t.Fatalf("Register() error = %v", err)
				}
			}

			err := app.Parse(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			err:  &UnknownCommandError{Name: "thing"},
			want: "thing: unknown command",
		},
		{
			err:  &AlreadyExistsError{Name: "thing"},
			want: "thing: command already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()

			if got != tt.want {
				diff := gcmp.Diff(tt.want, got)

				t.Errorf("%T.Error() mismatch (-want +got):\n%s", tt.err, diff)
			}
		})
	}
}
