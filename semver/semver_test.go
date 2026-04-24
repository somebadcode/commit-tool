/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package semver

import (
	"reflect"
	"testing"

	gcmp "github.com/google/go-cmp/cmp"
)

func TestNew(t *testing.T) {
	type args struct {
		major, minor, patch uint32
	}

	tests := []struct {
		name string
		args args
		want Version
	}{
		{
			args: args{
				major: 1,
				minor: 2,
				patch: 3,
			},
			want: Version{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.major, tt.args.minor, tt.args.patch)

			if !gcmp.Equal(got, tt.want) {
				t.Errorf("New() mismatch (-want +got):\n%s", gcmp.Diff(got, tt.want))
			}
		})
	}
}

func TestVersion_Clone(t *testing.T) {
	type fields struct {
		major, minor, patch uint32
		prerelease          []Segment
		buildMetadata       []Segment
		hasVPrefix          bool
	}

	tests := []struct {
		name   string
		fields fields
		want   Version
	}{
		{
			fields: fields{
				major:      0,
				minor:      4,
				patch:      19,
				hasVPrefix: true,
			},
			want: Version{
				Major:      0,
				Minor:      4,
				Patch:      19,
				HasVPrefix: true,
			},
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Version{
				Major:         tt.fields.major,
				Minor:         tt.fields.minor,
				Patch:         tt.fields.patch,
				HasVPrefix:    tt.fields.hasVPrefix,
				Prerelease:    tt.fields.prerelease,
				BuildMetadata: tt.fields.buildMetadata,
			}

			got := v.Clone()

			if !gcmp.Equal(got, tt.want) {
				t.Errorf("Clone() mismatch (-want +got):\n%s", gcmp.Diff(got, tt.want))
			}
		})
	}
}

func TestVersion_String(t *testing.T) {
	type fields struct {
		major, minor, patch uint32
		hasVPrefix          bool
		prerelease          []Segment
		buildMetadata       []Segment
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			fields: fields{
				major:      0,
				minor:      4,
				patch:      19,
				hasVPrefix: true,
			},
			want: "v0.4.19",
		},
		{
			fields: fields{
				major: 0,
				minor: 4,
				patch: 19,
				prerelease: []Segment{
					NewSegment("alpha"),
				},
				buildMetadata: []Segment{
					NewSegment("abc123"),
				},
				hasVPrefix: true,
			},
			want: "v0.4.19-alpha+abc123",
		},
		{
			fields: fields{
				major: 0,
				minor: 4,
				patch: 19,
				prerelease: []Segment{
					NewSegment("beta"),
					NewSegment("eu"),
				},
				buildMetadata: []Segment{NewSegment("abc444")},
				hasVPrefix:    true,
			},
			want: "v0.4.19-beta.eu+abc444",
		},
		{
			fields: fields{
				major: 0,
				minor: 4,
				patch: 20,
				buildMetadata: []Segment{
					NewSegment("abc451"),
				},
			},
			want: "0.4.20+abc451",
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Version{
				Major:         tt.fields.major,
				Minor:         tt.fields.minor,
				Patch:         tt.fields.patch,
				HasVPrefix:    tt.fields.hasVPrefix,
				Prerelease:    tt.fields.prerelease,
				BuildMetadata: tt.fields.buildMetadata,
			}

			got := v.String()

			if got != tt.want {
				t.Errorf("String() mismatch (-want +got):\n%s", gcmp.Diff(got, tt.want))
			}
		})
	}
}

func TestParse(t *testing.T) {
	type args struct {
		s string
	}

	tests := []struct {
		name    string
		args    args
		want    Version
		wantErr bool
	}{
		{
			args: args{s: "0.4.19"},
			want: Version{
				Major: 0,
				Minor: 4,
				Patch: 19,
			},
		},
		{
			args:    args{s: "banana"},
			wantErr: true,
		},
		{
			args:    args{s: "0.0"},
			wantErr: true,
		},
		{
			args:    args{s: "00"},
			wantErr: true,
		},
		{
			args:    args{s: "999"},
			wantErr: true,
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if Compare(got, tt.want) != 0 {
				t.Errorf("Parse() mismatch (-want +got):\n%s", gcmp.Diff(got, tt.want))
			}
		})
	}
}

func TestVersion_UnmarshalText(t *testing.T) {
	type fields struct {
		major, minor, patch uint32
		hasVPrefix          bool
		prerelease          []Segment
		buildMetadata       []Segment
	}

	type args struct {
		data []byte
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Version
		wantErr bool
	}{
		{
			args: args{
				data: []byte("0.4.19"),
			},
			want: Version{
				Major: 0,
				Minor: 4,
				Patch: 19,
			},
		},
		{
			args: args{
				data: []byte("v0.4.19-alpha+abc123"),
			},
			want: Version{
				Major: 0,
				Minor: 4,
				Patch: 19,
				Prerelease: []Segment{
					NewSegment("alpha"),
				},
				BuildMetadata: []Segment{
					NewSegment("abc123"),
				},
				HasVPrefix: true,
			},
		},
		{
			args: args{
				data: []byte("0.4.19+abc123"),
			},
			want: Version{
				Major: 0,
				Minor: 4,
				Patch: 19,
				BuildMetadata: []Segment{
					NewSegment("abc123"),
				},
			},
		},
		{
			args: args{
				data: []byte("0.4"),
			},
			wantErr: true,
		},
		{
			args: args{
				data: []byte("0.4.19-0"),
			},
			want: Version{
				Major: 0,
				Minor: 4,
				Patch: 19,
				Prerelease: []Segment{
					NewNumericSegment(0),
				},
			},
		},
		{
			args: args{
				data: []byte("0.4.19-01"),
			},
			want: Version{
				Major: 0,
				Minor: 4,
				Patch: 19,
				Prerelease: []Segment{
					NewSegment("01"),
				},
			},
		},
		{
			args: args{
				data: []byte("0.4.19-0alpha"),
			},
			want: Version{
				Major: 0,
				Minor: 4,
				Patch: 19,
				Prerelease: []Segment{
					NewSegment("0alpha"),
				},
			},
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Version{
				Major:         tt.fields.major,
				Minor:         tt.fields.minor,
				Patch:         tt.fields.patch,
				HasVPrefix:    tt.fields.hasVPrefix,
				Prerelease:    tt.fields.prerelease,
				BuildMetadata: tt.fields.buildMetadata,
			}

			err := v.UnmarshalText(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if got := v; Compare(got, tt.want) != 0 {
				t.Errorf("UnmarshalText() mismatch (-want +got):\n%s", gcmp.Diff(got, tt.want))
			}
		})
	}
}

func TestVersion_IsZero(t *testing.T) {
	type fields struct {
		major, minor, patch uint32
		hasVPrefix          bool
		prerelease          []Segment
		buildMetadata       []Segment
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			fields: fields{},
			want:   true,
		},
		{
			fields: fields{
				major: 1,
			},
		},
		{
			fields: fields{
				minor: 1,
			},
		},
		{
			fields: fields{
				patch: 1,
			},
		},
		{
			fields: fields{
				minor: 1,
				buildMetadata: []Segment{
					NewSegment("banana"),
				},
			},
		},
		{
			fields: fields{
				major: 1,
				prerelease: []Segment{
					NewSegment("apple"),
				},
			},
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Version{
				Major:         tt.fields.major,
				Minor:         tt.fields.minor,
				Patch:         tt.fields.patch,
				HasVPrefix:    tt.fields.hasVPrefix,
				Prerelease:    tt.fields.prerelease,
				BuildMetadata: tt.fields.buildMetadata,
			}

			if got := v.IsZero(); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersion_IncreaseMajor(t *testing.T) {
	type fields struct {
		major, minor, patch uint32
		hasVPrefix          bool
		prerelease          []Segment
		buildMetadata       []Segment
	}

	tests := []struct {
		name   string
		fields fields
		want   Version
	}{
		{
			fields: fields{},
			want: Version{
				Major: 1,
			},
		},
		{
			fields: fields{
				minor: 5,
			},
			want: Version{
				Major: 1,
			},
		},
		{
			fields: fields{
				minor: 5,
				patch: 18,
			},
			want: Version{
				Major: 1,
			},
		},
		{
			fields: fields{
				patch: 6,
			},
			want: Version{
				Major: 1,
			},
		},
		{
			fields: fields{
				major: 1,
				minor: 65,
				patch: 14,
			},
			want: Version{
				Major: 2,
			},
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Version{
				Major:         tt.fields.major,
				Minor:         tt.fields.minor,
				Patch:         tt.fields.patch,
				HasVPrefix:    tt.fields.hasVPrefix,
				Prerelease:    tt.fields.prerelease,
				BuildMetadata: tt.fields.buildMetadata,
			}

			v.IncreaseMajor()

			if got := v; Compare(got, tt.want) != 0 {
				t.Errorf("IncreaseMajor() = %v, want %v", &got, &tt.want)
			}
		})
	}
}

func TestVersion_IncreaseMinor(t *testing.T) {
	type fields struct {
		major, minor, patch uint32
		hasVPrefix          bool
		prerelease          []Segment
		buildMetadata       []Segment
	}

	tests := []struct {
		name   string
		fields fields
		want   Version
	}{
		{
			fields: fields{
				major: 1,
			},
			want: Version{
				Major: 1,
				Minor: 1,
			},
		},
		{
			fields: fields{
				major: 1,
				patch: 1,
			},
			want: Version{
				Major: 1,
				Minor: 1,
			},
		},
		{
			fields: fields{
				major: 1,
				patch: 1,
				prerelease: []Segment{
					NewSegment("banana"),
				},
			},
			want: Version{
				Major: 1,
				Minor: 1,
			},
		},
		{
			fields: fields{
				major: 1,
				patch: 1,
				prerelease: []Segment{
					NewSegment("banana"),
				},
				buildMetadata: []Segment{
					NewSegment("slim"),
				},
			},
			want: Version{
				Major: 1,
				Minor: 1,
			},
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Version{
				Major:         tt.fields.major,
				Minor:         tt.fields.minor,
				Patch:         tt.fields.patch,
				HasVPrefix:    tt.fields.hasVPrefix,
				Prerelease:    tt.fields.prerelease,
				BuildMetadata: tt.fields.buildMetadata,
			}

			v.IncreaseMinor()

			if got := v; Compare(got, tt.want) != 0 {
				t.Errorf("IncreaseMinor() = %v, want %v", &got, &tt.want)
			}
		})
	}
}

func TestVersion_IncreasePatch(t *testing.T) {
	type fields struct {
		major, minor, patch uint32
		hasVPrefix          bool
		prerelease          []Segment
		buildMetadata       []Segment
	}

	tests := []struct {
		name   string
		fields fields
		want   Version
	}{
		{
			fields: fields{
				major: 1,
				minor: 1,
			},
			want: Version{
				Major: 1,
				Minor: 1,
				Patch: 1,
			},
		},
		{
			fields: fields{
				major: 1,
				minor: 1,
				patch: 1,
				prerelease: []Segment{
					NewSegment("wash"),
				},
			},
			want: Version{
				Major: 1,
				Minor: 1,
				Patch: 2,
			},
		},
		{
			fields: fields{
				minor: 1,
				patch: 0,
				prerelease: []Segment{
					NewSegment("wash"),
				},
				buildMetadata: []Segment{
					NewSegment("fmt"),
				},
			},
			want: Version{
				Minor: 1,
				Patch: 1,
			},
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Version{
				Major:         tt.fields.major,
				Minor:         tt.fields.minor,
				Patch:         tt.fields.patch,
				HasVPrefix:    tt.fields.hasVPrefix,
				Prerelease:    tt.fields.prerelease,
				BuildMetadata: tt.fields.buildMetadata,
			}

			v.IncreasePatch()

			if got := v; Compare(got, tt.want) != 0 {
				t.Errorf("IncreasePatch() = %v, want %v", &got, &tt.want)
			}
		})
	}
}

func TestVersion_SetPrerelease(t *testing.T) {
	type fields struct {
		major, minor, patch uint32
		hasVPrefix          bool
		prerelease          []Segment
		buildMetadata       []Segment
	}

	type args struct {
		segments []string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Version
		wantErr bool
	}{
		{
			fields: fields{
				minor: 1,
			},
			args: args{
				segments: []string{
					"banana",
				},
			},
			want: Version{
				Minor: 1,
				Prerelease: []Segment{
					NewSegment("banana"),
				},
			},
		},
		{
			fields: fields{
				minor: 1,
				prerelease: []Segment{
					NewSegment("banana"),
				},
			},
			args: args{
				segments: []string{
					"hmmm",
				},
			},
			want: Version{
				Minor: 1,
				Prerelease: []Segment{
					NewSegment("hmmm"),
				},
			},
		},
		{
			fields: fields{
				minor: 1,
				prerelease: []Segment{
					NewSegment("banana"),
				},
				buildMetadata: []Segment{
					NewSegment("slim"),
				},
			},
			args: args{
				segments: []string{
					"hmmm",
				},
			},
			want: Version{
				Minor: 1,
				Prerelease: []Segment{
					NewSegment("hmmm"),
				},
				BuildMetadata: []Segment{
					NewSegment("slim"),
				},
			},
		},
		{
			fields: fields{
				minor: 1,
			},
			args: args{
				segments: []string{
					"bad+segment",
				},
			},
			wantErr: true,
		},
		{
			fields: fields{
				minor: 1,
			},
			args: args{
				segments: []string{
					"08",
				},
			},
			want: Version{
				Minor: 1,
				Prerelease: []Segment{
					NewSegment("08"),
				},
			},
		},
		{
			fields: fields{
				minor: 1,
			},
			args: args{
				segments: []string{
					"8",
					"gbeef",
				},
			},
			want: Version{
				Minor: 1,
				Prerelease: []Segment{
					NewNumericSegment(8),
					NewSegment("gbeef"),
				},
			},
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Version{
				Major:         tt.fields.major,
				Minor:         tt.fields.minor,
				Patch:         tt.fields.patch,
				HasVPrefix:    tt.fields.hasVPrefix,
				Prerelease:    tt.fields.prerelease,
				BuildMetadata: tt.fields.buildMetadata,
			}

			err := v.SetPrerelease(tt.args.segments...)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetPrerelease() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if Compare(v, tt.want) != 0 {
				t.Errorf("SetPrerelease() = %v, want %v", v, tt.want)
			}
		})
	}
}

func TestVersion_SetBuildMetadata(t *testing.T) {
	type fields struct {
		major, minor, patch uint32
		hasVPrefix          bool
		prerelease          []Segment
		buildMetadata       []Segment
	}

	type args struct {
		segments []string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Version
		wantErr bool
	}{
		{
			fields: fields{
				minor: 1,
			},
			args: args{
				segments: []string{
					"banana",
				},
			},
			want: Version{
				Minor: 1,
				BuildMetadata: []Segment{
					NewSegment("banana"),
				},
			},
		},
		{
			fields: fields{
				minor: 1,
				buildMetadata: []Segment{
					NewSegment("banana"),
				},
			},
			args: args{
				segments: []string{
					"hmmm",
				},
			},
			want: Version{
				Minor: 1,
				BuildMetadata: []Segment{
					NewSegment("hmmm"),
				},
			},
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Version{
				Major:         tt.fields.major,
				Minor:         tt.fields.minor,
				Patch:         tt.fields.patch,
				HasVPrefix:    tt.fields.hasVPrefix,
				Prerelease:    tt.fields.prerelease,
				BuildMetadata: tt.fields.buildMetadata,
			}

			err := v.SetBuildMetadata(tt.args.segments...)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetBuildMetadata() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if Compare(v, tt.want) != 0 {
				t.Errorf("SetBuildMetadata() = %v, want %v", v, tt.want)
			}
		})
	}
}

func TestMustParse(t *testing.T) {
	type args struct {
		s string
	}

	tests := []struct {
		name      string
		args      args
		want      Version
		wantPanic bool
	}{
		{
			args: args{
				s: "0.1.0",
			},
			want: Version{
				Minor: 1,
			},
		},
		{
			args: args{
				s: "bad",
			},
			wantPanic: true,
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				err := recover()
				if (err != nil) != tt.wantPanic {
					t.Errorf("MustParse() panic = %v, wantPanic %v", err, tt.wantPanic)
				}
				t.Logf("MustParse() panic = %v, wantPanic %v", err, tt.wantPanic)
			}()

			v := MustParse(tt.args.s)

			if Compare(v, tt.want) != 0 {
				t.Errorf("MustParse() = %v, want %v", v, tt.want)
			}
		})
	}
}

func TestVersion_MarshalText(t *testing.T) {
	type fields struct {
		Major         uint32
		Minor         uint32
		Patch         uint32
		HasVPrefix    bool
		Prerelease    []Segment
		BuildMetadata []Segment
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			fields: fields{
				Major: 1,
			},
			want: []byte("1.0.0"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Version{
				Major:         tt.fields.Major,
				Minor:         tt.fields.Minor,
				Patch:         tt.fields.Patch,
				HasVPrefix:    tt.fields.HasVPrefix,
				Prerelease:    tt.fields.Prerelease,
				BuildMetadata: tt.fields.BuildMetadata,
			}
			got, err := v.MarshalText()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalText() got = %v, want %v", got, tt.want)
			}
		})
	}
}
