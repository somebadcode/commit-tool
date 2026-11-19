/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package bindings

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/netip"
	"net/url"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	gcmp "github.com/google/go-cmp/cmp"
)

type dummyNode struct {
	t       testing.TB
	A       string
	B       int
	require bool
}

func newDummyNode(t *testing.T, require bool) *dummyNode {
	return &dummyNode{
		t:       t,
		require: require,
	}
}

func (d *dummyNode) Bind(b *Bindings) {
	d.t.Helper()

	if d.require {
		b.BindString("a", &d.A, Required())
		b.BindInt("b", &d.B, Required())
		return
	}

	b.BindString("a", &d.A)
	b.BindInt("b", &d.B)
}

func wantError[T error](t testing.TB, err error) {
	t.Helper()

	if _, ok := errors.AsType[T](err); !ok {
		t.Errorf("got error %T (%v), want %T", err, err, *new(T))
	}
}

type Fields struct {
	String         string
	Int            int
	Int8           int8
	Int16          int16
	Int32          int32
	Int64          int64
	Uint           uint
	Uint8          uint8
	Uint16         uint16
	Uint32         uint32
	Uint64         uint64
	Float32        float32
	Float64        float64
	Bool           bool
	Bool2          bool
	Time           time.Time
	Duration       time.Duration
	Addr           netip.Addr
	AddrPort       netip.AddrPort
	NetPrefix      netip.Prefix
	Path           string
	FileMode       fs.FileMode
	StringSlice    []string
	IntSlice       []int
	Int8Slice      []int8
	Int16Slice     []int16
	Int32Slice     []int32
	Int64Slice     []int64
	UintSlice      []uint
	Uint8Slice     []uint8
	Uint16Slice    []uint16
	Uint32Slice    []uint32
	Uint64Slice    []uint64
	Float32Slice   []float32
	Float64Slice   []float64
	BoolSlice      []bool
	TimeSlice      []time.Time
	DurationSlice  []time.Duration
	AddrSlice      []netip.Addr
	AddrPortSlice  []netip.AddrPort
	NetPrefixSlice []netip.Prefix
	PathSlice      []string
	FileModeSlice  []fs.FileMode
}

func TestBindings(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want Fields
	}{
		{
			name: "full",
			args: []string{
				"string", "string",
				"int", "-1",
				"int8", "-2",
				"int16", "-3",
				"int32", "-4",
				"int64", "-5",
				"uint", "6",
				"uint8", "7",
				"uint16", "8",
				"uint32", "9",
				"uint64", "10",
				"float32", "11.0",
				"float64", "12.0",
				"bool", // boolean
				"bool2",
				"time", "2006-01-02T15:04:05Z",
				"duration", "20ms",
				"addr", "192.0.2.3",
				"addrPort", "192.0.2.3:400",
				"netPrefix", "192.0.2.3/24",
				"path", "/var/lib/foo",
				"fileMode", "755",
				"strings", "foo",
				"strings", "bar",
				"ints", "-1",
				"ints", "-2",
				"ints", "-3",
				"int8s", "-4",
				"int8s", "-5",
				"int8s", "-6",
				"int16s", "-7",
				"int16s", "-8",
				"int16s", "-9",
				"int32s", "-10",
				"int32s", "-11",
				"int32s", "-12",
				"int64s", "-13",
				"int64s", "-14",
				"int64s", "-15",
				"uints", "16",
				"uints", "17",
				"uints", "18",
				"uint8s", "19",
				"uint8s", "20",
				"uint8s", "21",
				"uint16s", "22",
				"uint16s", "23",
				"uint16s", "24",
				"uint32s", "25",
				"uint32s", "26",
				"uint32s", "27",
				"uint64s", "28",
				"uint64s", "29",
				"uint64s", "30",
				"float32s", "31.1",
				"float32s", "32.2",
				"float32s", "33.3",
				"float64s", "34.1",
				"float64s", "35.2",
				"float64s", "36.3",
				"times", "2006-01-02T15:04:05Z",
				"times", "2007-02-03T16:05:06+01:00",
				"durations", "10ms",
				"durations", "20ms",
				"addrs", "192.0.2.3",
				"addrs", "192.0.2.4",
				"addrPorts", "192.0.2.3:400",
				"addrPorts", "192.0.2.4:400",
				"netPrefixes", "192.0.2.3/24",
				"netPrefixes", "192.0.2.4/24",
				"paths", "/var/lib/foo",
				"paths", "/var/lib/bar",
				"fileModes", "755",
				"fileModes", "600",
			},
			want: Fields{
				String:       "string",
				Int:          -1,
				Int8:         -2,
				Int16:        -3,
				Int32:        -4,
				Int64:        -5,
				Uint:         6,
				Uint8:        7,
				Uint16:       8,
				Uint32:       9,
				Uint64:       10,
				Float32:      11.0,
				Float64:      12.0,
				Bool:         true,
				Time:         time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
				Duration:     20 * time.Millisecond,
				Addr:         netip.MustParseAddr("192.0.2.3"),
				AddrPort:     netip.MustParseAddrPort("192.0.2.3:400"),
				NetPrefix:    netip.MustParsePrefix("192.0.2.3/24"),
				Path:         "/var/lib/foo",
				FileMode:     0o755,
				StringSlice:  []string{"foo", "bar"},
				IntSlice:     []int{-1, -2, -3},
				Int8Slice:    []int8{-4, -5, -6},
				Int16Slice:   []int16{-7, -8, -9},
				Int32Slice:   []int32{-10, -11, -12},
				Int64Slice:   []int64{-13, -14, -15},
				UintSlice:    []uint{16, 17, 18},
				Uint8Slice:   []uint8{19, 20, 21},
				Uint16Slice:  []uint16{22, 23, 24},
				Uint32Slice:  []uint32{25, 26, 27},
				Uint64Slice:  []uint64{28, 29, 30},
				Float32Slice: []float32{31.1, 32.2, 33.3},
				Float64Slice: []float64{34.1, 35.2, 36.3},
				TimeSlice: []time.Time{
					time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
					time.Date(2007, 2, 3, 16, 5, 6, 0, time.FixedZone("CET", 1*60*60)),
				},
				DurationSlice: []time.Duration{
					10 * time.Millisecond,
					20 * time.Millisecond,
				},
				AddrSlice: []netip.Addr{
					netip.MustParseAddr("192.0.2.3"),
					netip.MustParseAddr("192.0.2.4"),
				},
				AddrPortSlice: []netip.AddrPort{
					netip.MustParseAddrPort("192.0.2.3:400"),
					netip.MustParseAddrPort("192.0.2.4:400"),
				},
				NetPrefixSlice: []netip.Prefix{
					netip.MustParsePrefix("192.0.2.3/24"),
					netip.MustParsePrefix("192.0.2.4/24"),
				},
				PathSlice: []string{
					"/var/lib/foo",
					"/var/lib/bar",
				},
				FileModeSlice: []fs.FileMode{
					0o755,
					0o600,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Fields

			bindings := New()

			bindings.BindString("string", &got.String, Required())
			bindings.BindInt("int", &got.Int, Required())
			bindings.BindInt("int8", &got.Int8, Required())
			bindings.BindInt("int16", &got.Int16, Required())
			bindings.BindInt("int32", &got.Int32, Required())
			bindings.BindInt("int64", &got.Int64, Required())
			bindings.BindUint("uint", &got.Uint, Required())
			bindings.BindUint("uint8", &got.Uint8, Required())
			bindings.BindUint("uint16", &got.Uint16, Required())
			bindings.BindUint("uint32", &got.Uint32, Required())
			bindings.BindUint("uint64", &got.Uint64, Required())
			bindings.BindFloat("float32", &got.Float32, Required())
			bindings.BindFloat("float64", &got.Float64, Required())
			bindings.BindBool("bool", &got.Bool, Required())
			bindings.BindBoolNeg("bool2", &got.Bool2, Required())
			bindings.BindTime("time", &got.Time, time.RFC3339, Required())
			bindings.BindDuration("duration", &got.Duration, Required())
			bindings.BindAddr("addr", &got.Addr, Required())
			bindings.BindAddrPort("addrPort", &got.AddrPort, Required())
			bindings.BindNetPrefix("netPrefix", &got.NetPrefix, Required())
			bindings.BindPath("path", &got.Path, Required())
			bindings.BindFileMode("fileMode", &got.FileMode, Required())
			bindings.BindStringSlice("strings", &got.StringSlice)
			bindings.BindIntSlice("ints", &got.IntSlice)
			bindings.BindIntSlice("int8s", &got.Int8Slice)
			bindings.BindIntSlice("int16s", &got.Int16Slice)
			bindings.BindIntSlice("int32s", &got.Int32Slice)
			bindings.BindIntSlice("int64s", &got.Int64Slice)
			bindings.BindUintSlice("uints", &got.UintSlice)
			bindings.BindUintSlice("uint8s", &got.Uint8Slice)
			bindings.BindUintSlice("uint16s", &got.Uint16Slice)
			bindings.BindUintSlice("uint32s", &got.Uint32Slice)
			bindings.BindUintSlice("uint64s", &got.Uint64Slice)
			bindings.BindFloatSlice("float32s", &got.Float32Slice)
			bindings.BindFloatSlice("float64s", &got.Float64Slice)
			bindings.BindBool("bool", &got.Bool)
			bindings.BindTimeSlice("times", time.RFC3339, &got.TimeSlice)
			bindings.BindDurationSlice("durations", &got.DurationSlice)
			bindings.BindAddrSlice("addrs", &got.AddrSlice)
			bindings.BindAddrPortSlice("addrPorts", &got.AddrPortSlice)
			bindings.BindNetPrefixSlice("netPrefixes", &got.NetPrefixSlice)
			bindings.BindPathSlice("paths", &got.PathSlice)
			bindings.BindFileModeSlice("fileModes", &got.FileModeSlice)

			_, err := bindings.Parse(tt.args, nil)
			if err != nil {
				t.Errorf("bindings.Parse() error = %v", err)
				return
			}

			var stringerOpt = gcmp.FilterValues(func(x, y any) bool {
				_, ok1 := x.(fmt.Stringer)
				_, ok2 := y.(fmt.Stringer)
				return ok1 && ok2
			}, gcmp.Transformer("String", func(v any) string {
				return v.(fmt.Stringer).String()
			}))

			diff := gcmp.Diff(got, tt.want, stringerOpt)
			if diff != "" {
				t.Errorf("bindings.BindStringSlice() returned diff (-want +got):\n%s", diff)
			}
		})

	}
}

func TestBindings_BindAddr(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    netip.Addr
		wantErr bool
	}{
		{
			name: "valid_ipv4",
			args: []string{"addr", "1.2.3.4"},
			want: netip.MustParseAddr("1.2.3.4"),
		},
		{
			name:    "invalid_ipv4",
			args:    []string{"addr", "256.2.3.4"},
			wantErr: true,
		},
		{
			name: "valid_ipv6ll_with_scope",
			args: []string{"addr", "fe80::1%en0"},
			want: netip.MustParseAddr("fe80::1%en0"),
		},
		{
			name: "valid_ipv6_ula",
			args: []string{"addr", "fdc0:bb2b:1eb:1:3c16:58f7:de56:a657"},
			want: netip.MustParseAddr("fdc0:bb2b:1eb:1:3c16:58f7:de56:a657"),
		},
		{
			name: "valid_ipv6_ula_with_scope",
			args: []string{"addr", "fdc0:bb2b:1eb:1:3c16:58f7:de56:a657%enp5s0"},
			want: netip.MustParseAddr("fdc0:bb2b:1eb:1:3c16:58f7:de56:a657%enp5s0"),
		},
		{
			name: "valid_ipv6_loopback",
			args: []string{"addr", "::1"},
			want: netip.MustParseAddr("::1"),
		},
		{
			name: "valid_ipv6_loopback_with_scope",
			args: []string{"addr", "::1%lo"},
			want: netip.MustParseAddr("::1%lo"),
		},
		{
			name: "missing_required",
			args: []string{},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
		{
			name:    "missing_value",
			args:    []string{"addr"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got netip.Addr

			b := New()

			b.BindAddr("addr", &got, tt.options...)

			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindAddr() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			if got.String() != tt.want.String() {
				t.Errorf("Bindings.BindAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_BindAddrPort(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    netip.AddrPort
		wantErr bool
	}{
		{
			name: "valid_ipv4",
			args: []string{"bind", "1.2.3.4:50000"},
			want: netip.MustParseAddrPort("1.2.3.4:50000"),
		},
		{
			name:    "valid_ipv4_missing_port",
			args:    []string{"bind", "1.2.3.4"},
			wantErr: true,
		},
		{
			name:    "valid_ipv4_invalid_port",
			args:    []string{"bind", "1.2.3.4:65539"},
			wantErr: true,
		},
		{
			name: "valid_ipv6_all",
			args: []string{"bind", "[::]:50000"},
			want: netip.MustParseAddrPort("[::]:50000"),
		},
		{
			name: "valid_ipv6_loopback",
			args: []string{"bind", "[::1]:50000"},
			want: netip.MustParseAddrPort("[::1]:50000"),
		},
		{
			name:    "missing_value",
			args:    []string{"bind"},
			wantErr: true,
		},
		{
			name: "missing_required",
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got netip.AddrPort

			b := New()

			b.BindAddrPort("bind", &got, tt.options...)

			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindAddrPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			if got.String() != tt.want.String() {
				t.Errorf("Bindings.BindAddrPort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_BindNetPrefix(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    netip.Prefix
		wantErr bool
	}{
		{
			name: "valid_ipv4",
			args: []string{"cidr", "127.0.0.0/8"},
			want: netip.MustParsePrefix("127.0.0.0/8"),
		},
		{
			name: "valid_ipv4_full",
			args: []string{"cidr", "192.168.0.128/32"},
			want: netip.MustParsePrefix("192.168.0.128/32"),
		},
		{
			name: "valid_ipv6",
			args: []string{"cidr", "fe80::/10"},
			want: netip.MustParsePrefix("fe80::/10"),
		},
		{
			name:    "missing_value",
			args:    []string{"cidr"},
			wantErr: true,
		},
		{
			name: "missing_required",
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got netip.Prefix

			b := New()

			b.BindNetPrefix("cidr", &got, tt.options...)

			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindNetPrefix() error = %v, wantErr %v", err, tt.wantErr)
			} else if tt.wantErr {
				return
			}

			if got.String() != tt.want.String() {
				t.Errorf("Bindings.BindNetPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_BindBool(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    bool
		wantErr bool
	}{
		{
			name: "valid",
			args: []string{"up"},
			want: true,
		},
		{
			name: "valid_required",
			args: []string{"up"},
			options: []Option{
				Required(),
			},
			want: true,
		},
		{
			name: "missing_required",
			args: []string{},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool

			b := New()

			b.BindBool("up", &got, tt.options...)

			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			if got != tt.want {
				t.Errorf("Bindings.BindBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_BindBoolNeg(t *testing.T) {
	type fields struct {
		up bool
	}

	tests := []struct {
		name    string
		args    []string
		options []Option
		fields  fields
		want    fields
		wantErr bool
	}{
		{
			name: "valid",
			args: []string{"down"},
			fields: fields{
				up: true,
			},
			want: fields{
				up: false,
			},
		},
		{
			name: "required",
			args: []string{"down"},
			options: []Option{
				Required(),
			},
			fields: fields{
				up: true,
			},
			want: fields{
				up: false,
			},
		},
		{
			name: "missing_required",
			args: []string{},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got fields

			b := New()

			b.BindBoolNeg("down", &got.up, tt.options...)

			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			if got.up != tt.want.up {
				t.Errorf("Bindings.BindBool() = %v, want %v", got.up, tt.want.up)
			}
		})
	}
}

func TestBindings_BindDuration(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    time.Duration
		wantErr bool
	}{
		{
			name: "valid",
			args: []string{"for", "1h"},
			want: 1 * time.Hour,
		},
		{
			name: "valid_nano",
			args: []string{"for", "771ns"},
			want: 771 * time.Nanosecond,
		},
		{
			name:    "invalid",
			args:    []string{"for", "asd"},
			wantErr: true,
		},
		{
			name: "required",
			args: []string{"for", "997us"},
			options: []Option{
				Required(),
			},
			want: 997 * time.Microsecond,
		},
		{
			name: "missing_required",
			args: []string{},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
		{
			name:    "missing_value",
			args:    []string{"for"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got time.Duration

			b := New()

			b.BindDuration("for", &got, tt.options...)

			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			if got.String() != tt.want.String() {
				t.Errorf("Bindings.BindDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_BindFileMode(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    os.FileMode
		wantErr bool
	}{
		{
			name: "valid",
			args: []string{"mode", "755"},
			want: 0o755,
		},
		{
			name:    "invalid",
			args:    []string{"mode", "asd"},
			wantErr: true,
		},
		{
			name: "required",
			args: []string{"mode", "755"},
			options: []Option{
				Required(),
			},
			want: 0o755,
		},
		{
			name: "missing_required",
			args: []string{},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
		{
			name:    "missing_value",
			args:    []string{"mode"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got os.FileMode

			b := New()

			b.BindFileMode("mode", &got, tt.options...)

			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindFileMode() error = %v, wantErr %v", err, tt.wantErr)
				return

			} else if tt.wantErr {
				return
			}

			if got != tt.want {
				t.Errorf("Bindings.BindFileMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_BindFloat(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    float64
		wantErr bool
	}{
		{
			name: "valid",
			args: []string{"float", "99.09"},
			want: 99.09,
		},
		{
			name:    "invalid",
			args:    []string{"float", "asd"},
			wantErr: true,
		},
		{
			name: "required",
			args: []string{"float", "123.9"},
			options: []Option{
				Required(),
			},
			want: 123.9,
		},
		{
			name: "missing_required",
			args: []string{},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
		{
			name:    "missing_value",
			args:    []string{"float"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got float64

			b := New()

			b.BindFloat("float", &got, tt.options...)

			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindFloat() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			if got != tt.want {
				t.Errorf("Bindings.BindFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_BindInt(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    int
		wantErr bool
	}{
		{
			name: "valid",
			args: []string{"int", "99"},
			want: 99,
		},
		{
			name:    "invalid",
			args:    []string{"int", "asd"},
			wantErr: true,
		},
		{
			name: "required",
			args: []string{"int", "-269"},
			options: []Option{
				Required(),
			},
			want: -269,
		},
		{
			name: "missing_required",
			args: []string{},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
		{
			name:    "missing_value",
			args:    []string{"int"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got int

			b := New()

			b.BindInt("int", &got, tt.options...)

			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			if got != tt.want {
				t.Errorf("Bindings.BindInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_BindUint(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    uint
		wantErr bool
	}{
		{
			name: "valid",
			args: []string{"uint", "99"},
			want: 99,
		},
		{
			name:    "invalid",
			args:    []string{"uint", "asd"},
			wantErr: true,
		},
		{
			name: "required",
			args: []string{"uint", "5000"},
			options: []Option{
				Required(),
			},
			want: 5000,
		},
		{
			name: "missing_required",
			args: []string{},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
		{
			name: "missing_value",
			args: []string{"uint"},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got uint

			b := New()

			b.BindUint("uint", &got, tt.options...)

			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindUint() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			if got != tt.want {
				t.Errorf("Bindings.BindUint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_BindPath(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    string
		wantErr bool
	}{
		{
			name: "valid",
			args: []string{"path", "/var/lib/foo"},
			want: "/var/lib/foo",
		},
		{
			name:    "invalid",
			args:    []string{"path", ""},
			wantErr: true,
		},
		{
			name: "required",
			args: []string{"path", "/var/lib/bar"},
			options: []Option{
				Required(),
			},
			want: "/var/lib/bar",
		},
		{
			name: "missing_required",
			args: []string{},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
		{
			name:    "missing_value",
			args:    []string{"path"},
			wantErr: true,
		},
		{
			name: "clean",
			args: []string{"path", "/var//lib/../lib/foo"},
			want: "/var/lib/foo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string

			b := New()

			b.BindPath("path", &got, tt.options...)

			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			if got != tt.want {
				t.Errorf("Bindings.BindPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_BindString(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    string
		wantErr bool
	}{
		{
			name: "valid",
			args: []string{"set", "hello"},
			want: "hello",
		},
		{
			name: "required",
			args: []string{"set", "hello"},
			options: []Option{
				Required(),
			},
			want: "hello",
		},
		{
			name: "missing_required",
			args: []string{},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
		{
			name:    "missing_value",
			args:    []string{"set"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string

			b := New()

			b.BindString("set", &got, tt.options...)

			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindString() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			if got != tt.want {
				t.Errorf("Bindings.BindString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_BindTime(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    time.Time
		wantErr bool
	}{
		{
			name: "valid",
			args: []string{"at", "2022-01-01T00:00:00Z"},
			want: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "invalid",
			args:    []string{"at", "2022-01-01T00:00:00Z0900"},
			wantErr: true,
		},
		{
			name: "required",
			args: []string{"at", "2022-01-01T00:00:00Z"},
			options: []Option{
				Required(),
			},
			want: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "missing_required",
			args: []string{},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
		{
			name:    "missing_value",
			args:    []string{"at"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got time.Time

			b := New()

			b.BindTime("at", &got, time.RFC3339, tt.options...)

			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			if got.Compare(tt.want) != 0 {
				t.Errorf("Bindings.BindTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_BindAddrSlice(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    []netip.Addr
		wantErr bool
	}{
		{
			name: "valid",
			args: []string{"addr", "192.168.0.1", "addr", "192.168.0.2"},
			want: []netip.Addr{
				netip.MustParseAddr("192.168.0.1"),
				netip.MustParseAddr("192.168.0.2"),
			},
		},
		{
			name:    "invalid",
			args:    []string{"addr", "192.168.0.666", "addr", "192.168.0.2"},
			wantErr: true,
		},
		{
			name: "required",
			args: []string{"addr", "192.168.0.1", "addr", "192.168.0.2"},
			options: []Option{
				Required(),
			},
			want: []netip.Addr{
				netip.MustParseAddr("192.168.0.1"),
				netip.MustParseAddr("192.168.0.2"),
			},
		},
		{
			name: "missing_required",
			args: []string{""},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
		{
			name:    "missing_value",
			args:    []string{"addr"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []netip.Addr

			b := New()

			b.BindAddrSlice("addr", &got, tt.options...)

			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindAddrSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			if slices.CompareFunc(got, tt.want, func(a netip.Addr, b netip.Addr) int {
				return strings.Compare(a.String(), b.String())
			}) != 0 {
				t.Errorf("Bindings.BindAddrSlice() = %v, want %v", got, tt.want)
			}

		})
	}
}

func TestBindings_BindAddrPortSlice(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    []netip.AddrPort
		wantErr bool
	}{
		{
			name: "valid",
			args: []string{"addr", "192.168.0.1:8443", "addr", "192.168.0.2:8443"},
			want: []netip.AddrPort{
				netip.MustParseAddrPort("192.168.0.1:8443"),
				netip.MustParseAddrPort("192.168.0.2:8443"),
			},
		},
		{
			name:    "invalid",
			args:    []string{"addr", "192.168.0.1:99999", "addr", "192.168.0.2:8443"},
			wantErr: true,
		},
		{
			name: "required",
			args: []string{"addr", "192.168.0.1:443", "addr", "192.168.0.2:80"},
			options: []Option{
				Required(),
			},
			want: []netip.AddrPort{
				netip.MustParseAddrPort("192.168.0.1:443"),
				netip.MustParseAddrPort("192.168.0.2:80"),
			},
		},
		{
			name: "missing_required",
			args: []string{""},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
		{
			name:    "missing_value",
			args:    []string{"addr"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []netip.AddrPort

			b := New()

			b.BindAddrPortSlice("addr", &got, tt.options...)

			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindAddrPortSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			if slices.CompareFunc(got, tt.want, func(a netip.AddrPort, b netip.AddrPort) int {
				return strings.Compare(a.String(), b.String())
			}) != 0 {
				t.Errorf("Bindings.BindAddrPortSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_BindNetPrefixSlice(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    []netip.Prefix
		wantErr bool
	}{
		{
			name: "valid_ipv4",
			args: []string{"cidr", "192.168.0.0/24", "cidr", "192.168.0.2/32"},
			want: []netip.Prefix{
				netip.MustParsePrefix("192.168.0.0/24"),
				netip.MustParsePrefix("192.168.0.2/32"),
			},
		},
		{
			name: "valid_ipv6",
			args: []string{"cidr", "fe80::/10", "cidr", "fd00::/16"},
			want: []netip.Prefix{
				netip.MustParsePrefix("fe80::/10"),
				netip.MustParsePrefix("fd00::/16"),
			},
		},
		{
			name:    "invalid",
			args:    []string{"cidr", "192.168.0.0/64", "cidr", "192.168.0.2/32"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []netip.Prefix

			b := New()

			b.BindNetPrefixSlice("cidr", &got, tt.options...)
			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindNetPrefixSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			if slices.CompareFunc(got, tt.want, func(a netip.Prefix, b netip.Prefix) int {
				return strings.Compare(a.String(), b.String())
			}) != 0 {
				t.Errorf("Bindings.BindNetPrefixSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_BindDurationSlice(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    []time.Duration
		wantErr bool
	}{
		{
			name: "valid",
			args: []string{"for", "1h", "for", "2m"},
			want: []time.Duration{
				1 * time.Hour,
				2 * time.Minute,
			},
		},
		{
			name:    "invalid",
			args:    []string{"for", "1q", "for", "2m"},
			wantErr: true,
		},
		{
			name: "required",
			args: []string{"for", "1m", "for", "2s"},
			options: []Option{
				Required(),
			},
			want: []time.Duration{
				1 * time.Minute,
				2 * time.Second,
			},
		},
		{
			name: "missing_required",
			args: []string{},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
		{
			name: "missing_value",
			args: []string{"for"},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []time.Duration

			b := New()

			b.BindDurationSlice("for", &got, tt.options...)
			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindDurationSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			if slices.CompareFunc(got, tt.want, func(a time.Duration, b time.Duration) int {
				return strings.Compare(a.String(), b.String())
			}) != 0 {
				t.Errorf("Bindings.BindDurationSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_BindTimeSlice(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    []time.Time
		wantErr bool
	}{
		{
			name: "valid",
			args: []string{"at", "2006-01-02T15:04:05Z", "at", "2026-11-20T18:54:14+01:00"},
			want: []time.Time{
				time.Date(2006, 01, 02, 15, 4, 5, 0, time.UTC),
				time.Date(2026, 11, 20, 18, 54, 14, 0, time.FixedZone("CET", 1*60*60)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []time.Time

			b := New()

			b.BindTimeSlice("at", time.RFC3339, &got, tt.options...)

			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindTimeSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			if slices.CompareFunc(got, tt.want, func(a time.Time, b time.Time) int {
				return strings.Compare(a.String(), b.String())
			}) != 0 {
				t.Errorf("Bindings.BindTimeSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_BindStringSlice(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    []string
		wantErr bool
	}{
		{
			name: "valid",
			args: []string{"text", "hello", "text", "world"},
			want: []string{"hello", "world"},
		},
		{
			name: "required",
			args: []string{"text", "hello", "text", "world"},
			options: []Option{
				Required(),
			},
			want: []string{"hello", "world"},
		},
		{
			name: "missing_required",
			args: []string{},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
		{
			name: "missing_value",
			args: []string{"text"},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string

			b := New()

			b.BindStringSlice("text", &got, tt.options...)

			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindStringSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			if slices.Compare(got, tt.want) != 0 {
				t.Errorf("Bindings.BindStringSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_BindFileModeSlice(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    []os.FileMode
		wantErr bool
	}{
		{
			name: "valid",
			args: []string{"mode", "777", "mode", "755"},
			want: []os.FileMode{
				os.FileMode(0o777),
				os.FileMode(0o755),
			},
		},
		{
			name: "required",
			args: []string{"mode", "777", "mode", "755"},
			options: []Option{
				Required(),
			},
			want: []os.FileMode{
				os.FileMode(0o777),
				os.FileMode(0o755),
			},
		},
		{
			name: "missing_required",
			args: []string{},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
		{
			name:    "missing_value",
			args:    []string{"mode"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []os.FileMode

			b := New()

			b.BindFileModeSlice("mode", &got, tt.options...)

			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindFileModeSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			if slices.Compare(got, tt.want) != 0 {
				t.Errorf("Bindings.BindFileModeSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_BindPathSlice(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    []string
		wantErr bool
	}{
		{
			name: "valid",
			args: []string{"path", "/var/lib/foo", "path", "/var/lib/bar"},
			want: []string{"/var/lib/foo", "/var/lib/bar"},
		},
		{
			name: "clean",
			args: []string{"path", "/var/lib/../lib/foo", "path", "/var/lib/bar/foo/../"},
			want: []string{"/var/lib/foo", "/var/lib/bar"},
		},
		{
			name: "required",
			args: []string{},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
		{
			name: "missing_required",
			args: []string{"path"},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string

			b := New()

			b.BindPathSlice("path", &got, tt.options...)

			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindPathSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			if slices.Compare(got, tt.want) != 0 {
				t.Errorf("Bindings.BindPathSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_BindFloatSlice(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    []float64
		wantErr bool
	}{
		{
			name: "valid",
			args: []string{"float", "3.14159", "float", "1.0"},
			want: []float64{3.14159, 1.0},
		},
		{
			name: "required",
			args: []string{"float", "3.14159", "float", "1.0"},
			want: []float64{3.14159, 1.0},
		},
		{
			name: "missing_required",
			args: []string{},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
		{
			name:    "missing_value",
			args:    []string{"float"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []float64

			b := New()

			b.BindFloatSlice("float", &got, tt.options...)
			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindFloatSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			if slices.Compare(got, tt.want) != 0 {
				t.Errorf("Bindings.BindFloatSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_BindIntSlice(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    []int
		wantErr bool
	}{
		{
			name: "valid",
			args: []string{"int", "-999", "int", "256"},
			want: []int{-999, 256},
		},
		{
			name:    "invalid",
			args:    []string{"int", "fff", "int", "256"},
			wantErr: true,
		},
		{
			name: "required",
			args: []string{"int", "-999", "int", "32768"},
			want: []int{-999, 32768},
		},
		{
			name: "missing_required",
			args: []string{},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
		{
			name:    "missing_value",
			args:    []string{"int"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []int

			b := New()

			b.BindIntSlice("int", &got, tt.options...)

			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindIntSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			if slices.Compare(got, tt.want) != 0 {
				t.Errorf("Bindings.BindIntSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_BindUintSlice(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    []uint
		wantErr bool
	}{
		{
			name: "valid",
			args: []string{"uint", "999", "uint", "256"},
			want: []uint{999, 256},
		},
		{
			name:    "invalid",
			args:    []string{"uint", "-999", "uint", "256"},
			wantErr: true,
		},
		{
			name: "required",
			args: []string{"uint", "999", "uint", "256"},
			options: []Option{
				Required(),
			},
			want: []uint{999, 256},
		},
		{
			name: "missing_required",
			args: []string{},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
		{
			name:    "missing_value",
			args:    []string{"uint"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []uint

			b := New()

			b.BindUintSlice("uint", &got, tt.options...)

			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindUintSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			if slices.Compare(got, tt.want) != 0 {
				t.Errorf("Bindings.BindUintSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_BindLogLevel(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    slog.Level
		wantErr bool
	}{
		{
			name: "debug",
			args: []string{"log", "debug"},
			want: slog.LevelDebug,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got slog.Level

			b := New()

			b.BindLogLevel("log", &got, tt.options...)

			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindLogLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			if got != tt.want {
				t.Errorf("Bindings.BindLogLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_BindLogLevelVar(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    slog.Level
		wantErr bool
	}{
		{
			name: "debug",
			args: []string{"log", "debug"},
			want: slog.LevelDebug,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got slog.LevelVar

			b := New()

			b.BindLogLevelVar("log", &got, tt.options...)

			_, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.BindLogLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			if got.Level() != tt.want {
				t.Errorf("Bindings.BindLogLevel() = %v, want %v", got.String(), tt.want)
			}
		})
	}
}

func TestBindings_BindURL(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    string
		wantErr bool
	}{
		{
			name: "valid",
			args: []string{"url", "https://example.com/"},
			want: "https://example.com/",
		},
		{
			name:    "invalid",
			args:    []string{"url", "://broken"},
			wantErr: true,
		},
		{
			name: "required",
			args: []string{"url", "https://example.com/"},
			options: []Option{
				Required(),
			},
			want: "https://example.com/",
		},
		{
			name: "missing_required",
			args: []string{},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
		{
			name:    "missing_value",
			args:    []string{"url"},
			wantErr: true,
		},
	}

	t.Run("value", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var got url.URL

				b := New()

				b.BindURL("url", &got, tt.options...)

				_, err := b.Parse(tt.args, nil)
				if (err != nil) != tt.wantErr {
					t.Fatalf("Bindings.BindURL() error = %v, wantErr %v", err, tt.wantErr)
				}

				if !tt.wantErr && got.String() != tt.want {
					t.Errorf("Bindings.BindURL() = %v, want %v", got.String(), tt.want)
				}
			})
		}
	})

	t.Run("reference", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var got *url.URL

				b := New()

				b.BindURL("url", &got, tt.options...)

				_, err := b.Parse(tt.args, nil)
				if (err != nil) != tt.wantErr {
					t.Fatalf("Bindings.BindURL() error = %v, wantErr %v", err, tt.wantErr)
				}

				if !tt.wantErr && got == nil {
					t.Fatalf("Bindings.BindURL() did not return a URL")
				}

				if !tt.wantErr && got.String() != tt.want {
					t.Errorf("Bindings.BindURL() = %v, want %v", got.String(), tt.want)
				}
			})
		}
	})
}

func TestBindings_BindURLSlice(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		options []Option
		want    []string
		wantErr bool
	}{
		{
			name: "valid",
			args: []string{"url", "https://example.com/", "url", "https://example.org"},
			want: []string{"https://example.com/", "https://example.org"},
		},
		{
			name:    "invalid",
			args:    []string{"url", "://broken", "url", "https://example.org"},
			wantErr: true,
		},
	}

	t.Run("value", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var got []url.URL

				b := New()

				b.BindURLSlice("url", &got)

				_, err := b.Parse(tt.args, nil)
				if (err != nil) != tt.wantErr {
					t.Fatalf("Bindings.BindURLSlice() error = %v, wantErr %v", err, tt.wantErr)
				}

				gotSlice := make([]string, 0, len(got))
				for _, u := range got {
					gotSlice = append(gotSlice, u.String())
				}

				if !tt.wantErr && slices.Compare(gotSlice, tt.want) != 0 {
					t.Errorf("Bindings.BindURLSlice() = %v, want %v", gotSlice, tt.want)
				}
			})
		}
	})

	t.Run("reference", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var got []*url.URL

				b := New()

				b.BindURLSlice("url", &got)

				_, err := b.Parse(tt.args, nil)
				if (err != nil) != tt.wantErr {
					t.Fatalf("Bindings.BindURLSlice() error = %v, wantErr %v", err, tt.wantErr)
				}

				if !tt.wantErr && got == nil {
					t.Fatalf("Bindings.BindURLSlice() did not return URLs")
				}

				gotSlice := make([]string, 0, len(got))
				for _, u := range got {
					gotSlice = append(gotSlice, u.String())
				}

				if !tt.wantErr && slices.Compare(gotSlice, tt.want) != 0 {
					t.Errorf("Bindings.BindURLSlice() = %v, want %v", gotSlice, tt.want)
				}
			})
		}
	})
}

func TestBindings_BindNode(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		require  bool
		bindings func(*Bindings)
		options  []Option
		want     dummyNode
		wantRem  []string
		wantErr  bool
	}{
		{
			name: "valid",
			args: []string{"dummy", "a", "abc", "b", "123"},
			want: dummyNode{
				A: "abc",
				B: 123,
			},
		},
		{
			name: "missing_required",
			args: []string{"set", "hello, world"},
			bindings: func(b *Bindings) {
				var s string
				b.BindString("set", &s)
			},
			options: []Option{
				Required(),
			},
			wantErr: true,
		},
		{
			name:    "no_sub_match",
			args:    []string{"dummy", "asd"},
			wantRem: []string{"asd"},
		},
		{
			name:    "invalid",
			args:    []string{"dummy"},
			require: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New()

			got := newDummyNode(t, tt.require)

			b.BindNode("dummy", got, tt.options...)

			if tt.bindings != nil {
				tt.bindings(b)
			}

			gotRem, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				return
			}

			got.t = nil

			if tt.wantRem != nil && slices.Compare(gotRem, tt.wantRem) != 0 {
				t.Errorf("Bindings.BindNode() rem = %v, want %v", gotRem, tt.wantRem)
			}

			if !reflect.DeepEqual(got, &tt.want) {
				t.Errorf("Bindings.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_Parse(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		setup        func(*Bindings)
		options      []Option
		want         []string
		wantErr      bool
		wantErrExact func(testing.TB, error)
	}{
		{
			name: "unique_prefix_match",
			setup: func(b *Bindings) {
				var dev, dest string
				b.BindString("device", &dev)
				b.BindString("destination", &dest)
			},
			args: []string{"dev", "en0"},
		},
		{
			name: "ambiguous_prefix_match",
			setup: func(b *Bindings) {
				var dev, dest string
				b.BindString("device", &dev)
				b.BindString("destination", &dest)
			},
			args:         []string{"de", "en0"},
			wantErr:      true,
			wantErrExact: wantError[*AmbiguousKeywordError],
		},
		{
			name: "missing_key",
			setup: func(b *Bindings) {
				var dev, dest string
				b.BindString("device", &dev)
				b.BindString("destination", &dest, Required())
			},
			args:         []string{"dev", "en0"},
			wantErr:      true,
			wantErrExact: wantError[*MissingKeywordError],
		},
		{
			name: "missing_value",
			setup: func(b *Bindings) {
				var dev, dest string
				b.BindString("device", &dev)
				b.BindString("destination", &dest)
			},
			args:         []string{"dev", "en0", "dest"},
			wantErr:      true,
			wantErrExact: wantError[*MissingValueError],
		},
		{
			name: "no_match",
			setup: func(b *Bindings) {
				var dev, dest string
				b.BindString("device", &dev)
				b.BindString("destination", &dest)
			},
			args: []string{"other", "en0"},
			want: []string{"other", "en0"},
		},
		{
			name: "missing_required_no_match",
			setup: func(b *Bindings) {
				var dev, dest string
				b.BindString("device", &dev)
				b.BindString("destination", &dest, Required())
			},
			args:         []string{"other", "en0"},
			want:         []string{"other", "en0"},
			wantErr:      true,
			wantErrExact: wantError[*UnknownKeywordError],
		},
		{
			name: "trailing_args",
			setup: func(b *Bindings) {
				var dev, dest string
				b.BindString("device", &dev)
				b.BindString("destination", &dest)
			},
			args: []string{"dev", "en0", "--", "hello", "world"},
			want: []string{"hello", "world"},
		},
		{
			name: "repeating_non_repeatdestle",
			setup: func(b *Bindings) {
				var dev, dest string
				b.BindString("device", &dev)
				b.BindString("destination", &dest)
			},
			args: []string{"dev", "en0", "dev", "en1"},
			want: []string{"dev", "en1"},
		},
		{
			name: "no_arg_error",
			setup: func(b *Bindings) {
				var v bool
				b.BindValue("fail", &v, func(_ string) (bool, error) {
					return false, errors.New("boom")
				}, NoArgument())
			},
			args:    []string{"fail"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New()

			tt.setup(b)

			got, err := b.Parse(tt.args, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bindings.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if tt.wantErr {
				if tt.wantErrExact != nil {
					tt.wantErrExact(t, err)
				}
				return
			}

			if slices.Compare(got, tt.want) != 0 {
				t.Errorf("Bindings.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBindings_All(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*Bindings)
		stopAt string
		want   []Info
	}{
		{
			setup: func(b *Bindings) {
				var dev, dest string
				b.BindString("device", &dev, WithDescription("network device name"))
				b.BindString("destination", &dest, WithDescription("network destination name"))
			},
			want: []Info{
				{
					Description: "network destination name",
				},
				{
					Description: "network device name",
				},
			},
		},
		{
			setup: func(b *Bindings) {
				var dev, dest string
				b.BindString("device", &dev, WithDescription("network device name"))
				b.BindString("destination", &dest, WithDescription("network destination name"))
			},
			stopAt: "device",
			want: []Info{
				{
					Description: "network destination name",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New()

			tt.setup(b)

			got := make([]Info, 0)

			for keyword, info := range b.All() {
				if keyword == tt.stopAt {
					break
				}

				got = append(got, Info{
					Description: info.Description,
					Required:    info.Required,
					Repeatable:  info.Repeatable,
					NoArg:       info.NoArg,
				})
			}

			if !reflect.DeepEqual(got, tt.want) {
				diff := gcmp.Diff(got, tt.want)
				t.Errorf("Bindings.All() diff (-want +got):\n%s", diff)
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
			err:  &UnknownKeywordError{Keyword: "device"},
			want: "device: unknown keyword",
		},
		{
			err:  &AmbiguousKeywordError{Keyword: "de", Matches: []string{"destination", "device"}},
			want: "de: ambiguous keyword matches(destination, device)",
		},
		{
			err:  &MissingKeywordError{Keyword: "device"},
			want: "device: missing keyword",
		},
		{
			err:  &MissingValueError{Keyword: "device"},
			want: "device: missing value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("UnknownKeywordError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
