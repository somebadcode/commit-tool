package revisionflag

import (
	"reflect"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
)

func TestNew(t *testing.T) {
	type args struct {
		rev plumbing.Revision
	}
	tests := []struct {
		name string
		args args
		want *Revision
	}{
		{
			name: "",
			args: args{
				rev: "main",
			},
			want: &Revision{
				Rev: "main",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.rev); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRevision_Set(t *testing.T) {
	type fields struct {
		rev plumbing.Revision
	}
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			fields: fields{
				rev: "refs/heads/main",
			},
			args: args{
				s: "HEAD",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Revision{
				Rev: tt.fields.rev,
			}
			if err := r.Set(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRevision_String(t *testing.T) {
	type fields struct {
		rev plumbing.Revision
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			fields: fields{
				rev: "HEAD",
			},
			want: "HEAD",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Revision{
				Rev: tt.fields.rev,
			}
			if got := r.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRevision_Type(t *testing.T) {
	type fields struct {
		rev plumbing.Revision
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			fields: fields{
				rev: "refs/heads/foo",
			},
			want: FlagType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Revision{
				Rev: tt.fields.rev,
			}
			if got := r.Type(); got != tt.want {
				t.Errorf("Type() = %v, want %v", got, tt.want)
			}
		})
	}
}
