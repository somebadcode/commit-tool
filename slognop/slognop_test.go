package slognop_test

import (
	"context"
	"log/slog"
	"reflect"
	"testing"

	"github.com/somebadcode/commit-tool/slognop"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want slog.Handler
	}{
		{
			want: slognop.Handler{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := slognop.New().Handler(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandler_Enabled(t *testing.T) {
	type args struct {
		in0 context.Context
		in1 slog.Level
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			args: args{
				in0: context.TODO(),
				in1: slog.LevelError,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := slognop.Handler{}
			if got := n.Enabled(tt.args.in0, tt.args.in1); got != tt.want {
				t.Errorf("Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandler_Handle(t *testing.T) {
	type args struct {
		in0 context.Context
		in1 slog.Record
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args: args{
				in0: context.TODO(),
				in1: slog.Record{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := slognop.Handler{}
			if err := n.Handle(tt.args.in0, tt.args.in1); (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandler_WithAttrs(t *testing.T) {
	type args struct {
		in0 []slog.Attr
	}
	tests := []struct {
		name string
		args args
		want slog.Handler
	}{
		{
			args: args{
				in0: []slog.Attr{slog.String("key", "value")},
			},
			want: slognop.Handler{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := slognop.Handler{}
			if got := n.WithAttrs(tt.args.in0); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithAttrs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandler_WithGroup(t *testing.T) {
	type args struct {
		in0 string
	}
	tests := []struct {
		name string
		args args
		want slog.Handler
	}{
		{
			args: args{
				in0: "hello",
			},
			want: slognop.Handler{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := slognop.Handler{}
			if got := n.WithGroup(tt.args.in0); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}
