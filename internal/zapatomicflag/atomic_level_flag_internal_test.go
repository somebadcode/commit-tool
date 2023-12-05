package zapatomicflag

import (
	"reflect"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestAtomicLevelFlag_Set(t *testing.T) {
	type fields struct {
		level zap.AtomicLevel
	}
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    zapcore.Level
		wantErr bool
	}{
		{
			fields: fields{
				level: zap.NewAtomicLevelAt(zap.ErrorLevel),
			},
			args: args{
				s: "debug",
			},
			want:    zap.DebugLevel,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lvl := AtomicLevelFlag{
				level: tt.fields.level,
			}

			if err := lvl.Set(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got := lvl.level.Level(); got != tt.want {
				t.Errorf("Set() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAtomicLevelFlag_String(t *testing.T) {
	type fields struct {
		level zap.AtomicLevel
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			fields: fields{
				level: zap.NewAtomicLevelAt(zap.WarnLevel),
			},
			want: "warn",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lvl := AtomicLevelFlag{
				level: tt.fields.level,
			}
			if got := lvl.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAtomicLevelFlag_Type(t *testing.T) {
	type fields struct {
		level zap.AtomicLevel
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			fields: fields{
				level: zap.NewAtomicLevelAt(zap.ErrorLevel),
			},
			want: TypeLevel,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lvl := AtomicLevelFlag{
				level: tt.fields.level,
			}
			if got := lvl.Type(); got != tt.want {
				t.Errorf("Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		level zap.AtomicLevel
	}
	tests := []struct {
		name string
		args args
		want *AtomicLevelFlag
	}{
		{
			args: args{
				level: zap.NewAtomicLevel(),
			},
			want: &AtomicLevelFlag{
				level: zap.NewAtomicLevelAt(zap.InfoLevel),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.level); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
